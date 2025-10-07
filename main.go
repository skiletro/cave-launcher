package main

import (
	"encoding/json"
	"errors"
	"image/color"
	"io"
	"os"
	"path/filepath"
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
		metadataFile, err := os.Open(filePath + ".json")
		if err != nil {
			panic("aaaa!!! file doesn't exist somehow")
		}
		defer metadataFile.Close()

		byteValue, _ := io.ReadAll(metadataFile)

		var parsedData map[string]string
		json.Unmarshal([]byte(byteValue), &parsedData)

		return parsedData["name"], parsedData["icon"], nil
	}
}

func createButtons(files []shortcutEntry) []*fyne.Container {
	buttons := make([]*fyne.Container, len(files))

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

		label := container.New(layout.NewCenterLayout(), canvas.NewText(labelString, color.White))

		img := canvas.NewImageFromResource(imageResource)
		img.FillMode = canvas.ImageFillStretch
		img.SetMinSize(fyne.NewSize(1920/5, 1080/5))

		button := widget.NewButton("Launch", func() {
			dialog.Message("%s", file.filepath).Title("Clicked").Info()
		})

		buttons[index] = container.New(layout.NewVBoxLayout(), label, img, button)
	}
	return buttons
}

func main() {
	allowedExtensions := []string{"", ".exe", ".lnk"}
	files, _ := listFiles("/Users/jamie/Desktop/Test", allowedExtensions)

	a := app.New()
	w := a.NewWindow("Hello")

	content := container.New(layout.NewGridLayout(3))
	for _, b := range createButtons(files) {
		content.Add(b)
	}

	scrollContainer := container.NewScroll(content)
	scrollContainer.SetMinSize(fyne.NewSize(1280, 720))

	w.SetContent(scrollContainer)

	w.ShowAndRun()
}
