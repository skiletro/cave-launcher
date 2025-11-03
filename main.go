package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type shortcutEntry struct {
	filename string
	filedir  string
	filepath string
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
			filename: name,
			filedir:  dir,
			filepath: filepath.Join(dir, name),
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

func openFile(path string) error {
	switch runtime.GOOS {
	case "windows":
		_, err := Start("cmd", "/c", "start", "", path)
		return err
	case "darwin":
		return exec.Command("open", path).Start()
	default: // linux and others
		return exec.Command("xdg-open", path).Start()
	}
}

func Start(args ...string) (p *os.Process, err error) {
	if args[0], err = exec.LookPath(args[0]); err == nil {
		var procAttr os.ProcAttr
		procAttr.Files = []*os.File{
			os.Stdin,
			os.Stdout, os.Stderr,
		}
		p, err := os.StartProcess(args[0], args, &procAttr)
		if err == nil {
			return p, nil
		}
	}
	return nil, err
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
				imageResource, _ = fyne.LoadResourceFromPath(filepath.Join(file.filedir, imagePath))
			}
		}

		label := widget.NewRichTextWithText(labelString)
		label.Truncation = fyne.TextTruncateEllipsis
		label.Resize(fyne.NewSize(180, 50))

		labelContainer := container.NewWithoutLayout(label)
		labelContainer.Resize(fyne.NewSize(180, 50))

		img := canvas.NewImageFromResource(imageResource)
		img.FillMode = canvas.ImageFillStretch
		img.SetMinSize(fyne.NewSize(180, 135))

		button := widget.NewButton("Launch", func() {
			openFile(file.filepath)
		})

		entry := container.New(layout.NewVBoxLayout(), labelContainer, img, button)

		entries[index] = entry
	}
	return entries
}

func main() {
	a := app.New()
	w := a.NewWindow("CAVE Launcher")

	executable, err := os.Executable()
	if err != nil {
		panic(err)
	}

	var fileDir string = filepath.Join(filepath.Dir(executable), "shortcuts")
	fmt.Println(fileDir)
	var allowedExtensions []string = []string{".exe", ".lnk", ""}

	var content *fyne.Container = container.NewCenter(widget.NewRichTextWithText("No shortcuts found."))

	var individualEntryWidth float32 = 180.0
	var columnAmountByDefault float32 = 3.0

	files, err := listFiles(fileDir, allowedExtensions)
	if err == nil {
		content = container.NewGridWrap(fyne.NewSize(individualEntryWidth, 250))
		for _, b := range createButtons(files) {
			content.Add(b)
		}
	}

	scrollContainer := container.NewScroll(content)
	scrollContainer.SetMinSize(fyne.NewSize((individualEntryWidth*columnAmountByDefault)+(4*(columnAmountByDefault-1)), 450))

	w.SetContent(scrollContainer)
	w.SetFixedSize(true)

	w.ShowAndRun()
}
