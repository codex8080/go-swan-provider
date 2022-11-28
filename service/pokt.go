package service

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/filswan/go-swan-lib/client/swan"
	"github.com/filswan/go-swan-lib/client/web"
	"github.com/filswan/go-swan-lib/utils"
	"github.com/gin-gonic/gin"
	"math/big"
	"net/http"
	"strconv"
	"swan-provider/common"
	"swan-provider/config"
	"swan-provider/extern/docker"
	mydc "swan-provider/extern/docker"
	api "swan-provider/extern/pocket"
	"swan-provider/logs"
	"swan-provider/models"
	"time"
)

type Status struct {
	Version     string
	Address     string
	Height      string
	Balance     string
	Award       string
	Jailed      string
	JailedBlock string
	JailedUntil string
}

type PoktService struct {
	PoktApiUrl           string
	PoktAccessToken      string
	PoktScanInterval     time.Duration
	ApiHeartbeatInterval time.Duration
	PoktServerApiUrl     string
	PoktServerApiPort    int

	dkImage    string
	dkName     string
	dkConfPath string

	dkCli *mydc.DockerCli

	PoktAddress string
	AlarmBlc    big.Int
	CurStatus   Status
}

var myPoktSvr *PoktService

func GetMyPoktService() *PoktService {
	if myPoktSvr == nil {
		confPokt := config.GetConfig().Pokt

		myPoktSvr = &PoktService{
			PoktApiUrl:           confPokt.PoktApiUrl,
			PoktAccessToken:      confPokt.PoktAccessToken,
			PoktScanInterval:     confPokt.PoktScanInterval,
			ApiHeartbeatInterval: config.GetConfig().Main.SwanApiHeartbeatInterval,
			PoktServerApiUrl:     confPokt.PoktServerApiUrl,
			PoktServerApiPort:    confPokt.PoktServerApiPort,
			dkImage:              confPokt.PoktDockerImage,
			dkName:               confPokt.PoktDockerName,
			dkConfPath:           confPokt.PoktConfigPath,
			CurStatus:            Status{},
		}
		myPoktSvr.dkCli = docker.GetMyCli(myPoktSvr.dkImage, myPoktSvr.dkName, myPoktSvr.dkConfPath)

		logs.GetLog().Debugf("New myPoktSvr :%+v ", *myPoktSvr)

		return myPoktSvr
	}
	return myPoktSvr
}

func (psvc *PoktService) GetCli() *mydc.DockerCli {
	if psvc.dkCli == nil {
		psvc.dkCli = docker.GetMyCli(psvc.dkImage, psvc.dkName, psvc.dkConfPath)
		logs.GetLog().Infof("GetCli New Docker Cli")
	}
	return psvc.dkCli
}
func (psvc *PoktService) StartPoktContainer(op []string) {

	cli := psvc.dkCli
	if !cli.PoktCtnExist() {

		fs := flag.NewFlagSet("Start", flag.ExitOnError)
		passwd := fs.String("passwd", "", "password for create account")
		err := fs.Parse(op[1:])
		if *passwd == "" || err != nil {
			printPoktUsage()
			panic("need password for create account.")
			return
		}

		pass := *passwd
		logs.GetLog().Debug("POCKET_CORE_PASSPHRASE=", pass)
		env := []string{"POCKET_CORE_KEY=", "POCKET_CORE_PASSPHRASE=" + pass}

		accCmd := []string{"pocket", "accounts", "create"}
		cli.PoktCtnPullAndCreate(accCmd, env, true)
		cli.PoktCtnStart()

		for {
			if cli.PoktCtnExist() {
				logs.GetLog().Info("Wait for Creating Account...")
				cli.PoktCtnList()
				time.Sleep(time.Second * 3)
				continue
			}
			break
		}

		runCmd := []string{
			"pocket",
			"start",
			"--seeds=03b74fa3c68356bb40d58ecc10129479b159a145@seed1.mainnet.pokt.network:20656,64c91701ea98440bc3674fdb9a99311461cdfd6f@seed2.mainnet.pokt.network:21656,0057ee693f3ce332c4ffcb499ede024c586ae37b@seed3.mainnet.pokt.network:22856,9fd99b89947c6af57cd0269ad01ecb99960177cd@seed4.mainnet.pokt.network:23856,f2a4d0ec9d50ea61db18452d191687c899c3ca42@seed5.mainnet.pokt.network:24856,f2a9705924e8d0e11fed60484da2c3d22f7daba8@seed6.mainnet.pokt.network:25856,582177fd65dd03806eeaa2e21c9049e653672c7e@seed7.mainnet.pokt.network:26856,2ea0b13ab823986cfb44292add51ce8677b899ad@seed8.mainnet.pokt.network:27856,a5f4a4cd88db9fd5def1574a0bffef3c6f354a76@seed9.mainnet.pokt.network:28856,d4039bd71d48def9f9f61f670c098b8956e52a08@seed10.mainnet.pokt.network:29856,5c133f08ed297bb9e21e3e42d5f26e0f7d2b2832@poktseed100.chainflow.io:26656,361b1936d3fbe516628ebd6a503920fc4fc0f6a7@seed.pokt.rivet.cloud:26656",
			"--mainnet"}
		cli.PoktCtnCreateRun(runCmd, env, false)

	}

	cli.PoktCtnStart()
}

