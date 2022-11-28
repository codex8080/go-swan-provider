package models

import "time"

type Response struct {
	Status  string `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
	Data    string `json:"data"`
}

///////////////////////////////////////////////////////////////////////////////

type VersionResponse struct {
	Status  string      `json:"status"`
	Code    string      `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    VersionData `json:"data"`
}

type VersionData struct {
	Version string `json:"version"`
}

///////////////////////////////////////////////////////////////////////////////

type HeightPoktParams struct {
}

type HeightData struct {
	Height uint64 `json:"height"`
}

type HeightResponse struct {
	Status  string     `json:"status"`
	Code    string     `json:"code"`
	Message string     `json:"message,omitempty"`
	Data    HeightData `json:"data"`
}

///////////////////////////////////////////////////////////////////////////////

type BalancePoktParams struct {
	Height  uint64 `json:"height"`
	Address string `json:"address"`
}

type BalanceData struct {
	Balance uint64 `json:"balance"`
}

type BalanceCmdData struct {
	Height  uint64 `json:"height"`
	Address string `json:"address"`
	Balance string `json:"balance"`
}

type BalanceHttpResponse struct {
	Status  string         `json:"status"`
	Code    string         `json:"code"`
	Message string         `json:"message,omitempty"`
	Data    BalanceCmdData `json:"data"`
}

///////////////////////////////////////////////////////////////////////////////

type PoktSignInfoParams struct {
	Height  int64  `json:"height"`
	Address string `json:"address"`
	Page    int    `json:"page,omitempty"`
	PerPage int    `json:"per_page,omitempty"`
}

type SigningInfo struct {
	Address     string `json:"address"`
	StartHeight int64  `json:"start_height"`
	IndexOffset int64  `json:"index_offset"`

	JailedUntil         time.Time `json:"jailed_until"`
	MissedBlocksCounter int64     `json:"missed_blocks_counter"`
	JailedBlocksCounter int64     `json:"jailed_blocks_counter"`
}

type PoktSigningInfoResponse struct {
	Result     []*SigningInfo `json:"result"`
	Page       int            `json:"page"`
	TotalPages int            `json:"total_pages"`
}

///////////////////////////////////////////////////////////////////////////////

type PoktNodeParams struct {
	Height  uint64 `json:"height"`
	Address string `json:"address"`
}

type NodeData struct {
	Address       string    `json:"address"`
	PublicKey     string    `json:"public_key"`
	Jailed        bool      `json:"jailed"`
	Status        uint64    `json:"status"`
	Chains        []string  `json:"chains"`
	ServiceURL    string    `json:"service_url"`
	StakedTokens  string    `json:"tokens"`
	UnstakingTime time.Time `json:"unstaking_time"`
	OutputAddress string    `json:"output_address,omitempty"`
}

type PoktNodeResponse struct {
	Result     []*NodeData `json:"result"`
	Page       int         `json:"page"`
	TotalPages int         `json:"total_pages"`
}

///////////////////////////////////////////////////////////////////////////////

type PoktSupplyParams struct {
	Height uint64 `json:"height"`
}

type PoktSupplyResponse struct {
	NodeStaked    string `json:"node_staked"`
	AppStaked     string `json:"app_staked"`
	Dao           string `json:"dao"`
	TotalStaked   string `json:"total_staked"`
	TotalUnStaked string `json:"total_unstaked"`
	Total         string `json:"total"`
}

///////////////////////////////////////////////////////////////////////////////

type SignInfo struct {
	Address     string `json:"address"`
	StartHeight uint64 `json:"start_height"`
	IndexOffset uint64 `json:"index_offset"`

	JailedUntil         time.Time `json:"jailed_until"`
	MissedBlocksCounter uint64    `json:"missed_blocks_counter"`
	JailedBlocksCounter uint64    `json:"jailed_blocks_counter"`
}

type SignInfoResponse struct {
	Result     []*SignInfo `json:"result"`
	Page       int         `json:"page"`
	TotalPages int         `json:"total_pages"`
}

///////////////////////////////////////////////////////////////////////////////

type StatusParams struct {
	Address string `json:"address"`
}

type StatusResponse struct {
	Status  string     `json:"status"`
	Code    string     `json:"code"`
	Message string     `json:"message,omitempty"`
	Data    StatusData `json:"data"`
}

type StatusData struct {
	Version     string    `json:"version"`
	Height      uint64    `json:"height"`
	Address     string    `json:"address"`
	Balance     uint64    `json:"balance"`
	Award       string    `json:"award"`
	Jailed      bool      `json:"jailed"`
	JailedBlock uint64    `json:"jailedBlock"`
	JailedUntil time.Time `json:"jailedUntil"`
}

///////////////////////////////////////////////////////////////////////////////

type ThresholdResponse struct {
	Status  string        `json:"status"`
	Code    string        `json:"code"`
	Message string        `json:"message,omitempty"`
	Data    ThresholdData `json:"data"`
}

type ThresholdData struct {
	Address   string `json:"address"`
	Threshold uint64 `json:"threshold"`
	Active    bool   `json:"active"`
}

type ThresholdParams struct {
	Address   string `json:"address"`
	Threshold uint64 `json:"threshold"`
}

///////////////////////////////////////////////////////////////////////////////

type CustodialResponse struct {
	Status  string        `json:"status"`
	Code    string        `json:"code"`
	Message string        `json:"message,omitempty"`
	Data    CustodialData `json:"data"`
}

type CustodialData struct {
	Result string `json:"result"`
}

type CustodialParams struct {
	Address       string `json:"address"`
	Amount        string `json:"amount"`
	RelayChainIDs string `json:"relay_chain_ids"`
	ServiceURI    string `json:"service_url"`
	NetworkID     string `json:"network_id"`
	Fee           string `json:"fee"`
	IsBefore      string `json:"is_before"`
}

///////////////////////////////////////////////////////////////////////////////

type NonCustodialResponse struct {
	Status  string           `json:"status"`
	Code    string           `json:"code"`
	Message string           `json:"message,omitempty"`
	Data    NonCustodialData `json:"data"`
}

type NonCustodialData struct {
	Result string `json:"result"`
}

type NonCustodialParams struct {
	PubKey        string `json:"public_key"`
	OutputAddr    string `json:"output_addr"`
	Amount        string `json:"amount"`
	RelayChainIDs string `json:"relay_chain_ids"`
	ServiceURI    string `json:"service_url"`
	NetworkID     string `json:"network_id"`
	Fee           string `json:"fee"`
	IsBefore      string `json:"is_before"`
}

///////////////////////////////////////////////////////////////////////////////
