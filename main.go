package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gingerycode/tgspoil/spoiler"
)

const (
	fps             = 30
	durationSeconds = 5
	outputDir       = "frames"
	outputVideo     = "spoiler.mp4"

	dotDensity    = 0.004 // % of pixels covered in dots per frame
	blurSigma     = 45.0
	dotSize       = 6
	shimmerSpeed  = 9
	darkenPercent = 0.3
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Enter source image path")
		return
	}

	imagePath := os.Args[1]
	totalFrames := fps * durationSeconds

	if err := os.RemoveAll(outputDir); err != nil {
		fmt.Printf("Failed to clean output directory: %v\n", err)
		return
	}
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		fmt.Printf("Failed to create output directory: %v\n", err)
		return
	}

	err := spoiler.GenerateSpoilerFrames(
		imagePath,
		outputDir,
		totalFrames,
		blurSigma,
		dotDensity,
		dotSize,
		shimmerSpeed,
		darkenPercent,
	)
	if err != nil {
		fmt.Printf("Error generating frames: %v\n", err)
		return
	}

	cmd := exec.Command("ffmpeg",
		"-y",
		"-framerate", fmt.Sprint(fps),
		"-i", filepath.Join(outputDir, "frame_%04d.png"),
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		outputVideo,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("ffmpeg failed: %v\n", err)
		return
	}

	if err := os.RemoveAll(outputDir); err != nil {
		fmt.Printf("Warning: failed to remove frames folder: %v\n", err)
	}

	fmt.Println("Spoiler video created:", outputVideo)
}