func (psvc *PoktService) StartScan() {
	url := psvc.PoktApiUrl
	height, err := api.PoktApiGetHeight(url)
	if err != nil {
		logs.GetLog().Error(err)
	}
	logs.GetLog().Info("Pokt Get Current Height=", height)

}

func (psvc *PoktService) SendPoktHeartbeatRequest(swanClient *swan.SwanClient) {

	params := ""
	confPokt := config.GetPoktConfig().Pokt
	selfUrl := utils.UrlJoin(confPokt.PoktServerApiUrl, API_POCKET_V1)

	apiUrl := utils.UrlJoin(selfUrl, "status")
	response, err := web.HttpGetNoToken(apiUrl, params)
	if err != nil {
		fmt.Printf("Heartbeat Get Pocket Status err: %s \n", err)
		return
	}

	res := &models.StatusResponse{}
	err = json.Unmarshal(response, res)
	if err != nil {
		fmt.Printf("Heartbeat Parse Response (%s) err: %s \n", response, err)
		return
	}
	// Swan Server Is Not Ready!
	//stat := swan.PocketHeartbeatOnlineParams{
	//	Address:     res.Data.Address,
	//	Version:     res.Data.Version,
	//	Height:      res.Data.Height,
	//	Balance:     res.Data.Balance,
	//	Award:       res.Data.Award,
	//	Jailed:      res.Data.Jailed,
	//	JailedBlock: res.Data.JailedBlock,
	//	JailedUntil: res.Data.JailedUntil,
	//}
	//err = swanClient.SendPoktHeartbeatRequest(stat)
	//if err != nil {
	//	fmt.Printf("Heartbeat Send err: %s \n", err)
	//	return
	//}

	{
		title := color.New(color.FgGreen).Sprintf("%s", "Version")
		value := color.New(color.FgYellow).Sprintf("%s", res.Data.Version)
		fmt.Printf("%s\t\t: %s\n", title, value)

		title = color.New(color.FgGreen).Sprintf("%s", "Height")
		value = color.New(color.FgYellow).Sprintf("%d", res.Data.Height)
		fmt.Printf("%s\t\t: %s\n", title, value)

		title = color.New(color.FgGreen).Sprintf("%s", "Address")
		value = color.New(color.FgYellow).Sprintf("%s", res.Data.Address)
		fmt.Printf("%s\t\t: %s\n", title, value)

		title = color.New(color.FgGreen).Sprintf("%s", "Balance")
		value = color.New(color.FgYellow).Sprintf("%d", res.Data.Balance)
		fmt.Printf("%s\t\t: %s\n", title, value)

		title = color.New(color.FgGreen).Sprintf("%s", "Jailed")
		value = color.New(color.FgRed).Sprintf("%t", res.Data.Jailed)
		fmt.Printf("%s\t\t: %s\n", title, value)

		title = color.New(color.FgGreen).Sprintf("%s", "JailedBlock")
		value = color.New(color.FgRed).Sprintf("%d", res.Data.JailedBlock)
		fmt.Printf("%s\t: %s\n", title, value)

		title = color.New(color.FgGreen).Sprintf("%s", "JailedUntil")
		value = color.New(color.FgRed).Sprintf("%s", res.Data.JailedUntil)
		fmt.Printf("%s\t: %s\n", title, value)
	}

	return
}

