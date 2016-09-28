package server

import (
	"crypto/sha256"
	"encoding/base64"
	log "github.com/Sirupsen/logrus"
	"github.com/oxfeeefeee/appgo"
	"github.com/oxfeeefeee/appgo/toolkit"
	"io"
	"os"
	"strings"
)

type versioning struct {
	prefixMap map[string]string
	hashCache *toolkit.Map
}

func newVersioning() *versioning {
	return &versioning{
		make(map[string]string),
		toolkit.NewMap(),
	}
}

func (v *versioning) addMap(prefix, dir string) {
	v.prefixMap[prefix] = dir
}

func (v *versioning) getStatic(path string) string {
	return "//" + appgo.Conf.CdnDomain + v.getStaticPathParam(path)
}

func (v *versioning) getStaticPathParam(path string) string {
	if h := v.hashCache.GetString(path); h != "" {
		return path + "?ver=" + h
	}
	filepath := ""
	for prefix, dir := range v.prefixMap {
		if strings.HasPrefix(path, prefix) {
			filepath = dir + "/" + path[len(prefix):]
		}
	}
	if filepath == "" {
		return path
	}
	f, err := os.Open(filepath)
	if err != nil {
		log.WithFields(log.Fields{
			"error":    err,
			"filepath": filepath,
		}).Error("Error reading static file")
		return path
	}
	defer f.Close()
	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		log.WithFields(log.Fields{
			"error":    err,
			"filepath": filepath,
		}).Error("Error finding sha256 of static file")
		return path
	}
	h := base64.RawURLEncoding.EncodeToString(hasher.Sum(nil))
	v.hashCache.Add(path, h)
	return path + "?ver=" + h
}
