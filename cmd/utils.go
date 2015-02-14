package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var (
	ErrNoSbrfile = errors.New("Not in an 'sbr' workspace")
	ErrNoWd      = errors.New("Cannot find out the working dir")
)

func FindRootCmd() (dir string) {
	wd, err := FindRoot()
	if err != nil {
		fmt.Println(err)
		os.Exit(CodeNoWorkingDir)
	}
	return wd
}

//FindRoot get the current working dir and search for a .sbr file upwards
func FindRoot() (dir string, err error) {
	root, err := os.Getwd()
	if err != nil {
		log.Printf("getwd error %v", err)
		return root, ErrNoSbrfile
	}
	path := root
	//loop until I've reached the root, or found the .sbr
	for ; !FileExists(filepath.Join(path, ".sbr")) && path != "/"; path = filepath.Dir(path) {
	}

	if path != "/" {
		return path, nil
	} else {
		return root, ErrNoSbrfile
	}
}

//FileExists check if a path exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
