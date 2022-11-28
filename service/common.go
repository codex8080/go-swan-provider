package service

import (
	"flag"
	"fmt"
	"github.com/filswan/go-swan-lib/client/web"
	"github.com/gin-gonic/gin"
	cors "github.com/itsjamie/gin-cors"
	"strconv"
	"strings"
	"swan-provider/common/constants"
	"swan-provider/config"
	"swan-provider/logs"
	"swan-provider/models"
	"time"

	"github.com/filswan/go-swan-lib/client"
	libconstants "github.com/filswan/go-swan-lib/constants"
	libmodel "github.com/filswan/go-swan-lib/model"
	"github.com/filswan/go-swan-lib/utils"

	"encoding/json"
	"github.com/fatih/color"
	"github.com/filswan/go-swan-lib/client/lotus"
	"github.com/filswan/go-swan-lib/client/swan"
)

const ARIA2_TASK_STATUS_ERROR = "error"
const ARIA2_TASK_STATUS_WAITING = "waiting"
const ARIA2_TASK_STATUS_ACTIVE = "active"
const ARIA2_TASK_STATUS_COMPLETE = "complete"

const DEAL_STATUS_CREATED = "Created"
const DEAL_STATUS_WAITING = "Waiting"
const DEAL_STATUS_SUSPENDING = "Suspending"

const DEAL_STATUS_DOWNLOADING = "Downloading"
const DEAL_STATUS_DOWNLOADED = "Downloaded"
const DEAL_STATUS_DOWNLOAD_FAILED = "DownloadFailed"

const DEAL_STATUS_IMPORT_READY = "ReadyForImport"
const DEAL_STATUS_IMPORTING = "FileImporting"
const DEAL_STATUS_IMPORTED = "FileImported"
const DEAL_STATUS_IMPORT_FAILED = "ImportFailed"
const DEAL_STATUS_ACTIVE = "DealActive"

const ONCHAIN_DEAL_STATUS_ERROR = "StorageDealError"
const ONCHAIN_DEAL_STATUS_ACTIVE = "StorageDealActive"
const ONCHAIN_DEAL_STATUS_NOTFOUND = "StorageDealNotFound"
const ONCHAIN_DEAL_STATUS_WAITTING = "StorageDealWaitingForData"
const ONCHAIN_DEAL_STATUS_ACCEPT = "StorageDealAcceptWait"
const ONCHAIN_DEAL_STATUS_SEALING = "StorageDealSealing"
const ONCHAIN_DEAL_STATUS_AWAITING = "StorageDealAwaitingPreCommit"

const LOTUS_IMPORT_NUMNBER = 20 //Max number of deals to be imported at a time
const LOTUS_SCAN_NUMBER = 100   //Max number of deals to be scanned at a time

const API_POCKET_V1 = "/api/pocket/v1"

var aria2Client *client.Aria2Client
var swanClient *swan.SwanClient

var swanService *SwanService
var aria2Service *Aria2Service
var lotusService *LotusService
var poktService *PoktService

func ParsePoktCmd(cmd []string) {
	if len(cmd) < 2 {
		printPoktUsage()
		return
	}

	subCmd := cmd[1]
	switch subCmd {
	case "start":
		cmdPoktStart(cmd[1:])
		poktHttpServer()
	case "version":
		cmdPoktVersion()
	case "nodeaddr":
		cmdPoktNodeAddr()
	case "balance":
		cmdPoktBalance(cmd[1:])
	case "custodial":
		cmdPoktCustodial(cmd[1:])
	case "noncustodial":
		cmdPoktNonCustodial(cmd[1:])
	case "status":
		cmdPoktStatus()
	default:
		printPoktUsage()
	}

}

func poktStartScan() {
	logs.GetLog().Info("Start...")
	for {
		poktService.StartScan()
		logs.GetLog().Info("Sleeping...")
		//TODO: config
		time.Sleep(time.Second * poktService.PoktScanInterval)
	}
}

func sendHeartbeat2Swan() {
	for {
		logs.GetLog().Info("Start...")
		poktService.SendPoktHeartbeatRequest(swanClient)
		logs.GetLog().Info("Sleeping...")
		time.Sleep(time.Second * poktService.ApiHeartbeatInterval)
	}
}

