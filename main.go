package main

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/sqweek/dialog"
)

type shortcutEntry struct {
	filename   string
	filedir    string
	filepath   string
	customname string
	customicon string
}

//go:generate fyne bundle -o bundled.go fallback.png

func listFiles(dir string, allowedExtensions []string) ([]shortcutEntry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	// build a quick-lookup set
	extSet := make(map[string]struct{}, len(allowedExtensions))
	for _, ext := range allowedExtensions {
		extSet[ext] = struct{}{}
	}

	var out []shortcutEntry // we havent a clue ow long itll be
	for _, e := range entries {
		// skip directories
		info, err := e.Info()
		if err != nil || info.Mode()&os.ModeDir != 0 {
			continue
		}

		name := e.Name()
		ext := strings.ToLower(filepath.Ext(name)) // includes the leading dot
		if _, ok := extSet[ext]; !ok {
			continue // unwanted extension
		}

		entry := shortcutEntry{
			filename:   name,
			filedir:    dir,
			filepath:   filepath.Join(dir, name),
			customicon: "gopher.png",
		}

		out = append(out, entry)
	}

	return out, nil
}

func fetchMetadataIfExists(filePath string) (prettyName string, imagePath string, err error) {
	if _, err := os.Stat(filePath + ".json"); errors.Is(err, os.ErrNotExist) {
		return "", "", errors.ErrUnsupported
	} else {
		metadataFile, _ := os.Open(filePath + ".json")

		defer metadataFile.Close()

		byteValue, _ := io.ReadAll(metadataFile)

		var parsedData map[string]string
		json.Unmarshal([]byte(byteValue), &parsedData)

		return parsedData["name"], parsedData["icon"], nil
	}
}

func createButtons(files []shortcutEntry) []*fyne.Container {
	entries := make([]*fyne.Container, len(files))

	for index, file := range files {
		labelString := "file: " + file.filename
		var imageResource fyne.Resource = resourceFallbackPng

		prettyName, imagePath, err := fetchMetadataIfExists(file.filepath)
		if err == nil {
			if prettyName != "" {
				labelString = prettyName
			}

			if imagePath != "" {
				imageResource, _ = fyne.LoadResourceFromPath(file.filedir + "/" + imagePath)
			}
		}

		// label := container.New(layout.NewCenterLayout(), canvas.NewText(labelString, color.White))
		label := widget.NewRichTextWithText(labelString)
		label.Wrapping = fyne.TextWrapWord

		img := canvas.NewImageFromResource(imageResource)
		img.FillMode = canvas.ImageFillStretch
		img.SetMinSize(fyne.NewSize(180, 135))

		button := widget.NewButton("Launch", func() {
			dialog.Message("%s", file.filepath).Title("Clicked").Info()
		})

		entry := container.New(layout.NewVBoxLayout(), label, img, button)

		entries[index] = entry
	}
	return entries
}

func main() {
	var (
		fileDir           string
		allowedExtensions []string
	)

	switch runtime.GOOS {
	case "darwin":
		fileDir = "/Users/jamie/Desktop/Test"
		allowedExtensions = []string{"", ".lnk"}
	case "windows":
		fileDir = "C:\\Users\\mmu\\Desktop"
		allowedExtensions = []string{".exe", ".lnk"}
	default:
		panic("File Dir not set! Cannot detect the OS.")
	}
	files, err := listFiles(fileDir, allowedExtensions)
	if err != nil {
		panic("Can't get files.")
	}

	a := app.New()
	w := a.NewWindow("CAVE Launcher")

	var individualEntryWidth float32 = 180.0
	var columnAmountByDefault float32 = 3.0

	content := container.NewGridWrap(fyne.NewSize(individualEntryWidth, 250))
	for _, b := range createButtons(files) {
		content.Add(b)
	}

	scrollContainer := container.NewScroll(content)
	scrollContainer.SetMinSize(fyne.NewSize((individualEntryWidth*columnAmountByDefault)+(4*(columnAmountByDefault-1)), 450))

	w.SetContent(scrollContainer)
	w.SetFixedSize(true)

	w.ShowAndRun()
}
