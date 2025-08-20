package urltool

import (
	"errors"
	"net/url"
	"path"
)

// GetBasePath 获取URL路径的最后一节
func GetBasePath(targeturl string) (string, error) {
	myUrl, err := url.Parse(targeturl)
	if err != nil {
		return "", err
	}
	if len(myUrl.Host) == 0 {
		return "", errors.New("no host in targetUrl")
	}

	return path.Base(myUrl.Path), nil
}
