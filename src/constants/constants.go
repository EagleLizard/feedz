package constants

import (
	"log"
	"path/filepath"
	"runtime"
)

const DataDirName = "data"
const TestFeedsFileName = "test_feeds.txt"

var BaseDir string
var DataDirPath string
var TestFeedsFilePath string
var TestFeedFilePath string

func init() {
	BaseDir = getBaseDir()
	DataDirPath = filepath.Join(BaseDir, DataDirName)
	TestFeedsFilePath = filepath.Join(DataDirPath, TestFeedsFileName)
	TestFeedFilePath = filepath.Join(DataDirPath, "feed.syntax.fm.xml")
}

func getBaseDir() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Error getting BaseDir")
	}
	baseDir := filepath.Dir(filepath.Dir(filepath.Dir(filename)))
	return baseDir
}
