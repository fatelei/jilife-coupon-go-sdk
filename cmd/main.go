package main

import (
	"context"
	"flag"
	"fmt"
	jilife_coupon "github.com/fatelei/jilife-coupon-go-sdk/pkg"
)

func main() {
	var appID string
	var appKey string
	var endpoint string
	var phone string
	var orderID string
	var planNo string

	flag.StringVar(&appID, "appID", "", "app id")
	flag.StringVar(&appKey, "appKey", "", "app key")
	flag.StringVar(&endpoint, "endpoint", "", "endpoint")
	flag.StringVar(&phone, "phone", "", "phone")
	flag.StringVar(&orderID, "orderID", "", "order id")
	flag.StringVar(&planNo, "planNo", "", "plan no")
	flag.Parse()

	if len(appID) > 0 && len(appKey) > 0 && len(endpoint) > 0 && len(phone) > 0 && len(orderID) > 0 && len(planNo) > 0 {
		ctl := jilife_coupon.NewJiLifeCoupon(appID, appKey, endpoint)
		resp, err := ctl.SendCouponViaPhone(context.Background(), phone, planNo, orderID)
		if err != nil {
			panic(err)
		}
		fmt.Printf("resp %+v", resp)
	}
}
