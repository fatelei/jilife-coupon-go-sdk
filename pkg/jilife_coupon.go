package jilife

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ReqType string

const (
	UserIDReqType ReqType = "USER_ID"
	CardNoReqType ReqType = "CARNO"
	TelReqType    ReqType = "TEL"
)

type JiLifeCoupon struct {
	appID      string
	appKey     string
	endpoint   string
	reqSource  string
	reqVersion string
}

type JiLifeCallback struct {
}

type ChannelItem struct {
	MiniAmt    float64 `json:"miniAmt"`
	PayChannel string  `json:"payChannel"`
}

type Coupon struct {
	PlanNo             string        `json:"planNo"`
	CouponsNo          string        `json:"couponsNo"`
	CouponsName        string        `json:"couponsName"`
	CouponsType        string        `json:"couponsType"`
	CouponsAmt         string        `json:"couponsAmt"`
	CouponsStatus      string        `json:"couponsStatus"`
	EffectiveStartTime string        `json:"effectiveStartTime"`
	EffectiveEndTime   string        `json:"effectiveEndTime"`
	CheckAmt           string        `json:"checkAmt"`
	CheckTime          string        `json:"checkTime"`
	MaxUseAmt          string        `json:"maxUseAmt"`
	ReceiveTime        string        `json:"receiveTime"`
	ReceiveMobile      string        `json:"receiveMobile"`
	ReceiveVehicleNo   string        `json:"receiveVehicleNo"`
	UseChannel         string        `json:"useChannel"`
	MinUseAmt          string        `json:"minUseAmt"`
	MinParkingHour     string        `json:"minPakingHour"`
	MaxParkingHour     string        `json:"maxPakingHour"`
	MinOrderAmt        string        `json:"minOrderAmt"`
	ParkList           []string      `json:"parkList"`
	PayChannel         []ChannelItem `json:"payChannel"`
}

type IssueCouponResponse struct {
	ResultCode string   `json:"resultCode"`
	Success    bool     `json:"success"`
	Message    string   `json:"message"`
	Obj        []Coupon `json:"obj,omitempty"`
}

func PKCSPadding(plaintext []byte, blockSize int) []byte {
	padding := blockSize - len(plaintext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(plaintext, padtext...)
}

func PKCSUnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func NewJiLifeCoupon(appID, appKey, endpoint, reqSource string) *JiLifeCoupon {
	return &JiLifeCoupon{
		appID:      appID,
		appKey:     appKey,
		endpoint:   endpoint,
		reqSource:  reqSource,
		reqVersion: "V1.0.0",
	}
}

func (p JiLifeCoupon) generateCommonParam() map[string]interface{} {
	return map[string]interface{}{
		"appId":      p.appID,
		"signType":   "MD5",
		"timestamp":  strconv.FormatInt(time.Now().UnixMilli(), 10),
		"nonce":      strings.Replace(uuid.New().String(), "-", "", -1),
		"reqVersion": p.reqVersion,
		"reqSource":  p.reqSource,
	}
}

func (p JiLifeCoupon) signParam(param map[string]interface{}) string {
	keys := make([]string, 0, len(param))
	for k := range param {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	queryAry := make([]string, 0)
	for _, key := range keys {
		value := param[key]
		switch value.(type) {
		case string:
			queryAry = append(queryAry, fmt.Sprintf("%s=%s", key, value.(string)))
		case []string:
			rawValue, _ := json.Marshal(value)
			queryAry = append(queryAry, fmt.Sprintf("%s=%s", key, string(rawValue)))
		}

	}
	queryAry = append(queryAry, p.appKey)
	value := strings.Join(queryAry, "&")
	println(value)
	hash := md5.Sum([]byte(value))
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

func (p JiLifeCoupon) sendRequest(ctx context.Context, path string, body []byte) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", p.endpoint, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	req.Header.Add("appId", p.appID)
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		return nil, err
	}
	var resp *http.Response
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	rawResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return rawResp, nil
}

func (p JiLifeCoupon) IssueCoupons(ctx context.Context, businessNo, reqData string, reqType ReqType, planNoList []string) (*IssueCouponResponse, error) {
	param := p.generateCommonParam()
	param["businessNo"] = businessNo
	param["reqType"] = string(reqType)
	param["reqData"] = reqData
	param["redeemCode"] = ""
	param["planNoList"] = planNoList
	sign := p.signParam(param)
	param["sign"] = sign
	if postRawBody, err := json.Marshal(&param); err == nil {
		rawResp, err := p.sendRequest(ctx, "api/coupon-server/issueCoupons", postRawBody)
		if err != nil {
			return nil, err
		}
		var couponResp IssueCouponResponse
		if err = json.Unmarshal(rawResp, &couponResp); err != nil {
			return nil, err
		}
		return &couponResp, nil
	} else {
		return nil, err
	}
}

func (p JiLifeCoupon) QueryCoupons(ctx context.Context, reqData string, reqType ReqType, startAt *string, endAt *string) (*IssueCouponResponse, error) {
	param := p.generateCommonParam()
	param["reqType"] = string(reqType)
	param["reqData"] = reqData
	if startAt != nil {
		param["receiveStartTime"] = *startAt
	}
	if endAt != nil {
		param["receiveEndTime"] = *endAt
	}

	sign := p.signParam(param)
	param["sign"] = sign
	if postRawBody, err := json.Marshal(&param); err == nil {
		rawResp, err := p.sendRequest(ctx, "api/coupon-server/queryCouponsList", postRawBody)
		if err != nil {
			return nil, err
		}
		var couponResp IssueCouponResponse
		if err = json.Unmarshal(rawResp, &couponResp); err != nil {
			return nil, err
		}
		return &couponResp, nil
	} else {
		return nil, err
	}
}
