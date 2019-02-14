package mailtrack

import (
	"time"

	"github.com/patrickmn/go-cache"
)

const (
	// CTranslatorKey 取翻译实例的键
	CTranslatorKey = "c_tras_k"
	// LZhHansCN 大陆普通话
	LZhHansCN = "zh_CN"
	// LEn 英语
	LEn = "en"
)

// WebConfig 网站配置
type WebConfig struct {
	Domain string
}

// WC 网站配置
var WC WebConfig

// Cache 全局缓存工具
var Cache = cache.New(5*time.Minute, 10*time.Minute)

func init() {
	WC.Domain = "localhost:8080"
}
