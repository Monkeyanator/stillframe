package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Monkeyanator/stillframe/pkg/stillframe"
)

var (
	videoPath    = flag.String("video", "", "Path to the video file")
	subtitlePath = flag.String("subtitle", "", "Path to the VTT subtitle file")
	outputPath   = flag.String("output", "", "Path for the output file")
)

func main() {
	flag.Parse()

	if *videoPath == "" || *subtitlePath == "" {
		fmt.Println("Error: all flags must be provided")
		flag.PrintDefaults()
		os.Exit(1)
	}

	result, err := stillframe.Render(*videoPath, *subtitlePath, *outputPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Rendered frame to %s\n", result.Path)
}
