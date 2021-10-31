package main

type BucketInfo struct {
	Up_hosts  []string `json:"up_hosts"`
	Rs_hosts  []string `json:"rs_hosts"`
	Bucket    string   `json:"bucket"`
	Part      uint     `json:"part"`
	Ak        string   `json:"ak"`
	Sk        string   `json:"sk"`
	Addr      string   `json:"addr"`
	Delete    bool     `json:"delete"`
	Down_path string   `json:"down_path"`
	Sim       bool     `json:"sim"`
}

type SectorInfo struct {
	SectorNumber string
	SectorSize   string
	Paux         string
	Cache        []string
	Sealed       string
}
