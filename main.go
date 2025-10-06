package main

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
)

func listFiles(dir string, allowedExtensions []string) (map[string]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	// build a quick-lookup set
	extSet := make(map[string]struct{}, len(allowedExtensions))
	for _, ext := range allowedExtensions {
		extSet[ext] = struct{}{}
	}

	out := make(map[string]string)
	for _, e := range entries {
		info, err := e.Info()
		if err != nil || info.Mode()&os.ModeDir != 0 { // skip dirs
			continue
		}

		name := e.Name()
		ext := strings.ToLower(filepath.Ext(name)) // includes the leading dot
		if _, ok := extSet[ext]; !ok {
			continue // unwanted extension
		}
		out[name] = filepath.Join(dir, name)
	}
	return out, nil
}

func metadataExists(filePath string) bool {
	if _, err := os.Stat(filePath + ".json"); errors.Is(err, os.ErrNotExist) {
		return false
	} else {
		return true
	}
}

func getFileMetadata(filePath string) (name string, image string) {
	return "Yuh", "pluh"
}

func main() {
	var daoutput string
	var dalist map[string]string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Value(&daoutput).
				// Height(8).
				Title("Experiences and Tools").
				OptionsFunc(func() []huh.Option[string] {
					options := make([]string, 0, len(dalist))
					for k, j := range dalist {
						if metadataExists(j) {
							name, image := getFileMetadata(j)
							m := k + " " + name + " " + image
							options = append(options, m)
						} else {
							options = append(options, k)
						}
					}
					return huh.NewOptions(options...)
				}, &dalist),
		),
	)

	path := "/Users/jamie/Desktop/Test"
	extensions := []string{"", ".app", ".exe"}

	if files, err := listFiles(path, extensions); err == nil {
		dalist = files
	} else {
		log.Fatal(err)
	}

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

	// cmd := exec.Command("open", dalist[daoutput])
	// if err := cmd.Run(); err != nil {
	// 	println(err.Error())
	// }
	println(dalist[daoutput])
}
