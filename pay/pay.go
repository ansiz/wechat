package pay

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"hash"
	"sort"
	"time"

	"github.com/astaxie/beego/logs"

	"kshare/webserver/modules/wechat/context"
	"kshare/webserver/modules/wechat/util"
)

const (
	unifiedOrderURL = "https://api.mch.weixin.qq.com/pay/unifiedorder"
	dftTradeType    = "JSAPI"
	tplQueryString  = "appid=%s&body=%s&mch_id=%s&nonce_str=%s&notify_url=%s&openid=%s&out_trade_no=%s&spbill_create_ip=%s&total_fee=%s&trade_type=%s&key=%s"
)

// Manager struct extends context.
type Manager struct {
	*context.Context
}

// PayParams represents the extra parameters for unifiedorder.
type PayParams struct {
	TotalFee   string
	CreateIP   string `json:"-"`
	Body       string
	OutTradeNo string
	OpenID     string `json:"-"`
}

// JSAPIParams returns the parameters for JSAPI payment.
type JSAPIParams struct {
	AppID     string
	Timestamp int64
	NonceStr  string
	PrePayID  string
	SignType  string
	Sign      string
}

// UnifiedOrderResp represents the unified order response.
type UnifiedOrderResp struct {
	CommonResp
	AppID      string `xml:"appid,omitempty"`
	MchID      string `xml:"mch_id,omitempty"`
	NonceStr   string `xml:"nonce_str,omitempty"`
	Sign       string `xml:"sign,omitempty"`
	ResultCode string `xml:"result_code,omitempty"`
	TradeType  string `xml:"trade_type,omitempty"`
	PrePayID   string `xml:"prepay_id,omitempty"`
	CodeURL    string `xml:"code_url,omitempty"`
	ErrCode    string `xml:"err_code,omitempty"`
	ErrCodeDes string `xml:"err_code_des,omitempty"`
}

// UnifiedOrderRequest represents the unified order request parameters.
type UnifiedOrderRequest struct {
	AppID          string `xml:"appid"`
	MchID          string `xml:"mch_id"`
	DeviceInfo     string `xml:"device_info,omitempty"`
	NonceStr       string `xml:"nonce_str"`
	Sign           string `xml:"sign"`
	SignType       string `xml:"sign_type,omitempty"`
	Body           string `xml:"body"`
	Detail         string `xml:"detail,omitempty"`
	Attach         string `xml:"attach,omitempty"`      //附加数据
	OutTradeNo     string `xml:"out_trade_no"`          //商户订单号
	FeeType        string `xml:"fee_type,omitempty"`    //标价币种
	TotalFee       string `xml:"total_fee"`             //标价金额
	SpbillCreateIP string `xml:"spbill_create_ip"`      //终端IP
	TimeStart      string `xml:"time_start,omitempty"`  //交易起始时间
	TimeExpire     string `xml:"time_expire,omitempty"` //交易结束时间
	GoodsTag       string `xml:"goods_tag,omitempty"`   //订单优惠标记
	NotifyURL      string `xml:"notify_url"`            //通知地址
	TradeType      string `xml:"trade_type"`            //交易类型
	ProductID      string `xml:"product_id,omitempty"`  //商品ID
	LimitPay       string `xml:"limit_pay,omitempty"`   //是否允许信用卡支付
	OpenID         string `xml:"openid,omitempty"`      //用户标识
	SceneInfo      string `xml:"scene_info,omitempty"`  //场景信息
}

// CommonResp represents the common response.
type CommonResp struct {
	ReturnCode string `xml:"return_code"`
	ReturnMsg  string `xml:"return_msg"`
}

// Notify represents the WeChat payment notification.
type Notify struct {
	CommonResp
	AppID              string `xml:"appid"`
	MchID              string `xml:"mch_id"`
	DeviceInfo         string `xml:"device_info"`
	NonceStr           string `xml:"nonce_str"`
	Sign               string `xml:"sign"`
	SignType           string `xml:"sign_type"`
	ResultCode         string `xml:"result_code"`
	ErrCode            string `xml:"err_code"`
	ErrCodeDes         string `xml:"err_code_des"`
	OpenID             string `xml:"openid"`
	IsSubscribe        string `xml:"is_subscribe"`
	TradeType          string `xml:"trade_type"`
	BankType           string `xml:"bank_type"`
	TotalFee           int    `xml:"total_fee"`
	SettlementTotalFee int    `xml:"settlement_total_fee"`
	FeeType            string `xml:"fee_type"`
	CashFee            int    `xml:"cash_fee"`
	CashFeeType        string `xml:"cash_fee_type"`
	CouponFee          int    `xml:"coupon_fee"`
	CouponCount        int    `xml:"coupon_count"`
	TransactionID      string `xml:"transaction_id"`
	OutTradeNo         string `xml:"out_trade_no"`
	Attach             string `xml:"attach"`
	TimeEnd            string `xml:"time_end"`
}

