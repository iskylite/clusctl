package utils

import (
	"crypto/md5"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
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

func ConvertSize(size string) (int, error) {
	re := regexp.MustCompile(`^(\d+)([kKmM]?)$`)
	matches := re.FindAllStringSubmatch(size, -1)
	if len(matches) == 0 || len(matches[0]) != 3 {
		return 0, fmt.Errorf("block size syntax error")
	}
	blockSizeStr := matches[0][1]
	blockSize, err := strconv.Atoi(blockSizeStr)
	if err != nil {
		return 0, err
	}
	unit := strings.ToLower(matches[0][2])
	switch unit {
	case "m":
		blockSize *= 1024 * 1024
	case "k":
		blockSize *= 1024
	case "":

	default:
		return 0, fmt.Errorf("block size syntax error")
	}
	return blockSize, nil
}
