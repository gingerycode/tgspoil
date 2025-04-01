package spoiler

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/disintegration/imaging"
)

type shimmerDot struct {
	X, Y  int
	Phase float64 // between 0 and 2π
	Gray  uint8   // fixed R=G=B
}

func ensureEvenSize(img image.Image) image.Image {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w%2 == 0 && h%2 == 0 {
		return img
	}
	nw, nh := w+(w%2), h+(h%2)
	return imaging.PasteCenter(imaging.New(nw, nh, color.Black), img)
}

func generateShimmerDots(img image.Image, count int, grayMin, grayMax int) []shimmerDot {
	bounds := img.Bounds()
	dots := make([]shimmerDot, count)

	for i := range dots {
		gray := grayMin + rand.Intn(grayMax-grayMin+1)
		dots[i] = shimmerDot{
			X:     rand.Intn(bounds.Dx()),
			Y:     rand.Intn(bounds.Dy()),
			Phase: rand.Float64() * 2 * math.Pi,
			Gray:  uint8(gray),
		}
	}

	return dots
}

func drawShimmerFrame(base image.Image, dots []shimmerDot, frame int, totalFrames int, dotSize int, shimmerSpeed float64, darkenPercent float64) *image.RGBA {
	bounds := base.Bounds()
	out := image.NewRGBA(bounds)
	draw.Draw(out, bounds, base, image.Point{}, draw.Src)

	darkenAmount := int(float64(255) * darkenPercent)
	overlayColor := color.RGBA{0, 0, 0, uint8(darkenAmount)}
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			bg := out.RGBAAt(x, y)
			blended := alphaBlend(overlayColor, bg)
			out.SetRGBA(x, y, blended)
		}
	}

	// centerFrame := float64(frame) / float64(totalFrames) * 2 * math.Pi
	centerFrame := shimmerSpeed * 2 * math.Pi * (float64(frame) / float64(totalFrames))

	for _, dot := range dots {
		alpha := 0.5 + 0.5*math.Sin(centerFrame+dot.Phase)
		a := uint8(alpha * 255)
		dotColor := color.RGBA{dot.Gray, dot.Gray, dot.Gray, a}

		for dx := 0; dx < dotSize; dx++ {
			for dy := 0; dy < dotSize; dy++ {
				x := dot.X + dx
				y := dot.Y + dy
				if x < bounds.Dx() && y < bounds.Dy() {
					// Blend dot over the image
					bg := out.RGBAAt(x, y)
					blend := alphaBlend(dotColor, bg)
					out.SetRGBA(x, y, blend)
				}
			}
		}
	}

	return out
}

func alphaBlend(fg, bg color.RGBA) color.RGBA {
	alpha := float64(fg.A) / 255
	invAlpha := 1 - alpha
	return color.RGBA{
		R: uint8(float64(fg.R)*alpha + float64(bg.R)*invAlpha),
		G: uint8(float64(fg.G)*alpha + float64(bg.G)*invAlpha),
		B: uint8(float64(fg.B)*alpha + float64(bg.B)*invAlpha),
		A: 255,
	}
}

func GenerateSpoilerFrames(imagePath, outputDir string, totalFrames int, blurSigma float64, dotDensity float64, dotSize int, shimmerSpeed float64, darkenPercent float64) error {
	rand.Seed(time.Now().UnixNano())

	srcImg, err := imaging.Open(imagePath)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}

	srcImg = ensureEvenSize(srcImg)
	blurred := imaging.Blur(srcImg, blurSigma)

	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	dotCount := int(float64(blurred.Bounds().Dx()*blurred.Bounds().Dy()) * dotDensity)
	if dotCount < 1 {
		return fmt.Errorf("dot count must be at least 1")
	}

	dots := generateShimmerDots(blurred, dotCount, 220, 250)

	for i := range totalFrames {
		frame := drawShimmerFrame(blurred, dots, i, totalFrames, dotSize, shimmerSpeed, darkenPercent)
		path := filepath.Join(outputDir, fmt.Sprintf("frame_%04d.png", i))
		file, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("failed to create frame: %w", err)
		}
		if err := png.Encode(file, frame); err != nil {
			return fmt.Errorf("failed to encode frame: %w", err)
		}
		file.Close()
	}

	return nil
}
