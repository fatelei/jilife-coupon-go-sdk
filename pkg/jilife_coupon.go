package jilife_coupon

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type JiLifeCoupon struct {
	appID    string
	appKey   string
	endpoint string
}

type JiLifeCallback struct {
}

type Coupon struct {
	CouponNo    string  `json:"couponNo"`
	CouponValue float64 `json:"couponValue"`
}

type JiLifeIssueByPhoneResponse struct {
	ResultCode string `json:"resultCode"`
	Message    string `json:"message"`
	Obj        Coupon `json:"obj"`
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

func NewJiLifeCoupon(appID, appKey, endpoint string) *JiLifeCoupon {
	return &JiLifeCoupon{
		appID:    appID,
		appKey:   appKey,
		endpoint: endpoint,
	}
}

func (p JiLifeCoupon) generateCommonParam() map[string]string {
	return map[string]string{
		"appId":     p.appID,
		"signType":  "MD5",
		"timestamp": strconv.FormatInt(time.Now().Unix(), 10),
		"nonce":     strings.Replace(uuid.New().String(), "-", "", -1),
	}
}

func AesDecryptCBC(crypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCSUnPadding(origData)
	return origData, nil
}

func (p JiLifeCoupon) signParam(param map[string]string) string {
	keys := make([]string, 0, len(param))
	for k := range param {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	queryParam := url.Values{}
	for _, key := range keys {
		queryParam.Add(key, param[key])
	}
	value := fmt.Sprintf("%s&%s", queryParam.Encode(), p.appKey)
	hash := md5.Sum([]byte(value))
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

func (p JiLifeCoupon) encryptBody(body map[string]string) (string, error) {
	key := []byte(p.appKey)
	cipherBlock, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	var src []byte
	if src, err = json.Marshal(body); err == nil {
		blockSize := cipherBlock.BlockSize()
		src = PKCSPadding(src, blockSize)
		blockMode := cipher.NewCBCEncrypter(cipherBlock, key[:blockSize])
		crypted := make([]byte, len(src))
		blockMode.CryptBlocks(crypted, src)
		return hex.EncodeToString(crypted), nil
	} else {
		return "", err
	}
}

func (p JiLifeCoupon) SendCouponViaPhone(ctx context.Context, phone, planNo, outOrderNo string) (*JiLifeIssueByPhoneResponse, error) {
	param := p.generateCommonParam()
	param["telephone"] = phone
	param["outOrderNo"] = outOrderNo
	param["planNo"] = planNo
	sign := p.signParam(param)
	param["sign"] = sign
	body, err := p.encryptBody(param)
	if err != nil {
		return nil, err
	}
	postData := map[string]string{
		"encryptedData": body,
	}
	var postRawBody []byte
	if postRawBody, err = json.Marshal(&postData); err == nil {
		url := fmt.Sprintf("%s/%s", p.endpoint, "api/coupon-server/issueByPhone")
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(postRawBody))
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
		var couponResp JiLifeIssueByPhoneResponse
		if err = json.Unmarshal(rawResp, &couponResp); err != nil {
			return nil, err
		}
		return &couponResp, nil
	}
	return nil, err
}
