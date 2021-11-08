package service

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"swan-provider/common/constants"
	"swan-provider/config"
	"time"

	"github.com/filswan/go-swan-lib/client"
	"github.com/filswan/go-swan-lib/client/swan"
	"github.com/filswan/go-swan-lib/logs"
	libmodel "github.com/filswan/go-swan-lib/model"
	"github.com/filswan/go-swan-lib/utils"
)

type Aria2Service struct {
	MinerFid    string
	DownloadDir string
}

func GetAria2Service() *Aria2Service {
	aria2Service := &Aria2Service{
		MinerFid:    config.GetConfig().Main.MinerFid,
		DownloadDir: config.GetConfig().Aria2.Aria2DownloadDir,
	}

	_, err := os.Stat(aria2Service.DownloadDir)
	if err != nil {
		logs.GetLogger().Error(constants.ERROR_LAUNCH_FAILED)
		logs.GetLogger().Error("Your download directory:", aria2Service.DownloadDir, " not exists.")
		logs.GetLogger().Fatal(constants.INFO_ON_HOW_TO_CONFIG)
	}

	return aria2Service
}

func (aria2Service *Aria2Service) findNextDealReady2Download(swanClient *swan.SwanClient) *libmodel.OfflineDeal {
	deals := swanClient.SwanGetOfflineDeals(aria2Service.MinerFid, DEAL_STATUS_CREATED, "1")
	if len(deals) == 0 {
		deals = swanClient.SwanGetOfflineDeals(aria2Service.MinerFid, DEAL_STATUS_WAITING, "1")
	}

	if len(deals) > 0 {
		offlineDeal := deals[0]
		return &offlineDeal
	}

	return nil
}

func (aria2Service *Aria2Service) CheckDownloadStatus4Deal(aria2Client *client.Aria2Client, swanClient *swan.SwanClient, deal libmodel.OfflineDeal, gid string) {
	aria2Status := aria2Client.GetDownloadStatus(gid)
	if aria2Status == nil {
		UpdateStatusAndLog(deal, DEAL_STATUS_DOWNLOAD_FAILED, "get download status failed for gid:"+gid, "no response from aria2")
		return
	}

	if aria2Status.Error != nil {
		UpdateStatusAndLog(deal, DEAL_STATUS_DOWNLOAD_FAILED, "get download status failed for gid:"+gid, aria2Status.Error.Message)
		return
	}

	if len(aria2Status.Result.Files) != 1 {
		UpdateStatusAndLog(deal, DEAL_STATUS_DOWNLOAD_FAILED, "get download status failed for gid:"+gid, "wrong file amount")
		return
	}

	result := aria2Status.Result
	file := result.Files[0]
	filePath := file.Path
	fileSize := utils.GetInt64FromStr(file.Length)

	switch result.Status {
	case ARIA2_TASK_STATUS_ERROR:
		UpdateStatusAndLog(deal, DEAL_STATUS_DOWNLOAD_FAILED, "download error for gid:"+gid, result.ErrorMessage)
	case ARIA2_TASK_STATUS_ACTIVE:
		fileSizeDownloaded := utils.GetFileSize(filePath)
		completedLen := utils.GetInt64FromStr(file.CompletedLength)
		var completePercent float64 = 0
		if fileSize > 0 {
			completePercent = float64(completedLen) / float64(fileSize) * 100
		}
		downloadSpeed := utils.GetInt64FromStr(result.DownloadSpeed) / 1000
		note := fmt.Sprintf("downloading, complete: %.2f%%, speed: %dKiB", completePercent, downloadSpeed)
		UpdateDealInfoAndLog(deal, DEAL_STATUS_DOWNLOADING, &filePath, &fileSizeDownloaded, note)
	case ARIA2_TASK_STATUS_COMPLETE:
		fileSizeDownloaded := utils.GetFileSize(filePath)
		if fileSizeDownloaded >= 0 {
			UpdateDealInfoAndLog(deal, DEAL_STATUS_DOWNLOADED, &filePath, &fileSizeDownloaded, gid)
		} else {
			UpdateDealInfoAndLog(deal, DEAL_STATUS_DOWNLOAD_FAILED, &filePath, &fileSizeDownloaded, "file not found on its download path")
		}
	default:
		UpdateDealInfoAndLog(deal, DEAL_STATUS_DOWNLOAD_FAILED, &filePath, &fileSize, result.ErrorMessage)
	}
}

