package jilife_coupon

import "testing"

func TestGenerateSign(t *testing.T) {
	ctl := NewJiLifeCoupon("123456", "123456789", "https://foo.com")
	param := ctl.generateCommonParam()
	sign := ctl.signParam(param)
	println(sign)
}
