package pocket

import (
	"encoding/json"
	"github.com/filswan/go-swan-lib/client/web"
	"github.com/filswan/go-swan-lib/utils"
	"swan-provider/logs"
	"swan-provider/models"
)

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