func (aria2Service *Aria2Service) CheckDownloadStatus(aria2Client *client.Aria2Client, swanClient *swan.SwanClient) {
	downloadingDeals := swanClient.SwanGetOfflineDeals(aria2Service.MinerFid, DEAL_STATUS_DOWNLOADING)

	for _, deal := range downloadingDeals {
		gid := strings.Trim(deal.Note, " ")
		if gid == "" {
			UpdateStatusAndLog(deal, DEAL_STATUS_DOWNLOAD_FAILED, "download gid not found in offline_deals.note")
			continue
		}

		aria2Service.CheckDownloadStatus4Deal(aria2Client, swanClient, deal, gid)
	}
}

func (aria2Service *Aria2Service) StartDownload4Deal(deal libmodel.OfflineDeal, aria2Client *client.Aria2Client, swanClient *swan.SwanClient) {
	logs.GetLogger().Info(GetLog(deal, "start downloading"))
	urlInfo, err := url.Parse(deal.FileSourceUrl)
	if err != nil {
		UpdateStatusAndLog(deal, DEAL_STATUS_DOWNLOAD_FAILED, "parse source file url error,", err.Error())
		return
	}

	outFilename := urlInfo.Path
	if strings.HasPrefix(urlInfo.RawQuery, "filename=") {
		outFilename = strings.TrimPrefix(urlInfo.RawQuery, "filename=")
		outFilename = filepath.Join(urlInfo.Path, outFilename)
	}
	outFilename = strings.TrimLeft(outFilename, "/")

	today := time.Now()
	timeStr := fmt.Sprintf("%d%02d", today.Year(), today.Month())
	outDir := filepath.Join(aria2Service.DownloadDir, strconv.Itoa(deal.UserId), timeStr)

	aria2Download := aria2Client.DownloadFile(deal.FileSourceUrl, outDir, outFilename)

	if aria2Download == nil {
		UpdateStatusAndLog(deal, DEAL_STATUS_DOWNLOAD_FAILED, "no response when asking aria2 to download")
		return
	}

	if aria2Download.Error != nil {
		UpdateStatusAndLog(deal, DEAL_STATUS_DOWNLOAD_FAILED, aria2Download.Error.Message)
		return
	}

	if aria2Download.Gid == "" {
		UpdateStatusAndLog(deal, DEAL_STATUS_DOWNLOAD_FAILED, "no gid returned when asking aria2 to download")
		return
	}

	aria2Service.CheckDownloadStatus4Deal(aria2Client, swanClient, deal, aria2Download.Gid)
}

func (aria2Service *Aria2Service) StartDownload(aria2Client *client.Aria2Client, swanClient *swan.SwanClient) {
	downloadingDeals := swanClient.SwanGetOfflineDeals(aria2Service.MinerFid, DEAL_STATUS_DOWNLOADING)
	countDownloadingDeals := len(downloadingDeals)
	if countDownloadingDeals >= ARIA2_MAX_DOWNLOADING_TASKS {
		return
	}

	for i := 1; i <= ARIA2_MAX_DOWNLOADING_TASKS-countDownloadingDeals; i++ {
		deal2Download := aria2Service.findNextDealReady2Download(swanClient)
		if deal2Download == nil {
			logs.GetLogger().Info("No offline deal to download")
			break
		}

		aria2Service.StartDownload4Deal(*deal2Download, aria2Client, swanClient)
		time.Sleep(1 * time.Second)
	}
}