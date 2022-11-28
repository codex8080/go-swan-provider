package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"swan-provider/logs"
)

type Configuration struct {
	Port     int    `toml:"port"`
	Release  bool   `toml:"release"`
	LogLevel string `toml:"log_level"`
	Aria2    aria2  `toml:"aria2"`
	Main     main   `toml:"main"`
	Bid      bid    `toml:"bid"`
	Lotus    lotus  `toml:"lotus"`
	Pokt     pokt   `toml:"pokt"`
}

type lotus struct {
	ClientApiUrl      string `toml:"client_api_url"`
	MarketApiUrl      string `toml:"market_api_url"`
	MarketAccessToken string `toml:"market_access_token"`
}

type pokt struct {
	PoktApiUrl        string        `toml:"pokt_api_url"`
	PoktAccessToken   string        `toml:"pokt_access_token"`
	PoktAddress       string        `toml:"pokt_address"`
	PoktDockerImage   string        `toml:"pokt_docker_image"`
	PoktDockerName    string        `toml:"pokt_docker_name"`
	PoktConfigPath    string        `toml:"pokt_config_path"`
	PoktScanInterval  time.Duration `toml:"pokt_scan_interval"`
	PoktServerApiUrl  string        `toml:"pokt_server_api_url"`
	PoktServerApiPort int           `toml:"pokt_server_api_port"`
}

type aria2 struct {
	Aria2DownloadDir         string   `toml:"aria2_download_dir"`
	Aria2CandidateDirs       []string `toml:"aria2_candidate_dirs"`
	Aria2Host                string   `toml:"aria2_host"`
	Aria2Port                int      `toml:"aria2_port"`
	Aria2Secret              string   `toml:"aria2_secret"`
	Aria2AutoDeleteCarFile   bool     `toml:"aria2_auto_delete_car_file"`
	Aria2MaxDownloadingTasks int      `toml:"aria2_max_downloading_tasks"`
}

type main struct {
	SwanApiUrl               string        `toml:"api_url"`
	SwanApiKey               string        `toml:"api_key"`
	SwanAccessToken          string        `toml:"access_token"`
	SwanApiHeartbeatInterval time.Duration `toml:"api_heartbeat_interval"`
	MinerFid                 string        `toml:"miner_fid"`
	LotusImportInterval      time.Duration `toml:"import_interval"`
	LotusScanInterval        time.Duration `toml:"scan_interval"`
}

type bid struct {
	BidMode             int `toml:"bid_mode"`
	ExpectedSealingTime int `toml:"expected_sealing_time"`
	StartEpoch          int `toml:"start_epoch"`
	AutoBidDealPerDay   int `toml:"auto_bid_deal_per_day"`
}

var config *Configuration

func InitConfig() {
	configPath := os.Getenv("SWAN_PATH")
	if configPath == "" {
		homedir, err := os.UserHomeDir()
		if err != nil {
			logs.GetLog().Fatal("Cannot get home directory.")
		}

		configPath = filepath.Join(homedir, ".swan/")
	}

	configFile := filepath.Join(configPath, "provider/config.toml")
	logs.GetLog().Debug("Your config file is:", configFile)
	if metaData, err := toml.DecodeFile(configFile, &config); err != nil {
		logs.GetLog().Fatal("error:", err)
	} else {
		if !requiredFieldsAreGiven(metaData) {
			logs.GetLog().Fatal("required fields not given")
		}
	}

	InitPoktConfig(filepath.Join(configPath, "provider/config-pokt.toml"))

}

func requiredFieldsAreGiven(metaData toml.MetaData) bool {
	requiredFields := [][]string{
		{"port"},
		{"release"},

		{"lotus"},
		{"pokt"},
		{"aria2"},
		{"main"},
		{"bid"},

		{"lotus", "client_api_url"},
		{"lotus", "market_api_url"},
		{"lotus", "market_access_token"},

		{"aria2", "aria2_download_dir"},
		{"aria2", "aria2_host"},
		{"aria2", "aria2_port"},
		{"aria2", "aria2_secret"},
		{"aria2", "aria2_max_downloading_tasks"},
		{"aria2", "aria2_auto_delete_car_file"},

		{"main", "api_url"},
		{"main", "miner_fid"},
		{"main", "import_interval"},
		{"main", "scan_interval"},
		{"main", "api_key"},
		{"main", "access_token"},
		{"main", "api_heartbeat_interval"},

		{"bid", "bid_mode"},
		{"bid", "expected_sealing_time"},
		{"bid", "start_epoch"},
		{"bid", "auto_bid_deal_per_day"},
	}

	for _, v := range requiredFields {
		if !metaData.IsDefined(v...) {
			logs.GetLog().Fatal("required conf fields ", v)
		}
	}

	return true
}

func InitPoktConfig(configFile string) {
	logs.GetLog().Debug("Your pokt config file is:", configFile)

	if metaData, err := toml.DecodeFile(configFile, &config); err != nil {
		logs.GetLog().Fatal("error:", err)
	} else {
		if !requiredPoktAreGiven(metaData) {
			logs.GetLog().Fatal("required fields not given")
		}
	}
}

func requiredPoktAreGiven(metaData toml.MetaData) bool {
	requiredFields := [][]string{
		{"pokt", "pokt_api_url"},
		{"pokt", "pokt_access_token"},
		{"pokt", "pokt_docker_image"},
		{"pokt", "pokt_docker_name"},
		{"pokt", "pokt_config_path"},
		{"pokt", "pokt_scan_interval"},
		{"pokt", "pokt_server_api_url"},
		{"pokt", "pokt_server_api_port"},
	}

	for _, v := range requiredFields {
		if !metaData.IsDefined(v...) {
			logs.GetLog().Fatal("required conf fields ", v)
		}
	}

	return true
}

func GetConfig() Configuration {
	if config == nil {
		InitConfig()
	}
	return *config
}

func GetPoktConfig() Configuration {
	if config == nil {
		InitConfig()
	}
	// fmt.Printf("%+v\n", config.Pokt)
	return *config
}
