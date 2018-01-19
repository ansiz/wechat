package template

import (
	"encoding/json"
	"fmt"

	"kshare/webserver/modules/wechat/context"
	"kshare/webserver/modules/wechat/util"
)

const (
	templateSendURL = "https://api.weixin.qq.com/cgi-bin/message/template/send"
	templateListURL = "https://api.weixin.qq.com/cgi-bin/template/get_all_private_template"
)

// Manager struct extends context.
type Manager struct {
	*context.Context
}

// NewManager returns a template manager
func NewManager(context *context.Context) *Manager {
	return &Manager{context}
}

//Message 发送的模板消息内容
type Message struct {
	ToUser     string               `json:"touser"`          // 必须, 接受者OpenID
	TemplateID string               `json:"template_id"`     // 必须, 模版ID
	URL        string               `json:"url,omitempty"`   // 可选, 用户点击后跳转的URL, 该URL必须处于开发者在公众平台网站中设置的域中
	Color      string               `json:"color,omitempty"` // 可选, 整个消息的颜色, 可以不设置
	Data       map[string]*DataItem `json:"data"`            // 必须, 模板数据

	MiniProgram struct {
		AppID    string `json:"appid"`    //所需跳转到的小程序appid（该小程序appid必须与发模板消息的公众号是绑定关联关系）
		PagePath string `json:"pagepath"` //所需跳转到小程序的具体页面路径，支持带参数,（示例index?foo=bar）
	} `json:"miniprogram"` //可选,跳转至小程序地址
}

//DataItem 模版内某个 .DATA 的值
type DataItem struct {
	Value string `json:"value"`
	Color string `json:"color,omitempty"`
}

type resTemplateSend struct {
	util.CommonError

	MsgID int64 `json:"msgid"`
}

type resTemplateList struct {
	util.CommonError
	Templates interface{} `json:"template_list"`
}

//Send 发送模板消息
func (m *Manager) Send(msg *Message) (msgID int64, err error) {
	var accessToken string
	accessToken, err = m.GetAccessToken()
	if err != nil {
		return
	}
	uri := fmt.Sprintf("%s?access_token=%s", templateSendURL, accessToken)
	response, err := util.PostJSON(uri, msg)

	var result resTemplateSend
	err = json.Unmarshal(response, &result)
	if err != nil {
		return
	}
	if result.ErrCode != 0 {
		err = fmt.Errorf("template msg send error : errcode=%v , errmsg=%v", result.ErrCode, result.ErrMsg)
		return
	}
	msgID = result.MsgID
	return
}

// ListTemplates lists all tempaltes
func (m *Manager) ListTemplates() (interface{}, error) {
	accessToken, err := m.GetAccessToken()
	if err != nil {
		return nil, err
	}
	uri := fmt.Sprintf("%s?access_token=%s", templateListURL, accessToken)
	response, err := util.HTTPGet(uri)

	var result resTemplateList
	err = json.Unmarshal(response, &result)
	if err != nil {
		return nil, err
	}
	if result.ErrCode != 0 {
		err = fmt.Errorf("template msg list error : errcode=%v , errmsg=%v", result.ErrCode, result.ErrMsg)
		return nil, err
	}
	return result, nil
}
