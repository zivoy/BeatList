package main

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
)

var defaultLoc fyne.URI
var lastOpened fyne.URI
var CacheDir fyne.URI
var logFile fyne.URI

func initStorage(a fyne.App) {
	// default file loc
	executable, err := os.Executable()
	loc := "playlist.bplist"
	if err == nil {
		loc = path.Join(filepath.Dir(executable), loc)
	}
	defaultLoc = storage.NewFileURI(loc)
	lastOpened, _ = storage.ParseURI(a.Preferences().StringWithFallback("lastOpened", defaultLoc.String()))

	// caching
	CacheDir = storage.NewFileURI(filepath.Join(os.TempDir(), "BeatList"))
	err = os.MkdirAll(CacheDir.Path(), os.ModePerm)
	if err != nil {
		CacheDir, _ = storage.Child(a.Storage().RootURI(), "Cache")
		_ = os.MkdirAll(CacheDir.Path(), os.ModePerm)
	}
}

func initLogging(a fyne.App) io.Closer {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	logFile, _ = storage.Child(a.Storage().RootURI(), "latest.log")
	err := storage.Delete(logFile)
	if err != nil {
		log.Println(err)
	}
	f, err := storage.Writer(logFile)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
		return nil
	}

	multi := io.MultiWriter(f, os.Stdout)
	log.SetOutput(multi)
	return f
}

func hash(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

func getCached(idName string, readExisting func(reader io.Reader) (interface{}, error), isBlank func(item interface{}) bool,
	getNew func() (interface{}, error), storeNew func(writer io.Writer, item interface{}) error) interface{} {
	cacheLoc, _ := storage.Child(CacheDir, idName)
	var cacheInfo interface{}

	// try getting existing
	exists, err := storage.Exists(cacheLoc)
	if err == nil && exists {
		r, _ := storage.Reader(cacheLoc)
		cacheInfo, err = readExisting(r)
		defer closeFile(r)
		if isBlank(cacheInfo) || err != nil {
			_ = storage.Delete(cacheLoc)
			log.Printf("%s encountered an error or was blank, %e", idName, err)
			//return nil
		} else {
			return cacheInfo
		}
	} else if err != nil {
		log.Println("error in file existence", err)
	}

	// get new
	cacheInfo, err = getNew()
	if err != nil {
		log.Println("problem fetching new item", err)
		return nil
	}
	// write to cache
	w, err := storage.Writer(cacheLoc)
	if w != nil {
		defer func(w fyne.URIWriteCloser) {
			err := w.Close()
			if err != nil {
				log.Println("problem saving", idName, err)
			}
		}(w)
	}
	if err != nil {
		log.Println(err)
	} else {
		err = storeNew(w, cacheInfo)
		if err != nil {
			log.Println(err)
		}
	}
	return cacheInfo
}

func readLogs() string {
	reader, err := storage.Reader(logFile)
	defer closeFile(reader)
	if err != nil {
		log.Println(err)
		return ""
	}
	logs, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Println(err)
		return ""
	}
	return strings.Trim(string(logs), "\n ")
}

func closeFile(file io.Closer) {
	if file != nil {
		err := file.Close()
		if err != nil {
			if _, file, no, ok := runtime.Caller(1); ok {
				log.Printf("%s:%d: %s", filepath.Base(file), no, err)
			} else {
				log.Println(err)
			}
		}
	}
}

func getBaseDir(uri fyne.URI) fyne.URI {
	return storage.NewFileURI(filepath.Dir(uri.Path()))
}
