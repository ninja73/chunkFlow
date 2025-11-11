package tests

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"testing"
)

const distributorURL = "http://localhost:8080"

func uploadFile(t *testing.T, url, fieldName, filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("cannot open test file: %v", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(fieldName, file.Name())
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		t.Fatalf("failed to copy file: %v", err)
	}
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("upload returned non-200 code: %d", resp.StatusCode)
	}

	fileID, _ := io.ReadAll(resp.Body)

	return string(fileID)
}

func downloadFile(t *testing.T, url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("download failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("server returned non-200 code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("cannot read body: %v", err)
	}

	return data
}

func TestUploadDownload(t *testing.T) {
	testFile := "./testdata/test.bin"

	fileID := uploadFile(t, distributorURL+"/upload", "file", testFile)
	downloaded := downloadFile(t, distributorURL+"/download/"+fileID)

	original, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("cannot read original: %v", err)
	}

	if !bytes.Equal(downloaded, original) {
		t.Fatalf("files are not equal")
	}
}