func printPoktUsage() {
	fmt.Println("SUBCMD:")
	fmt.Println("    pocket")
	fmt.Println("USAGE:")
	fmt.Println("    swan-provider pocket start")
	fmt.Println("    swan-provider pocket version")
	fmt.Println("    swan-provider pocket nodeaddr")
	fmt.Println("    swan-provider pocket balance --addr=0123456789012345678901234567890123456789")
	fmt.Println("    swan-provider pocket status")
	//fmt.Println("    swan-provider pocket custodial")
	//fmt.Println("                         --operatorAddress=0123456789012345678901234567890123456789")
	//fmt.Println("                         --amount=15100000000")
	//fmt.Println("                         --relayChainIDs=0001,0021")
	//fmt.Println("                         --serviceURI=https://pokt.rocks:443")
	//fmt.Println("                         --networkID=mainnet")
	//fmt.Println("                         --fee=10000")
	//fmt.Println("                         --isBefore=false")
	//fmt.Println("    swan-provider pocket noncustodial")
	//fmt.Println("                         --operatorPublicKey=0123456789012345678901234567890123456789012345678901234567890123")
	//fmt.Println("                         --outputAddress=0123456789012345678901234567890123456789")
	//fmt.Println("                         --amount=15100000000")
	//fmt.Println("                         --relayChainIDs=0001,0021")
	//fmt.Println("                         --serviceURI=https://pokt.rocks:443")
	//fmt.Println("                         --networkID=mainnet")
	//fmt.Println("                         --fee=10000")
	//fmt.Println("                         --isBefore=false")

}

func cmdPoktStart(op []string) {
	config.GetPoktConfig()
	logs.SetLogLevel(config.GetPoktConfig().LogLevel)

	poktService = GetMyPoktService()
	checkPoktConfig()

	poktService.StartPoktContainer(op)
	for {
		ver, err := poktService.dkCli.PoktCtnExecVersion()
		if err == nil {
			logs.GetLog().Info("Pocket Node Version Is:", ver.Version)
			time.Sleep(time.Second * 3)
			break
		}
		logs.GetLog().Info("Wait for Pocket Node Available...")
		time.Sleep(time.Second * 3)
	}

	//
	acc, err := poktService.dkCli.PoktCtnExec([]string{"pocket", "accounts", "list"})
	if err != nil {
		logs.GetLog().Error("Get Pocket Accounts error:", err)
	}
	poktService.PoktAddress = strings.Split(acc, "(0) ")[1][0:40]
	poktService.CurStatus.Address = poktService.PoktAddress
	logs.GetLog().Info("Pocket Accounts is:", poktService.PoktAddress)

	// not ready for heartbeat
	go sendHeartbeat2Swan()

	go poktStartScan()
}

func poktHttpServer() {
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	r.Use(cors.Middleware(cors.Config{
		Origins:         "*",
		Methods:         "GET, PUT, POST, DELETE",
		RequestHeaders:  "Origin, Authorization, Content-Type",
		ExposedHeaders:  "",
		MaxAge:          50 * time.Second,
		Credentials:     true,
		ValidateHeaders: false,
	}))

	apiv1 := r.Group(API_POCKET_V1)
	{
		apiv1.GET("/version", HttpGetPoktVersion)
		apiv1.GET("/height", HttpGetPoktCurHeight)
		apiv1.GET("/nodeaddr", HttpGetPoktNodeAddr)
		apiv1.GET("/status", HttpGetPoktStatus)

		apiv1.POST("/balance", HttpGetPoktBalance)
		apiv1.POST("/threshold", HttpGetPoktThreshold)
		apiv1.POST("/custodial", HttpGetPoktCustodial)
		apiv1.POST("/noncustodial", HttpGetPoktNonCustodial)
	}

	port := config.GetPoktConfig().Pokt.PoktServerApiPort
	err := r.Run(":" + strconv.Itoa(port))
	if err != nil {
		logs.GetLog().Fatal(err)
	}
}

func cmdPoktVersion() {
	params := ""
	confPokt := config.GetPoktConfig().Pokt
	selfUrl := utils.UrlJoin(confPokt.PoktServerApiUrl, API_POCKET_V1)

	apiUrl := utils.UrlJoin(selfUrl, "version")
	response, err := web.HttpGetNoToken(apiUrl, params)
	if err != nil {
		fmt.Printf("Get Pocket Version err: %s \n", err)
		return
	}

	res := &models.VersionResponse{}
	err = json.Unmarshal(response, res)
	if err != nil {
		fmt.Printf("Parse Response (%s) err: %s \n", response, err)
		return
	}
	title := color.New(color.FgGreen).Sprintf("%s", "Pocket Version")
	value := color.New(color.FgYellow).Sprintf("%s", res.Data.Version)
	fmt.Printf("%s\t: %s\n", title, value)

}

