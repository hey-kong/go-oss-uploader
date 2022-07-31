package main

import (
	"log"
	"regexp"
	"strings"
	"time"

	"go-oss-uploader/common/constants"
	"go-oss-uploader/common/ossutil"
	"go-oss-uploader/common/util"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/radovskyb/watcher"
)

func main() {
	// Get required environment variables
	watchPath := util.GetEnvOrPanic("WATCH_DATA_PATH")
	jobKey := util.GetEnvOrPanic("JOB_KEY")
	pathPattern := util.GetEnv("PATH_PATTERN", ".*[^swp]$")
	watchInterval, err := util.GetEnvAsInt("WATCH_INTERVAL", 2)
	if err != nil {
		log.Fatalf("Unable to parse WATCH_INTERVAL as integer\n")
	}

	// Instantiate OSS client
	client, err := oss.New(constants.Endpoint, constants.AccessKeyID, constants.AccessKeySecret)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	// Instantiate OSS bucket
	bucket, err := client.Bucket(constants.Bucket)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	// Instantiate watcher
	w := watcher.New()
	w.FilterOps(watcher.Create, watcher.Write, watcher.Remove, watcher.Rename)

	// Only files that match the regular expression during file listings will be watched
	if pathPattern != "" {
		r := regexp.MustCompile(pathPattern)
		w.AddFilterHook(watcher.RegexFilterHook(r, false))
	}

	go func() {
		for {
			select {
			case event := <-w.Event:
				if event.IsDir() {
					continue
				}
				filePath := strings.Replace(event.Path, watchPath, "", 1)
				if event.Op == watcher.Remove {
					objectKey := ossutil.GenObjectKey(jobKey, constants.DataHub, filePath)
					go ossutil.Remove(bucket, objectKey)
				} else if event.Op == watcher.Rename {
					oldFilePath := strings.Replace(event.OldPath, watchPath, "", 1)
					srcObject := ossutil.GenObjectKey(jobKey, constants.DataHub, oldFilePath)
					destObject := ossutil.GenObjectKey(jobKey, constants.DataHub, filePath)
					go ossutil.Rename(bucket, srcObject, destObject)
				} else {
					objectKey := ossutil.GenObjectKey(jobKey, constants.DataHub, filePath)
					go ossutil.Upload(bucket, event.Path, objectKey)
				}
			case err = <-w.Error:
				log.Fatalln(err)
			case <-w.Closed:
				return
			}
		}
	}()

	// Watch folder recursively for changes
	if err = w.AddRecursive(watchPath); err != nil {
		log.Fatalln(err)
	}
	// Start the watching process
	log.Printf("Watching: %v\n", watchPath)
	if err = w.Start(time.Duration(watchInterval) * time.Second); err != nil {
		log.Fatalln(err)
	}
}
