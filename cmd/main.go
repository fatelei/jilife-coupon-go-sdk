package main

import (
	"context"
	"flag"
	"fmt"
	jilife "github.com/fatelei/jilife-coupon-go-sdk/pkg"
	"time"
)

func main() {
	var appID string
	var appKey string
	var endpoint string
	var phone string
	var orderID string
	var planNo string
	var cmd string

	flag.StringVar(&cmd, "cmd", "issueCoupon", "cmd")
	flag.StringVar(&appID, "appID", "", "app id")
	flag.StringVar(&appKey, "appKey", "", "app key")
	flag.StringVar(&endpoint, "endpoint", "", "endpoint")
	flag.StringVar(&phone, "phone", "", "phone")
	flag.StringVar(&orderID, "orderID", "", "order id")
	flag.StringVar(&planNo, "planNo", "", "plan no") // CPP230202000009535
	flag.Parse()

	ctl := jilife.NewJiLifeCoupon(appID, appKey, endpoint, "red_envelope", time.Second*10)
	if cmd == "issueCoupon" {
		if len(appID) > 0 && len(appKey) > 0 && len(endpoint) > 0 && len(phone) > 0 && len(orderID) > 0 && len(planNo) > 0 {
			resp, err := ctl.IssueCoupons(context.Background(), orderID, phone, jilife.TelReqType, []string{planNo})
			if err != nil {
				panic(err)
			}
			fmt.Printf("resp %+v", resp)
		}
	} else if cmd == "queryCoupon" {
		resp, err := ctl.QueryCoupons(context.Background(), phone, jilife.TelReqType, nil, nil)
		if err != nil {
			panic(err)
		}
		fmt.Printf("resp %+v", resp)
	}
}
