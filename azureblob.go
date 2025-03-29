package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

type BlobClient interface {
	GetRange(string, string, int64, int64, []byte) (int64, error)
}

type azureBlobClient struct {
	client *azblob.Client
}

// buffer will be filled before returning a read. The assumption
// is you will mostly read sequentially. In the future perhaps
// this buffer could be informed by the fileheader so multiple files
// could be accessed sequentially without dumping the buffer.
// or, just make a new one per file.
type BlobReader struct {
	client    BlobClient
	mu        *sync.Mutex
	container string
	blobname  string
	blobsize  int64
	buffer    []byte
	bufStart  *int64
	bufEnd    *int64
}

func NewBlobClient(account string, key string) (BlobClient, error) {
	if account == "" || key == "" {
		return &azureBlobClient{}, fmt.Errorf("account and key may not be empty")
	}
	cred, err := azblob.NewSharedKeyCredential(account, key)
	if err != nil {
		return &azureBlobClient{}, err
	}

	c, err := azblob.NewClientWithSharedKeyCredential(fmt.Sprintf("https://%s.blob.core.windows.net/", account), cred, nil)
	if err != nil {
		return &azureBlobClient{}, nil
	}

	return &azureBlobClient{c}, err
}

func (c BlobReader) ReadAt(p []byte, off int64) (int, error) {
	log.Printf("requesting from %v to %v", off, off+int64(len(p)))
	c.mu.Lock()
	defer c.mu.Unlock()
	if off+int64(len(p)) > c.blobsize {
		return 0, fmt.Errorf("read is past EOF")
	}

	// requested range is in the buffer
	if off >= *c.bufStart && off+int64(len(p)) < *c.bufEnd {
		log.Printf("range is in buffer, buffer is %v to %v", *c.bufStart, *c.bufEnd)
		copy(p, c.buffer[int(off-*c.bufStart):int(off+int64(len(p))-*c.bufStart)])
		return len(p), nil
	}

	// requested range is not fully in the buffer. fill and return.
	if off > *c.bufEnd || off+int64(len(p)) < *c.bufStart {
		bytesToRead := min(int64(len(c.buffer)), c.blobsize-off)
		log.Printf("filling buffer off: %v, bytes: %v\n", off, bytesToRead)
		log.Printf("A buffer len %v cap %v", len(c.buffer), cap(c.buffer))
		n, err := c.client.GetRange(c.container, c.blobname, off, bytesToRead, c.buffer)
		log.Printf("B buffer len %v cap %v", len(c.buffer), cap(c.buffer))
		log.Printf("B p len %v cap %v", len(p), cap(p))
		if int(n) < int(bytesToRead) {
			return int(n), fmt.Errorf("read fewer than expected bytes")
		}
		if err != nil {
			return 0, err
		}
		*c.bufStart = off
		*c.bufEnd = off + n
		log.Printf("bufStart %v bufEnd %v", *c.bufStart, *c.bufEnd)
		// todo error if can't fill len(p) -- edge case
		copy(p, c.buffer[int(off-*c.bufStart):int(off+int64(len(p))-*c.bufStart)])
		return len(p), nil
	}

	// requested range is partially in the buffer. Fall through for now. todo improve.

	n, err := c.client.GetRange(c.container, c.blobname, off, int64(len(p)), p)
	if err != nil {
		return 0, err
	}
	if int(n) < len(p) {
		return int(n), fmt.Errorf("read fewer than expected bytes")
	}

	return int(n), err
}

func (c BlobReader) Close() {}

func (c *azureBlobClient) GetRange(container string, blobname string, start int64, count int64, buf []byte) (int64, error) {
	log.Printf("reading %v %v %v", blobname, start, count)
	return c.client.DownloadBuffer(
		context.TODO(),
		container,
		blobname,
		buf,
		&azblob.DownloadBufferOptions{Range: azblob.HTTPRange{Offset: start, Count: count}},
	)
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
