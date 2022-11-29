package pocket

import (
	"encoding/json"
	"github.com/filswan/go-swan-lib/client/web"
	"github.com/filswan/go-swan-lib/utils"
	"swan-provider/logs"
	"swan-provider/models"
)

func PoktApiGetVersion(url string) (string, error) {
	params := ""
	apiUrl := utils.UrlJoin(url, "")
	response, err := web.HttpGetNoToken(apiUrl, params)
	if err != nil {
		logs.GetLog().Error(err)
		return "", err
	}

	getVersionResponse := ""
	err = json.Unmarshal([]byte(response), &getVersionResponse)
	if err != nil {
		logs.GetLog().Error(err)
		return "", err
	}

	return getVersionResponse, nil
}

func PoktApiGetHeight(url string) (uint64, error) {
	params := &models.HeightPoktParams{}
	apiUrl := utils.UrlJoin(url, "/query/height")
	response, err := web.HttpPostNoToken(apiUrl, params)
	if err != nil {
		logs.GetLog().Error(err)
		return 0, err
	}

	poktRes := models.HeightData{}
	err = json.Unmarshal([]byte(response), &poktRes)
	if err != nil {
		logs.GetLog().Error(err)
		return 0, err
	}

	return poktRes.Height, nil
}

func PoktApiGetSupply(url string) (*models.PoktSupplyResponse, error) {
	params := &models.PoktSupplyParams{
		Height: 0,
	}
	apiUrl := utils.UrlJoin(url, "/query/supply")
	response, err := web.HttpPostNoToken(apiUrl, params)
	if err != nil {
		logs.GetLog().Error(err)
		return nil, err
	}

	poktRes := &models.PoktSupplyResponse{}
	err = json.Unmarshal([]byte(response), &poktRes)
	if err != nil {
		logs.GetLog().Error(err)
		return nil, err
	}

	return poktRes, nil
}

func PoktApiGetBalance(url string, height uint64, addr string) (*models.BalanceData, error) {
	params := &models.BalancePoktParams{
		Height:  height,
		Address: addr,
	}

	apiUrl := utils.UrlJoin(url, "/query/balance")
	response, err := web.HttpPostNoToken(apiUrl, params)
	if err != nil {
		logs.GetLog().Error(err)
		return nil, err
	}

	poktRes := &models.BalanceData{}
	err = json.Unmarshal([]byte(response), &poktRes)
	if err != nil {
		logs.GetLog().Error(err)
		return nil, err
	}

	return poktRes, nil
}

func PoktApiGetNode(url string, addr string) (*models.PoktNodeResponse, error) {
	params := &models.PoktNodeParams{
		Height:  0,
		Address: addr,
	}

	apiUrl := utils.UrlJoin(url, "/query/node")
	response, err := web.HttpPostNoToken(apiUrl, params)
	if err != nil {
		logs.GetLog().Error(err)
		return nil, err
	}

	poktRes := &models.PoktNodeResponse{}
	err = json.Unmarshal([]byte(response), &poktRes)
	if err != nil {
		logs.GetLog().Error(err)
		return nil, err
	}

	return poktRes, nil
}

func PoktApiGetSigningInfo(url string, addr string) (*models.PoktSigningInfoResponse, error) {
	params := &models.PoktSignInfoParams{
		Height:  0,
		Address: addr,
		Page:    0,
		PerPage: 100,
	}

	apiUrl := utils.UrlJoin(url, "/query/signinginfo")
	response, err := web.HttpPostNoToken(apiUrl, params)
	if err != nil {
		logs.GetLog().Error(err)
		return nil, err
	}

	poktRes := &models.PoktSigningInfoResponse{}
	err = json.Unmarshal([]byte(response), &poktRes)
	if err != nil {
		logs.GetLog().Error(err)
		return nil, err
	}

	return poktRes, nil
}
