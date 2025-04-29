package main

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

func main() {
	cred, _ := azblob.NewSharedKeyCredential("devstoreaccount1", "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==")
	c, err := azblob.NewClientWithSharedKeyCredential("http://127.0.0.1:10000/devstoreaccount1", cred, nil)
	if err != nil {
		fmt.Print(err)
	}
	_, err = c.CreateContainer(context.Background(), "test", nil)
	if err != nil {
		fmt.Print(err)
	}
	testzip64 := `UEsDBBQAAAAAAPA4TFYAAAAAAAAAAAAAAAAIACAAY29udGVudC9VVA0AB7QA6WO5AOljtADpY3V4
CwABBPUBAAAEFAAAAFBLAwQUAAgACADwOExWAAAAAAAAAAAKAAAAEAAgAGNvbnRlbnQvbW9vbi50
eHRVVA0AB7QA6WO2AOljtADpY3V4CwABBPUBAAAEFAAAAEvPz0/Jy0zPKOECAFBLBwgzmDbIDAAA
AAoAAABQSwMEFAAIAAgA6zhMVgAAAAAAAAAACwAAABEAIABjb250ZW50L2hlbGxvLnR4dFVUDQAH
qgDpY6wA6WOqAOljdXgLAAEE9QEAAAQUAAAAy0jNyckvzy/KSeECAFBLBwjmMkmaDQAAAAsAAABQ
SwECFAMUAAAAAADwOExWAAAAAAAAAAAAAAAACAAgAAAAAAAAAAAA7UEAAAAAY29udGVudC9VVA0A
B7QA6WO5AOljtADpY3V4CwABBPUBAAAEFAAAAFBLAQIUAxQACAAIAPA4TFYzmDbIDAAAAAoAAAAQ
ACAAAAAAAAAAAACkgUYAAABjb250ZW50L21vb24udHh0VVQNAAe0AOljtgDpY7QA6WN1eAsAAQT1
AQAABBQAAABQSwECFAMUAAgACADrOExW5jJJmg0AAAALAAAAEQAgAAAAAAAAAAAApIGwAAAAY29u
dGVudC9oZWxsby50eHRVVA0AB6oA6WOsAOljqgDpY3V4CwABBPUBAAAEFAAAAFBLBQYAAAAAAwAD
ABMBAAAcAQAAAAA=`
	b := make([]byte, base64.StdEncoding.DecodedLen(len(testzip64)))
	base64.StdEncoding.Decode(b, []byte(testzip64))
	_, err = c.UploadBuffer(context.Background(), "test", "content.zip", b, nil)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Print("uploaded content.zip to http://127.0.0.1:10000/devstoreaccount1/test")
}
