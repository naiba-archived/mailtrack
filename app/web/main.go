package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/patrickmn/go-cache"

	"github.com/coreos/go-oidc"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh_Hans_CN"
	"github.com/go-playground/universal-translator"
	"github.com/naiba/com"
	"github.com/naiba/mailtrack"
	"golang.org/x/oauth2"
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
	// SetLanguage
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
	// Index
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"T": c.MustGet(mailtrack.CTranslatorKey).(ut.Translator),
		})
	})
	r.GET("/oauth2/login", Oauth2Login)
	r.GET("/oauth2/callback", Oauth2LoginCallback)
	r.Run()
}

func isValidatorLanguage(lang string) bool {
	return len(lang) > 0 && (lang == "en" || lang == "zh_Hans_CN")
}

var oidcConfig *oidc.Config
var config oauth2.Config
var ctx context.Context
var verifier *oidc.IDTokenVerifier

func init() {
	ctx = context.Background()
	clientID := "1-gBtUEy"
	clientSecret := "eDib2wZci1MgNlXl"

	provider, err := oidc.NewProvider(ctx, "https://space.mentuo.com")
	if err != nil {
		log.Fatal(err)
	}
	oidcConfig = &oidc.Config{
		ClientID: clientID,
	}
	verifier = provider.Verifier(oidcConfig)
	config = oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  "http://" + mailtrack.WC.Domain + "/oauth2/callback",
		Scopes:       []string{oidc.ScopeOpenID, "profile"},
	}
}

// Oauth2Login 烧饼社群登录
func Oauth2Login(c *gin.Context) {
	state := com.RandomString(7)
	mailtrack.Cache.Set("o2ls"+state, true, cache.DefaultExpiration)
	c.Redirect(http.StatusFound, config.AuthCodeURL(state))
}

// Oauth2LoginCallback 烧饼社群登录回调
func Oauth2LoginCallback(c *gin.Context) {
	state, ok := mailtrack.Cache.Get("o2ls" + c.Query("state"))
	if !ok || !state.(bool) {
		c.String(http.StatusForbidden, "Failed to verify state")
		return
	}
	oauth2Token, err := config.Exchange(ctx, c.Query("code"))
	if err != nil {
		c.String(http.StatusBadRequest, "Failed to exchange token: "+err.Error())
		return
	}
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		c.String(http.StatusInternalServerError, "No id_token field in oauth2 token.")
		return
	}
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to verify ID Token: "+err.Error())
		return
	}

	resp := struct {
		OAuth2Token   *oauth2.Token
		IDTokenClaims *json.RawMessage // ID Token payload is just JSON.
	}{oauth2Token, new(json.RawMessage)}

	if err := idToken.Claims(&resp.IDTokenClaims); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	data, err := json.MarshalIndent(resp, "", "    ")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	var x map[string]map[string]interface{}
	err = json.Unmarshal(data, &x)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	oid := x["IDTokenClaims"]["sub"]
	username := x["IDTokenClaims"]["name"]
	log.Println(oid, username)
}