func cmdPoktNodeAddr() {
	params := ""
	confPokt := config.GetPoktConfig().Pokt
	selfUrl := utils.UrlJoin(confPokt.PoktServerApiUrl, API_POCKET_V1)

	apiUrl := utils.UrlJoin(selfUrl, "nodeaddr")
	response, err := web.HttpGetNoToken(apiUrl, params)
	if err != nil {
		fmt.Printf("Get Pocket Node Address err: %s \n", err)
		return
	}

	res := &models.Response{}
	err = json.Unmarshal(response, res)
	if err != nil {
		fmt.Printf("Parse Response (%s) err: %s \n", response, err)
		return
	}

	title := color.New(color.FgGreen).Sprintf("%s", "Node Address")
	value := color.New(color.FgYellow).Sprintf("%s", res.Data)
	fmt.Printf("%s\t: %s\n", title, value)

}

func cmdPoktBalance(op []string) {
	fs := flag.NewFlagSet("Balance", flag.ExitOnError)
	addr := fs.String("addr", "", "address to lookup")
	err := fs.Parse(op[1:])
	if *addr == "" || err != nil {
		printPoktUsage()
		return
	}

	params := &models.BalancePoktParams{Height: 0, Address: *addr}
	confPokt := config.GetPoktConfig().Pokt
	selfUrl := utils.UrlJoin(confPokt.PoktServerApiUrl, API_POCKET_V1)
	apiUrl := utils.UrlJoin(selfUrl, "balance")

	response, err := web.HttpPostNoToken(apiUrl, params)
	if err != nil {
		fmt.Printf("Get Pocket Balance err: %s params: %+v \n", err, params)
		return
	}

	res := &models.BalanceHttpResponse{}
	err = json.Unmarshal(response, res)
	if err != nil {
		fmt.Printf("Parse Response (%s) err: %s  params: %+v \n", response, err, params)
		return
	}

	title := color.New(color.FgGreen).Sprintf("%s", "Address")
	value := color.New(color.FgYellow).Sprintf("%s", res.Data.Address)
	fmt.Printf("%s\t: %s\n", title, value)

	title = color.New(color.FgGreen).Sprintf("%s", "Balance")
	value = color.New(color.FgYellow).Sprintf("%s", res.Data.Balance)
	fmt.Printf("%s\t: %s\n", title, value)

}

func cmdPoktCustodial(op []string) {

	fs := flag.NewFlagSet("Custodial", flag.ExitOnError)
	operatorAddress := fs.String("operatorAddress", "", "")
	amount := fs.String("amount", "", "")
	relayChainIDs := fs.String("relayChainIDs", "", "")
	serviceURI := fs.String("serviceURI", "", "")
	networkID := fs.String("networkID", "", "")
	fee := fs.String("fee", "", "")
	isBefore := fs.String("isBefore", "", "")

	err := fs.Parse(op[1:])
	if *operatorAddress == "" || *amount == "" || *relayChainIDs == "" || *serviceURI == "" || *networkID == "" || *fee == "" || *isBefore == "" || err != nil {
		printPoktUsage()
		return
	}

	params := ""
	confPokt := config.GetPoktConfig().Pokt
	selfUrl := utils.UrlJoin(confPokt.PoktServerApiUrl, API_POCKET_V1)

	apiUrl := utils.UrlJoin(selfUrl, "custodial")
	response, err := web.HttpPostNoToken(apiUrl, params)
	if err != nil {
		fmt.Printf("Get Pocket Custodial err: %s \n", err)
		return
	}

	res := &models.StatusResponse{}
	err = json.Unmarshal(response, res)
	if err != nil {
		fmt.Printf("Parse Response (%s) err: %s \n", response, err)
		return
	}

	fmt.Printf("Pocket Sratus is: %+v \n", res.Data)
}

