package wechat

import (
	"net/http"
	"sync"

	"kshare/webserver/modules/wechat/cache"
	"kshare/webserver/modules/wechat/context"
	"kshare/webserver/modules/wechat/js"
	"kshare/webserver/modules/wechat/material"
	"kshare/webserver/modules/wechat/menu"
	"kshare/webserver/modules/wechat/oauth"
	"kshare/webserver/modules/wechat/pay"
	"kshare/webserver/modules/wechat/server"
	"kshare/webserver/modules/wechat/template"
	"kshare/webserver/modules/wechat/user"
)

// Wechat struct
type Wechat struct {
	Context  *context.Context
	OAuth    *oauth.Manager
	Pay      *pay.Manager
	JSSDK    *js.Manager
	Material *material.Manager
	Template *template.Manager
}

// Config for user
type Config struct {
	AppID          string
	AppSecret      string
	Token          string
	EncodingAESKey string
	PayMchID       string
	PayNotifyURL   string
	PayKey         string
	Cache          cache.Cache
}

// NewWechat init
func NewWechat(cfg *Config) *Wechat {
	context := new(context.Context)
	copyConfigToContext(cfg, context)
	return &Wechat{
		Context:  context,
		OAuth:    oauth.NewManager(context),
		Pay:      pay.NewManager(context),
		JSSDK:    js.NewManager(context),
		Material: material.NewManager(context),
		Template: template.NewManager(context),
	}
}

func copyConfigToContext(cfg *Config, context *context.Context) {
	context.AppID = cfg.AppID
	context.AppSecret = cfg.AppSecret
	context.Token = cfg.Token
	context.EncodingAESKey = cfg.EncodingAESKey
	context.PayMchID = cfg.PayMchID
	context.PayNotifyURL = cfg.PayNotifyURL
	context.PayKey = cfg.PayKey
	context.Cache = cfg.Cache
	context.SetAccessTokenLock(new(sync.RWMutex))
	context.SetJsAPITicketLock(new(sync.RWMutex))
}

// GetServer 消息管理
func (wc *Wechat) GetServer(req *http.Request, writer http.ResponseWriter) *server.Server {
	wc.Context.Request = req
	wc.Context.Writer = writer
	return server.NewServer(wc.Context)
}

//GetAccessToken 获取access_token
func (wc *Wechat) GetAccessToken() (string, error) {
	return wc.Context.GetAccessToken()
}

// GetMenu 菜单管理接口
func (wc *Wechat) GetMenu() *menu.Menu {
	return menu.NewMenu(wc.Context)
}

// GetUser 用户管理接口
func (wc *Wechat) GetUser() *user.User {
	return user.NewUser(wc.Context)
}
