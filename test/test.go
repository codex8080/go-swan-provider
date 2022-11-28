package test

import (
	"swan-provider/service"

	"swan-provider/logs"
)

func Test() {
	TestFindNextDealReady2Download()
}

func TestFindNextDealReady2Download() {
	//aria2Client := service.SetAndCheckAria2Config()
	swanClient := service.SetAndCheckSwanConfig()
	aria2Service := service.GetAria2Service()
	offlineDeal := aria2Service.FindNextDealReady2Download(swanClient)
	if offlineDeal != nil {
		logs.GetLogger().Info(*offlineDeal)
	}
}
