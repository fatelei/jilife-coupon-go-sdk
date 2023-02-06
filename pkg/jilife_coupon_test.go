package jilife

import "testing"

func TestGenerateSign(t *testing.T) {
	ctl := NewJiLifeCoupon("123456", "123456789", "https://foo.com", "Test")
	param := ctl.generateCommonParam()
	sign := ctl.signParam(param)
	if sign == "E7D99526FA02FF68F87533DA0E65A768" {
		t.Fatal("sign is not equal to E7D99526FA02FF68F87533DA0E65A768")
	}
}
