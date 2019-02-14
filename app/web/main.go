package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh_Hans_CN"
	"github.com/go-playground/universal-translator"
	"github.com/naiba/mailtrack"
)

func main() {
	zh := zh_Hans_CN.New()
	utrans := ut.New(zh, zh, en.New())
	err := utrans.Import(ut.FormatJSON, "resource/translations")
	if err != nil {
		log.Fatal(err)
	}
	err = utrans.VerifyTranslations()
	if err != nil {
		log.Fatal(err)
	}

	r := gin.Default()
	r.Static("/static", "resource/static")
	r.LoadHTMLGlob("resource/template/**/*")
	r.Use(func(c *gin.Context) {
		lang := c.Query("locale")
		if !isValidatorLanguage(lang) {
			lang, _ = c.Cookie("locale")
			if !isValidatorLanguage(lang) {
				if strings.Contains(c.Request.Header.Get("Accept-Language"), "zh") {
					lang = mailtrack.LZhHansCN
				} else {
					lang = mailtrack.LEn
				}
			}
		} else {
			c.SetCookie("locale", lang, 60*60*24*365*2, "/", "", false, false)
		}
		tr, _ := utrans.GetTranslator(lang)
		c.Set(mailtrack.CTranslatorKey, tr)
	})
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"T": c.MustGet(mailtrack.CTranslatorKey).(ut.Translator),
		})
	})
	r.Run()
}

func isValidatorLanguage(lang string) bool {
	return len(lang) > 0 && (lang == "en" || lang == "zh_Hans_CN")
}
