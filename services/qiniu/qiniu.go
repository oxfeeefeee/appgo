package qiniu

import (
	qndigest "github.com/qiniu/api/auth/digest"
	"github.com/qiniu/api/conf"
	qnio "github.com/qiniu/api/io"
	"github.com/qiniu/api/rs"
	qnurl "github.com/qiniu/api/url"
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

func Init(params *QiniuParams) {
	conf.ACCESS_KEY = params.AccessKey
	conf.SECRET_KEY = params.Secret
	bucketName = params.Bucket
	domain = params.Domain
	putPolicy = rs.PutPolicy{
		Expires: uint32(params.Domain),
	}
	getPolicy = rs.GetPolicy{
		Expires: uint32(params.Domain),
	}
}

func GetUrl(key string, expiry uid.Uid, params string) string {
	baseUrl := makeBaseUrl(key)
	url := baseUrl + "?" + params
	if expiry == 0 {
		return getPolicy.MakeRequest(url, nil)
	}
	return makeRequest(url, expiry, nil)
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

func makeRequest(baseUrl string, expiry uid.Uid, mac *qndigest.Mac) (privateUrl string) {
	deadline := expiry.Time().Unix()

	if strings.Contains(baseUrl, "?") {
		baseUrl += "&e="
	} else {
		baseUrl += "?e="
	}
	baseUrl += strconv.FormatInt(deadline, 10)

	token := qndigest.Sign(mac, []byte(baseUrl))
	return baseUrl + "&token=" + token
}
