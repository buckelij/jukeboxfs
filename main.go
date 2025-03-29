package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"strings"
	"sync"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Info("hello world")
	c, err := NewBlobClient(os.Getenv("BLOB_ACCOUNT"), os.Getenv("BLOB_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	/* buf := make([]byte, 1024)
	n, err := c.GetRange(
		"data",
		"test.zip",
		0,
		400,
		buf,
	)
	log.Printf("bytes downloaded: %v %v", n, err) */

	// test.zip 581 , content.zip 84723573
	blobreader := BlobReader{
		client:    c,
		mu:        new(sync.Mutex),
		container: "data",
		blobname:  "content.zip",
		blobsize:  84723573,
		bufStart:  new(int64),
		bufEnd:    new(int64),
		buffer:    make([]byte, 16*1024*1024), // DefaultDownloadBlockSize is 4mb so 16/4 goroutines
	}
	r, err := zip.NewReader(io.ReaderAt(blobreader), blobreader.blobsize)
	if err != nil {
		log.Fatal(err)
	}

	// Iterate through the files in the archive,
	// printing some of their contents.
	for _, f := range r.File {
		logger.Info(fmt.Sprintf("Contents of %s:\n", f.Name))

		if f.FileHeader.FileInfo().IsDir() {
			continue
		}

		dst, _ := os.Create("/tmp/" + strings.Replace(f.Name, "/", "_", -1))
		defer dst.Close()

		rc, err := f.Open()
		if err != nil {
			log.Fatal(err)
		}
		// _, err = io.Copy(os.Stdout, rc)
		_, err = io.Copy(dst, rc)
		if err != nil {
			log.Fatal(err)
		}
		rc.Close()
	}
}
