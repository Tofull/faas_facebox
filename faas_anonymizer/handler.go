package main

// Anonymizer for OpenFaaS project
// Strongly inspired from https://blog.machinebox.io/how-i-built-an-image-proxy-server-to-anonymise-images-in-twenty-minutes-e550466ea09e
// Hacked by Loïc Messal (@tofull)

import (
	"encoding/binary"
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"github.com/machinebox/sdk-go/facebox"
)

func main() {
	
	var (
		faceboxDefault string = "http://localhost:yourFaceBoxPort"
		faceboxAddr = flag.String("facebox", faceboxDefault, "Facebox address")
	)
	
	// overwrite faceboxDefault with os env var
	if faceboxDefault := os.Getenv("facebox"); faceboxDefault != "" {
		faceboxAddr = &faceboxDefault
	}


	flag.Parse()
	fb := facebox.New(*faceboxAddr)
	
	// Read URL from stdin
	input, err := ioutil.ReadAll(os.Stdin)
	
	if err != nil {
		log.Fatalf("Cannot read input %s.\n", err)
        return
	}

	// Download the image from url
	resp, err := http.Get(string(input))
	if err != nil {
		fmt.Printf("download failed: "+err.Error())
		return
	}
	
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Printf("download failed: "+resp.Status)
		return
	}
	
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("download failed: "+err.Error())
		return
	}
	
	img, format, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		fmt.Printf("image: "+err.Error())
		return
	}
	
	// Detect faces from the image
	faces, err := fb.Check(bytes.NewReader(b))
	if err != nil {
		fmt.Printf("facebox: "+err.Error())
		return
	}

	// Anonymise it
	anonImg := anonymise(img, faces)

	// Write binary data in stdout
	buf := new(bytes.Buffer)
	switch format {
	case "jpeg":
			err := jpeg.Encode(buf, anonImg, &jpeg.Options{Quality: 100})
			if err != nil {
				fmt.Printf(err.Error())
				return
			}
		case "gif":
			err := gif.Encode(buf, anonImg, nil)
			if err != nil {
				fmt.Printf(err.Error())
				return
			}
		case "png":
			err := png.Encode(buf, anonImg)
			if err != nil {
				fmt.Printf(err.Error())
				return
			}
		default:
			fmt.Printf("unsupported format: "+format)
			return
	}

	binary.Write(os.Stdout, binary.LittleEndian, buf.Bytes())
}

// anonymise produces a new image with faces redacted.
// see https://becominghuman.ai/anonymising-images-with-go-and-machine-box-fd0866adb9f5
func anonymise(src image.Image, faces []facebox.Face) image.Image {
	dstImage := image.NewRGBA(src.Bounds())
	draw.Draw(dstImage, src.Bounds(), src, image.ZP, draw.Src)
	for _, face := range faces {
		faceRect := image.Rect(
			face.Rect.Left,
			face.Rect.Top,
			face.Rect.Left+face.Rect.Width,
			face.Rect.Top+face.Rect.Height,
		)
		facePos := image.Pt(face.Rect.Left, face.Rect.Top)
		draw.Draw(
			dstImage,
			faceRect,
			&image.Uniform{color.Black},
			facePos,
			draw.Src)
	}
	return dstImage
}