package qiniu

import (
	"github.com/oxfeeefeee/appgo"
	qndigest "github.com/qiniu/api.v6/auth/digest"
	"github.com/qiniu/api.v6/conf"
	qnio "github.com/qiniu/api.v6/io"
	"github.com/qiniu/api.v6/rs"
	qnurl "github.com/qiniu/api.v6/url"
	"io"
	"strconv"
	"strings"
)

var (
	bucketName string
	domain     string
	putPolicy  rs.PutPolicy
	getPolicy  rs.GetPolicy
)

type QiniuParams struct {
	AccessKey      string
	Secret         string
	Bucket         string
	Domain         string
	DefaultExpires int
}

func Init() {
	params := appgo.Conf.Qiniu
	conf.ACCESS_KEY = params.AccessKey
	conf.SECRET_KEY = params.Secret
	bucketName = params.Bucket
	domain = params.Domain
	putPolicy = rs.PutPolicy{
		Expires: uint32(params.DefaultExpires),
	}
	getPolicy = rs.GetPolicy{
		Expires: uint32(params.DefaultExpires),
	}
}

func GetUrl(key string, expiryUnix int64, params string) string {
	baseUrl := makeBaseUrl(key)
	url := baseUrl + "?" + params
	if expiryUnix == 0 {
		return getPolicy.MakeRequest(url, nil)
	}
	return makeRequest(url, expiryUnix, nil)
}

func PublicGetUrl(key string) string {
	return makeBaseUrl(key)
}

func PutToken(key string) string {
	putPolicy.Scope = bucketName + ":" + key
	return putPolicy.Token(nil)
}

func PutFile(key string, r io.Reader, size int64) error {
	putPolicy.Scope = bucketName + ":" + key
	token := putPolicy.Token(nil)
	//var ret qnio.PutRet
	return qnio.Put2(nil, nil /*&ret*/, token, key, r, size, nil)
}

func makeBaseUrl(key string) string {
	//return rs.MakeBaseUrl(domain, key)
	return "http://" + domain + "/" + qnurl.Escape(key)
}

func makeRequest(baseUrl string, expiryUnix int64, mac *qndigest.Mac) (privateUrl string) {
	if strings.Contains(baseUrl, "?") {
		baseUrl += "&e="
	} else {
		baseUrl += "?e="
	}
	baseUrl += strconv.FormatInt(expiryUnix, 10)

	token := qndigest.Sign(mac, []byte(baseUrl))
	return baseUrl + "&token=" + token
}