func cmdPoktNonCustodial(op []string) {

	fs := flag.NewFlagSet("Custodial", flag.ExitOnError)
	operatorPublicKey := fs.String("operatorPublicKey", "", "")
	outputAddress := fs.String("outputAddress", "", "")
	amount := fs.String("amount", "", "")
	relayChainIDs := fs.String("relayChainIDs", "", "")
	serviceURI := fs.String("serviceURI", "", "")
	networkID := fs.String("networkID", "", "")
	fee := fs.String("fee", "", "")
	isBefore := fs.String("isBefore", "", "")

	err := fs.Parse(op[1:])
	if *operatorPublicKey == "" || *outputAddress == "" || *amount == "" || *relayChainIDs == "" || *serviceURI == "" || *networkID == "" || *fee == "" || *isBefore == "" || err != nil {
		printPoktUsage()
		return
	}

	params := ""
	confPokt := config.GetPoktConfig().Pokt
	selfUrl := utils.UrlJoin(confPokt.PoktServerApiUrl, API_POCKET_V1)

	apiUrl := utils.UrlJoin(selfUrl, "nonCustodial")
	response, err := web.HttpPostNoToken(apiUrl, params)
	if err != nil {
		fmt.Printf("Get Pocket NonCustodial err: %s \n", err)
		return
	}

	res := &models.StatusResponse{}
	err = json.Unmarshal(response, res)
	if err != nil {
		fmt.Printf("Parse Response (%s) err: %s \n", response, err)
		return
	}

	fmt.Printf("Pocket NonCustodial is: %+v \n", res)
}

