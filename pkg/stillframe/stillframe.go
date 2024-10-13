package stillframe

import (
	"fmt"
	"image/color"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/asticode/go-astisub"
	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"github.com/google/uuid"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
)

// Result is the result from a render operation.
type Result struct {
	Path      string
	Text      string
	Timestamp string
}

// Render renders the still frame with the subtitle overlay.
func Render(videoPath, subtitlePath, outPath string) (*Result, error) {
	if outPath == "" {
		tmpDir := filepath.Join(os.TempDir(), "stillframe")
		outPath = filepath.Join(tmpDir, fmt.Sprintf("%s.%s", uuid.NewString(), "png"))
	}

	subs, err := astisub.OpenFile(subtitlePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read subtitle track: %w", err)
	}

	seed := rand.NewSource(time.Now().UnixNano())
	r := rand.New(seed)
	item := subs.Items[r.Int()%len(subs.Items)]

	imagePath, err := exportFrame(videoPath, item.StartAt)
	defer os.Remove(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to export frame: %w", err)
	}

	if err := overlayText(imagePath, outPath, item.String()); err != nil {
		return nil, fmt.Errorf("failed to overlay text: %w", err)
	}

	return &Result{
		Path:      outPath,
		Text:      item.String(),
		Timestamp: item.StartAt.String(),
	}, nil
}

func exportFrame(videoPath string, offset time.Duration) (string, error) {
	tmpDir := filepath.Join(os.TempDir(), "stillframe")
	err := os.MkdirAll(tmpDir, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("failed creating output dir: %v", err)
	}

	dstPath := filepath.Join(tmpDir, fmt.Sprintf("%s.%s", uuid.NewString(), "jpg"))
	formattedOffset := time.Unix(0, 0).UTC().Add(time.Duration(offset)).Format("15:04:05.00")

	if err := ffmpeg.
		Input(videoPath, ffmpeg.KwArgs{"ss": formattedOffset}).
		Output(dstPath, ffmpeg.KwArgs{
			"frames:v": "1",
			"q:v":      "2",
		}).
		OverWriteOutput().
		Run(); err != nil {
		return "", fmt.Errorf("failed generating image from %s: %w", videoPath, err)
	}

	return dstPath, nil
}

func overlayText(srcPath, dstPath, text string) error {
	img, err := gg.LoadImage(srcPath)
	if err != nil {
		return fmt.Errorf("failed to load image: %w", err)
	}

	imgCtx := gg.NewContextForImage(img)
	imgCtx.Push()
	tuneFontSize(imgCtx, text)
	textX := float64(imgCtx.Width()) / 2
	textY := float64(imgCtx.Height()) - imgCtx.FontHeight() - 50
	drawOutlinedText(imgCtx, text, textX, textY)
	imgCtx.Fill()
	imgCtx.Pop()
	return imgCtx.SavePNG(dstPath)
}

func tuneFontSize(imgCtx *gg.Context, text string) error {
	maxWidth := float64(imgCtx.Width()) * 0.8
	maxHeight := float64(imgCtx.Height()) * 0.1

	var prevFontFace font.Face
	size := 12.0
	for {
		fontFace, err := createUpscaledFont(size)
		if err != nil {
			return err
		}

		imgCtx.SetFontFace(fontFace)
		w, h := imgCtx.MeasureString(text)
		if w > maxWidth || h > maxHeight {
			if prevFontFace != nil {
				imgCtx.SetFontFace(prevFontFace)
			}
			break
		}

		prevFontFace = fontFace
		size += 10
	}

	return nil
}

func createUpscaledFont(size float64) (font.Face, error) {
	f, err := truetype.Parse(goregular.TTF)
	if err != nil {
		return nil, err
	}

	return truetype.NewFace(f, &truetype.Options{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	}), nil
}

func drawOutlinedText(imgCtx *gg.Context, text string, x, y float64) {
	offsets := []struct {
		dx, dy float64
	}{
		{-1, -1},
		{1, -1},
		{-1, 1},
		{1, 1},
	}

	imgCtx.SetColor(color.Black)
	for _, offset := range offsets {
		imgCtx.DrawStringAnchored(text, x+offset.dx, y+offset.dy, 0.5, 1)
	}

	imgCtx.SetRGB(1, 1, 0)
	imgCtx.DrawStringAnchored(text, x, y, 0.5, 1)
}