// NewManager returns an instance of pay package.
func NewManager(ctx *context.Context) *Manager {
	return &Manager{ctx}
}

// getPrepayResponse sends the unified order request, parses and returns the
// response.
func (m *Manager) getPrepayResponse(params *PayParams) (*UnifiedOrderResp, error) {
	nonceStr := util.RandomStr(32)
	str := fmt.Sprintf(tplQueryString, m.AppID, params.Body, m.PayMchID,
		nonceStr, m.PayNotifyURL, params.OpenID, params.OutTradeNo, params.CreateIP,
		params.TotalFee, dftTradeType, m.PayKey)
	sign := util.MD5Sum(str)
	request := UnifiedOrderRequest{
		AppID:          m.AppID,
		MchID:          m.PayMchID,
		NonceStr:       nonceStr,
		Sign:           sign,
		Body:           params.Body,
		OutTradeNo:     params.OutTradeNo,
		TotalFee:       params.TotalFee,
		SpbillCreateIP: params.CreateIP,
		NotifyURL:      m.PayNotifyURL,
		TradeType:      dftTradeType,
		OpenID:         params.OpenID,
	}
	data, err := util.PostXML(unifiedOrderURL, request)
	if err != nil {
		return nil, fmt.Errorf("send unified order request failed: %v", err)
	}
	res := UnifiedOrderResp{}
	err = xml.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	logs.Debug("WeChat unified order response: %s", data)
	if res.ReturnCode == "SUCCESS" {
		if res.ResultCode == "SUCCESS" {
			return &res, nil
		}
	}
	return nil, fmt.Errorf("get prepay id failed: errcode=%s, codedesc=%s errmsg=%s", res.ErrCode, res.ErrCodeDes, res.ReturnMsg)
}

// GenJSAPIParams generates the JSAPI required parameters.
func (m *Manager) GenJSAPIParams(params *PayParams) (*JSAPIParams, error) {
	order, err := m.getPrepayResponse(params)
	if err != nil {
		return nil, err
	}
	ts := time.Now().Unix()
	signParams := map[string]string{
		"appId":     order.AppID,
		"timeStamp": fmt.Sprint(ts),
		"nonceStr":  order.NonceStr,
		"package":   fmt.Sprintf("prepay_id=%s", order.PrePayID),
		"signType":  "MD5",
	}
	resp := &JSAPIParams{
		AppID:     order.AppID,
		Timestamp: ts,
		NonceStr:  order.NonceStr,
		PrePayID:  order.PrePayID,
		SignType:  "MD5",
		Sign:      JsPaySign(signParams, m.PayKey, nil),
	}
	return resp, nil
}

// JsPaySign generates parameters signature.
func JsPaySign(parameters map[string]string, apiKey string, fn func() hash.Hash) string {
	ks := make([]string, 0, len(parameters))
	for k := range parameters {
		if k == "sign" {
			continue
		}
		ks = append(ks, k)
	}
	sort.Strings(ks)

	if fn == nil {
		fn = md5.New
	}
	h := fn()

	buf := make([]byte, 256)
	for _, k := range ks {
		v := parameters[k]
		if v == "" {
			continue
		}

		buf = buf[:0]
		buf = append(buf, k...)
		buf = append(buf, '=')
		buf = append(buf, v...)
		buf = append(buf, '&')
		h.Write(buf)
	}
	buf = buf[:0]
	buf = append(buf, "key="...)
	buf = append(buf, apiKey...)
	h.Write(buf)

	signature := make([]byte, h.Size()*2)
	hex.Encode(signature, h.Sum(nil))
	return string(bytes.ToUpper(signature))
}