func HttpGetPoktVersion(c *gin.Context) {
	poktSvr := GetMyPoktService()
	cmdOut, err := poktSvr.GetCli().PoktCtnExecVersion()
	if err != nil {
		logs.GetLog().Error(err)
		c.JSON(http.StatusInternalServerError, common.CreateErrorResponse("-1", err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.CreateSuccessResponse(cmdOut))
}

func HttpGetPoktCurHeight(c *gin.Context) {
	poktSvr := GetMyPoktService()
	cmdOut, err := poktSvr.GetCli().PoktCtnExecHeight()
	if err != nil {
		logs.GetLog().Error(err)
		c.JSON(http.StatusInternalServerError, common.CreateErrorResponse("-1", err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.CreateSuccessResponse(cmdOut))
}

func HttpGetPoktNodeAddr(c *gin.Context) {
	poktSvr := GetMyPoktService()
	cmdOut, err := poktSvr.GetCli().PoktCtnExecNodeAddress()
	if err != nil {
		logs.GetLog().Error(err)
		c.JSON(http.StatusInternalServerError, common.CreateErrorResponse("-1", err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.CreateSuccessResponse(cmdOut))
}

func HttpGetPoktStatus(c *gin.Context) {
	poktSvr := GetMyPoktService()

	versionData, err := poktSvr.GetCli().PoktCtnExecVersion()
	if err != nil {
		logs.GetLog().Error(err)
		c.JSON(http.StatusInternalServerError, common.CreateErrorResponse("-1", err.Error()))
		return
	}

	heightData, err := poktSvr.GetCli().PoktCtnExecHeight()
	if err != nil {
		logs.GetLog().Error(err)
		c.JSON(http.StatusInternalServerError, common.CreateErrorResponse("-1", err.Error()))
		return
	}

	address, err := poktSvr.GetCli().PoktCtnExecNodeAddress()
	if err != nil {
		logs.GetLog().Error(err)
		c.JSON(http.StatusInternalServerError, common.CreateErrorResponse("-1", err.Error()))
		return
	}

	balanceData, err := poktSvr.GetCli().PoktCtnExecBalance(address)
	if err != nil {
		logs.GetLog().Error(err)
		c.JSON(http.StatusInternalServerError, common.CreateErrorResponse("-1", err.Error()))
		return
	}

	nodeData, err := poktSvr.GetCli().PoktCtnExecNode(address)
	if err != nil {
		logs.GetLog().Error(err)
		c.JSON(http.StatusInternalServerError, common.CreateErrorResponse("-1", err.Error()))
		return
	}

	signData, err := poktSvr.GetCli().PoktCtnExecSignInfo(address)
	if err != nil || len(signData) == 0 {
		logs.GetLog().Error(err)
		c.JSON(http.StatusInternalServerError, common.CreateErrorResponse("-1", err.Error()))
		return
	}
	signInfo := signData[0]

	data := &models.StatusData{
		Version:     versionData.Version,
		Height:      heightData.Height,
		Address:     address,
		Balance:     balanceData.Balance,
		Jailed:      nodeData.Jailed,
		JailedBlock: signInfo.JailedBlocksCounter,
		JailedUntil: signInfo.JailedUntil,
	}

	c.JSON(http.StatusOK, common.CreateSuccessResponse(data))
}

func HttpGetPoktBalance(c *gin.Context) {
	var params models.BalancePoktParams
	err := c.BindJSON(&params)
	if err != nil {
		logs.GetLog().Error(err)
		c.JSON(http.StatusOK, common.CreateErrorResponse("-1", err.Error()))
		return
	}

	poktSvr := GetMyPoktService()
	cmdOut, err := poktSvr.GetCli().PoktCtnExecBalance(params.Address)
	if err != nil {
		logs.GetLog().Error(err)
		c.JSON(http.StatusInternalServerError, common.CreateErrorResponse("-1", err.Error()))
		return
	}

	logs.GetLog().Debug("pocket query balance result:", cmdOut)

	data := &models.BalanceCmdData{
		Height:  params.Height,
		Address: params.Address,
		Balance: strconv.FormatUint(cmdOut.Balance, 10)}
	c.JSON(http.StatusOK, common.CreateSuccessResponse(data))
}

func HttpGetPoktThreshold(c *gin.Context) {
	var params models.ThresholdParams
	err := c.BindJSON(&params)
	if err != nil {
		logs.GetLog().Error(err)
		c.JSON(http.StatusOK, common.CreateErrorResponse("-1", err.Error()))
		return
	}

	data := &models.ThresholdData{
		Address:   params.Address,
		Threshold: params.Threshold,
		Active:    true,
	}
	c.JSON(http.StatusOK, common.CreateSuccessResponse(data))
}

///////////////////////////////////////////////////////////////////////////////

func HttpGetPoktCustodial(c *gin.Context) {
	var params models.CustodialParams
	err := c.BindJSON(&params)
	if err != nil {
		logs.GetLog().Error(err)
		c.JSON(http.StatusOK, common.CreateErrorResponse("-1", err.Error()))
		return
	}

	poktSvr := GetMyPoktService()
	result, err := poktSvr.GetCli().PoktCtnExecCustodial(
		params.Address,
		params.Amount,
		params.RelayChainIDs,
		params.ServiceURI,
		params.NetworkID,
		params.Fee,
		params.IsBefore,
	)
	if err != nil {
		logs.GetLog().Error(err)
		c.JSON(http.StatusInternalServerError, common.CreateErrorResponse("-1", err.Error()))
		return
	}

	data := &models.CustodialData{
		Result: result,
	}
	c.JSON(http.StatusOK, common.CreateSuccessResponse(data))
}

///////////////////////////////////////////////////////////////////////////////

func HttpGetPoktNonCustodial(c *gin.Context) {
	var params models.NonCustodialParams
	err := c.BindJSON(&params)
	if err != nil {
		logs.GetLog().Error(err)
		c.JSON(http.StatusOK, common.CreateErrorResponse("-1", err.Error()))
		return
	}

	poktSvr := GetMyPoktService()
	result, err := poktSvr.GetCli().PoktCtnExecNonCustodial(
		params.PubKey,
		params.OutputAddr,
		params.Amount,
		params.RelayChainIDs,
		params.ServiceURI,
		params.NetworkID,
		params.Fee,
		params.IsBefore,
	)
	if err != nil {
		logs.GetLog().Error(err)
		c.JSON(http.StatusInternalServerError, common.CreateErrorResponse("-1", err.Error()))
		return
	}

	data := &models.NonCustodialData{
		Result: result,
	}
	c.JSON(http.StatusOK, common.CreateSuccessResponse(data))
}
