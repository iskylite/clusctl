package utils

import (
	"crypto/md5"
	"fmt"
	"os"
	"time"

	"github.com/go-basic/uuid"
)

func Hostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}
	return hostname
}

func Md5sum(bytes []byte) string {
	return fmt.Sprintf("%x", md5.Sum(bytes))
}

func Isfile(fp string) bool {
	f, err := os.Stat(fp)
	if os.IsNotExist(err) {
		return false
	}
	return !f.IsDir()
}

func IsDir(fp string) bool {
	f, err := os.Stat(fp)
	if os.IsNotExist(err) {
		return false
	}
	return f.IsDir()
}

func UUID() string {
	return uuid.New()
}

func Timeout(timeout int) time.Duration {
	return time.Second * time.Duration(timeout)
}