func cmdPoktStatus() {
	params := ""
	confPokt := config.GetPoktConfig().Pokt
	selfUrl := utils.UrlJoin(confPokt.PoktServerApiUrl, API_POCKET_V1)

	apiUrl := utils.UrlJoin(selfUrl, "status")
	response, err := web.HttpGetNoToken(apiUrl, params)
	if err != nil {
		fmt.Printf("Get Pocket Status err: %s \n", err)
		return
	}

	res := &models.StatusResponse{}
	err = json.Unmarshal(response, res)
	if err != nil {
		fmt.Printf("Parse Response (%s) err: %s \n", response, err)
		return
	}

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

///////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////

func AdminOfflineDeal() {
	swanService = GetSwanService()
	aria2Service = GetAria2Service()
	lotusService = GetLotusService()

	aria2Client = SetAndCheckAria2Config()
	swanClient = SetAndCheckSwanConfig()
	checkMinerExists()
	checkLotusConfig()

	//logs.GetLogger().Info("swan token:", swanClient.SwanToken)
	swanService.UpdateBidConf(swanClient)
	go swanSendHeartbeatRequest()
	go aria2CheckDownloadStatus()
	go aria2StartDownload()
	go lotusStartImport()
	go lotusStartScan()
}

func SetAndCheckAria2Config() *client.Aria2Client {
	aria2DownloadDir := config.GetConfig().Aria2.Aria2DownloadDir
	aria2CandidateDirs := config.GetConfig().Aria2.Aria2CandidateDirs
	aria2Host := config.GetConfig().Aria2.Aria2Host
	aria2Port := config.GetConfig().Aria2.Aria2Port
	aria2Secret := config.GetConfig().Aria2.Aria2Secret
	aria2MaxDownloadingTasks := config.GetConfig().Aria2.Aria2MaxDownloadingTasks

	if !utils.IsDirExists(aria2DownloadDir) {
		err := fmt.Errorf("aria2 down load dir:%s not exits, please set config:aria2->aria2_download_dir", aria2DownloadDir)
		logs.GetLogger().Fatal(err)
	}

	for _, dir := range aria2CandidateDirs {
		if !utils.IsDirExists(dir) {
			err := fmt.Errorf("aria2 down load dir:%s not exits, please set config:aria2->aria2_candidate_dirs", dir)
			logs.GetLogger().Fatal(err)
		}
	}

	if len(aria2Host) == 0 {
		logs.GetLogger().Fatal("please set config:aria2->aria2_host")
	}

	aria2Client = client.GetAria2Client(aria2Host, aria2Secret, aria2Port)
	if aria2MaxDownloadingTasks <= 0 {
		logs.GetLogger().Warning("config [aria2].aria2_max_downloading_tasks is " + strconv.Itoa(aria2MaxDownloadingTasks) + ", no CAR file will be downloaded")
	}
	aria2ChangeMaxConcurrentDownloads := aria2Client.ChangeMaxConcurrentDownloads(strconv.Itoa(aria2MaxDownloadingTasks))
	if aria2ChangeMaxConcurrentDownloads == nil {
		err := fmt.Errorf("failed to set [aria2].aria2_max_downloading_tasks, please check the Aria2 service")
		logs.GetLogger().Fatal(err)
	}

	if aria2ChangeMaxConcurrentDownloads.Error != nil {
		err := fmt.Errorf(aria2ChangeMaxConcurrentDownloads.Error.Message)
		logs.GetLogger().Fatal(err)
	}
	return aria2Client
}

func SetAndCheckSwanConfig() *swan.SwanClient {
	var err error
	swanApiUrl := config.GetConfig().Main.SwanApiUrl
	swanApiKey := config.GetConfig().Main.SwanApiKey
	swanAccessToken := config.GetConfig().Main.SwanAccessToken

	if len(swanApiUrl) == 0 {
		logs.GetLogger().Fatal("please set config:main->api_url")
	}

	if len(swanApiKey) == 0 {
		logs.GetLogger().Fatal("please set config:main->api_key")
	}

	if len(swanAccessToken) == 0 {
		logs.GetLogger().Fatal("please set config:main->access_token")
	}

	swanClient, err := swan.GetClient(swanApiUrl, swanApiKey, swanAccessToken, "")
	if err != nil {
		logs.GetLogger().Error(err)
		logs.GetLogger().Error(constants.ERROR_LAUNCH_FAILED)
		logs.GetLogger().Fatal(constants.INFO_ON_HOW_TO_CONFIG)
	}

	return swanClient
}

func checkMinerExists() {
	err := swanService.SendHeartbeatRequest(swanClient)
	if err != nil {
		logs.GetLogger().Info(err)
		if strings.Contains(err.Error(), "Miner Not found") {
			logs.GetLogger().Error("Cannot find your miner:", swanService.MinerFid)
		}
		logs.GetLogger().Error(constants.ERROR_LAUNCH_FAILED)
		logs.GetLogger().Fatal(constants.INFO_ON_HOW_TO_CONFIG)
	}
}

func checkLotusConfig() {
	logs.GetLogger().Info("Start testing lotus config.")

	if lotusService == nil {
		logs.GetLogger().Fatal("error in config")
	}

	lotusMarket := lotusService.LotusMarket
	lotusClient := lotusService.LotusClient
	if utils.IsStrEmpty(&lotusMarket.ApiUrl) {
		logs.GetLogger().Fatal("please set config:lotus->market_api_url")
	}

	if utils.IsStrEmpty(&lotusMarket.AccessToken) {
		logs.GetLogger().Fatal("please set config:lotus->market_access_token")
	}

	if utils.IsStrEmpty(&lotusMarket.ClientApiUrl) {
		logs.GetLogger().Fatal("please set config:lotus->client_api_url")
	}

	isWriteAuth, err := lotus.LotusCheckAuth(lotusMarket.ApiUrl, lotusMarket.AccessToken, libconstants.LOTUS_AUTH_WRITE)
	if err != nil {
		logs.GetLogger().Error(err)
		logs.GetLogger().Fatal("please check config:lotus->market_api_url, lotus->market_access_token")
	}

	if !isWriteAuth {
		logs.GetLogger().Fatal("market access token should have write access right")
	}

	currentEpoch, err := lotusClient.LotusGetCurrentEpoch()
	if err != nil {
		logs.GetLogger().Error(err)
		logs.GetLogger().Fatal("please check config:lotus->client_api_url")
	}

	logs.GetLogger().Info("current epoch:", *currentEpoch)
	logs.GetLogger().Info("Pass testing lotus config.")
}

func checkPoktConfig() {
	//TODO:Check

	if poktService == nil {
		logs.GetLog().Fatal("error in config")
	}

	logs.GetLog().Info("Pass testing pocket config.")
}

func swanSendHeartbeatRequest() {
	for {
		logs.GetLogger().Info("Start...")
		swanService.SendHeartbeatRequest(swanClient)
		logs.GetLogger().Info("Sleeping...")
		time.Sleep(swanService.ApiHeartbeatInterval)
	}
}

func aria2CheckDownloadStatus() {
	for {
		logs.GetLogger().Info("Start...")
		aria2Service.CheckAndRestoreSuspendingStatus(aria2Client, swanClient)
		aria2Service.CheckDownloadStatus(aria2Client, swanClient)
		logs.GetLogger().Info("Sleeping...")
		time.Sleep(time.Minute)
	}
}

func aria2StartDownload() {
	for {
		logs.GetLogger().Info("Start...")
		aria2Service.StartDownload(aria2Client, swanClient)
		logs.GetLogger().Info("Sleeping...")
		time.Sleep(time.Minute)
	}
}

func lotusStartImport() {
	for {
		logs.GetLogger().Info("Start...")
		lotusService.StartImport(swanClient)
		logs.GetLogger().Info("Sleeping...")
		time.Sleep(lotusService.ImportIntervalSecond)
	}
}

func lotusStartScan() {
	for {
		logs.GetLogger().Info("Start...")
		lotusService.StartScan(swanClient)
		logs.GetLogger().Info("Sleeping...")
		time.Sleep(lotusService.ScanIntervalSecond)
	}
}

func UpdateDealInfoAndLog(deal *libmodel.OfflineDeal, newSwanStatus string, filefullpath *string, messages ...string) {
	note := ""
	if newSwanStatus != DEAL_STATUS_DOWNLOADING {
		note = GetNote(messages...)
		note = utils.FirstLetter2Upper(note)
	} else {
		note = messages[0]
	}

	if newSwanStatus == DEAL_STATUS_IMPORT_FAILED || newSwanStatus == DEAL_STATUS_DOWNLOAD_FAILED {
		logs.GetLogger().Warn(GetLog(deal, note))
	} else {
		logs.GetLogger().Info(GetLog(deal, note))
	}

	filefullpathTemp := deal.FilePath
	if filefullpath != nil {
		filefullpathTemp = *filefullpath
	}

	if deal.Status == newSwanStatus && deal.Note == note && deal.FilePath == filefullpathTemp {
		return
	}

	err := UpdateOfflineDeal(swanClient, deal.Id, newSwanStatus, &note, &filefullpathTemp)
	if err != nil {
		logs.GetLogger().Error(GetLog(deal, constants.UPDATE_OFFLINE_DEAL_STATUS_FAIL))
	} else {
		msg := GetLog(deal, "set status to:"+newSwanStatus, "set note to:"+note, "set filepath to:"+filefullpathTemp)
		if newSwanStatus == DEAL_STATUS_IMPORT_FAILED || newSwanStatus == DEAL_STATUS_DOWNLOAD_FAILED {
			logs.GetLogger().Warn(msg)
		} else {
			logs.GetLogger().Info(msg)
		}
	}
}

func UpdateStatusAndLog(deal *libmodel.OfflineDeal, newSwanStatus string, messages ...string) {
	UpdateDealInfoAndLog(deal, newSwanStatus, nil, messages...)
}

func GetLog(deal *libmodel.OfflineDeal, messages ...string) string {
	text := GetNote(messages...)
	msg := fmt.Sprintf("taskName:%s, dealCid:%s, %s", *deal.TaskName, deal.DealCid, text)
	return msg
}

func GetNote(messages ...string) string {
	separator := ","
	result := ""
	if messages == nil {
		return result
	}
	for _, message := range messages {
		if message != "" {
			result = result + separator + message
		}
	}

	result = strings.TrimPrefix(result, separator)
	result = strings.TrimSuffix(result, separator)
	return result
}

func GetOfflineDeals(swanClient *swan.SwanClient, dealStatus string, minerFid string, limit *int) []*libmodel.OfflineDeal {
	pageNum := 1
	params := swan.GetOfflineDealsByStatusParams{
		DealStatus: dealStatus,
		ForMiner:   true,
		MinerFid:   &minerFid,
		PageNum:    &pageNum,
		PageSize:   limit,
	}

	offlineDeals, err := swanClient.GetOfflineDealsByStatus(params)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil
	}

	return offlineDeals
}

func UpdateOfflineDeal(swanClient *swan.SwanClient, dealId int, status string, note, filePath *string) error {
	params := &swan.UpdateOfflineDealParams{
		DealId:   dealId,
		Status:   status,
		Note:     note,
		FilePath: filePath,
	}

	err := swanClient.UpdateOfflineDeal(*params)
	if err != nil {
		logs.GetLogger().Error()
		return err
	}

	return nil
}

func UpdateOfflineDealStatus(swanClient *swan.SwanClient, dealId int, status string) error {
	params := &swan.UpdateOfflineDealParams{
		DealId: dealId,
		Status: status,
	}

	err := swanClient.UpdateOfflineDeal(*params)
	if err != nil {
		logs.GetLogger().Error()
		return err
	}

	return nil
}
