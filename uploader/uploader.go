package uploader

import (
	"log"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/hey-kong/go-oss-uploader/common/configs"
	"github.com/hey-kong/go-oss-uploader/common/ossutil"
	"github.com/radovskyb/watcher"
)

func Upload() {
	// Instantiate OSS client
	client, err := oss.New(configs.Endpoint, configs.AccessKeyID, configs.AccessKeySecret)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	// Instantiate OSS bucket
	bucket, err := client.Bucket(configs.Bucket)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	// Instantiate watcher
	w := watcher.New()
	w.FilterOps(watcher.Create, watcher.Write, watcher.Remove, watcher.Rename)

	// Only files that match the regular expression during file listings will be watched
	if configs.PathPattern != "" {
		r := regexp.MustCompile(configs.PathPattern)
		w.AddFilterHook(watcher.RegexFilterHook(r, false))
	}

	go func() {
		for {
			select {
			case event := <-w.Event:
				if event.IsDir() {
					continue
				}
				file := strings.Replace(event.Path, configs.WatchPath, "", 1)
				if event.Op == watcher.Remove {
					objectKey := path.Join(configs.KeyPrefix, file)
					go ossutil.Remove(bucket, objectKey)
				} else if event.Op == watcher.Rename {
					oldFilePath := strings.Replace(event.OldPath, configs.WatchPath, "", 1)
					srcObject := path.Join(configs.KeyPrefix, oldFilePath)
					destObject := path.Join(configs.KeyPrefix, file)
					go ossutil.Rename(bucket, srcObject, destObject)
				} else {
					objectKey := path.Join(configs.KeyPrefix, file)
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
	if err = w.AddRecursive(configs.WatchPath); err != nil {
		log.Fatalln(err)
	}

	// Start the watching process
	log.Printf("Watching: %v\n", configs.WatchPath)
	if err = w.Start(time.Duration(configs.WatchInterval) * time.Second); err != nil {
		log.Fatalln(err)
	}
}
