//
// image.go
// a collection of image manipulation functions
// Copyright 2017 Akinmayowa Akinyemi
//

package utils

// cspell: ignore Lanczos, msword, docx, vndopenxmlformats, officedocumentwordprocessingmldocument

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/disintegration/imaging"
)

// ImgDataURIHeader the string sequence which is prefixed to image type data uris
const ImgDataURIHeader = "data:image/"

// FleDataURIHeader the string sequence which is prefixed to file type data uris
const FleDataURIHeader = "data:application/"

// SaveDataURI converts a datauri back to its original format
// the format type is deduced from the data uri header
func SaveDataURI(data []byte, fileName string) error {
	// read the header and check if the data supplied is a valid data uri
	// if !bytes.HasPrefix(data, []byte(ImgDataURIHeader)) || !bytes.HasPrefix(data, []byte(FleDataURIHeader)) {
	// 	return fmt.Errorf("data is not a valid datauri")
	// }
	if IsValidDataURI(data) == false {
		// log.Debugf("data is not a valid datauri: %s", string(data)[:15])
		return fmt.Errorf("data is not a valid datauri: %s", string(data)[:50])
	}

	dataHdr := []byte(ImgDataURIHeader)
	if bytes.HasPrefix(data, []byte(FleDataURIHeader)) {
		dataHdr = []byte(FleDataURIHeader)
	}
	src := bytes.Replace(data, dataHdr, []byte(""), 1)
	idx := bytes.Index(src, []byte(";"))
	if idx == -1 {
		return fmt.Errorf(
			"cant find mime type: %s --- %s",
			string(data)[:50], dataHdr,
		)
	}

	// get image data (in base64)
	src = src[idx+8 : len(src)]
	raw := make([]byte, len(src))
	_, err := base64.StdEncoding.Decode(raw, src)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(fileName, raw, 0644); err != nil {
		return err
	}

	return ioutil.WriteFile(fileName, raw, 0644)
}

// IsValidDataURI ...
func IsValidDataURI(data []byte) bool {
	// read the header and check if the data supplied is a valid data uri
	if bytes.HasPrefix(data, []byte(ImgDataURIHeader)) || bytes.HasPrefix(data, []byte(FleDataURIHeader)) {
		return true
	}

	return false
}

// GetDataURIType ...
func GetDataURIType(data []byte) (string, error) {

	if !IsValidDataURI(data) {
		return "", fmt.Errorf(
			"data is not a valid datauri: %s",
			string(data)[:50],
		)
	}

	dataHdr := []byte(ImgDataURIHeader)
	if bytes.HasPrefix(data, []byte(FleDataURIHeader)) {
		dataHdr = []byte(FleDataURIHeader)
	}
	src := bytes.Replace(data, dataHdr, []byte(""), 1)
	idx := bytes.Index(src, []byte(";"))
	if idx == -1 {
		return "", fmt.Errorf(
			"cant find mime type: %s --- %s",
			string(data)[:50], dataHdr,
		)
	}

	// get mime type
	mimeType := string(src[0:idx])
	if mimeType == "jpeg" {
		mimeType = "jpg"
	} else if strings.HasPrefix(mimeType, "svg") {
		mimeType = "svg"
	} else if strings.HasPrefix(mimeType, "pdf") {
		mimeType = "pdf"
	} else if strings.HasPrefix(mimeType, "msword") {
		mimeType = "doc"
	} else if strings.HasPrefix(mimeType, "vndopenxmlformats-officedocumentwordprocessingmldocument") {
		mimeType = "docx"
	} else if strings.HasPrefix(mimeType, "vnd.openxmlformats-officedocument.wordprocessingml.document") {
		mimeType = "docx"
	}

	return mimeType, nil
}

// ResizeImage ...
func ResizeImage(srcFile, dstFile string, width, height int) (err error) {

	srcImg, err := imaging.Open(srcFile)
	if err != nil {
		return nil
	}

	dstImg := imaging.Thumbnail(srcImg, width, height, imaging.Lanczos)
	err = imaging.Save(dstImg, dstFile)
	if err != nil {
		return
	}

	return nil
}
