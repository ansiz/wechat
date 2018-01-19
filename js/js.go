package js

import (
	"encoding/json"
	"fmt"
	"time"

	"kshare/webserver/modules/wechat/context"
	"kshare/webserver/modules/wechat/util"
)

const getTicketURL = "https://api.weixin.qq.com/cgi-bin/ticket/getticket?access_token=%s&type=jsapi"

// Manager struct
type Manager struct {
	*context.Context
}

// Config 返回给用户jssdk配置信息
type Config struct {
	AppID     string `json:"app_id"`
	Timestamp int64  `json:"timestamp"`
	NonceStr  string `json:"nonce_str"`
	Signature string `json:"signature"`
}

// resTicket 请求jsapi_tikcet返回结果
type resTicket struct {
	util.CommonError

	Ticket    string `json:"ticket"`
	ExpiresIn int64  `json:"expires_in"`
}

// NewManager returns a jssdk manager.
func NewManager(context *context.Context) *Manager {
	return &Manager{Context: context}
}

//GetConfig 获取jssdk需要的配置参数
//uri 为当前网页地址
func (m *Manager) GetConfig(uri string) (config *Config, err error) {
	config = new(Config)
	var ticketStr string
	ticketStr, err = m.GetTicket()
	if err != nil {
		return
	}

	nonceStr := util.RandomStr(16)
	timestamp := util.GetCurrTs()
	str := fmt.Sprintf("jsapi_ticket=%s&noncestr=%s&timestamp=%d&url=%s", ticketStr, nonceStr, timestamp, uri)
	sigStr := util.Signature(str)

	config.AppID = m.AppID
	config.NonceStr = nonceStr
	config.Timestamp = timestamp
	config.Signature = sigStr
	return
}

//GetTicket 获取jsapi_tocket
func (m *Manager) GetTicket() (ticketStr string, err error) {
	m.GetJsAPITicketLock().Lock()
	defer m.GetJsAPITicketLock().Unlock()

	//先从cache中取
	jsAPITicketCacheKey := fmt.Sprintf("jsapi_ticket_%s", m.AppID)
	val := m.Cache.Get(jsAPITicketCacheKey)
	if val != nil {
		ticketStr = val.(string)
		return
	}
	var ticket resTicket
	ticket, err = m.getTicketFromServer()
	if err != nil {
		return
	}
	ticketStr = ticket.Ticket
	return
}

//getTicketFromServer 强制从服务器中获取ticket
func (m *Manager) getTicketFromServer() (ticket resTicket, err error) {
	var accessToken string
	accessToken, err = m.GetAccessToken()
	if err != nil {
		return
	}

	var response []byte
	url := fmt.Sprintf(getTicketURL, accessToken)
	response, err = util.HTTPGet(url)
	err = json.Unmarshal(response, &ticket)
	if err != nil {
		return
	}
	if ticket.ErrCode != 0 {
		err = fmt.Errorf("getTicket Error : errcode=%d , errmsg=%s", ticket.ErrCode, ticket.ErrMsg)
		return
	}

	jsAPITicketCacheKey := fmt.Sprintf("jsapi_ticket_%s", m.AppID)
	expires := ticket.ExpiresIn - 1500
	err = m.Cache.Set(jsAPITicketCacheKey, ticket.Ticket, time.Duration(expires)*time.Second)
	return
}
