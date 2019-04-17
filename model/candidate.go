package model

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/iost-official/iost-api/model/db"
	"github.com/jszwec/csvutil"
)

var (
	offchainInfos = make(map[string]*CandidateOffchainInfo)

	emptyOffchainInfo = &CandidateOffchainInfo{}
)

func init() {
	initOffchainInfos()
}

func initOffchainInfos() {
	candidateFile := "config/candidate.csv"
	b, err := ioutil.ReadFile(candidateFile)
	if err != nil {
		log.Printf("Reading candidate file failed. err=%v", err)
		return
	}
	var cands []*CandidateOffchainInfo
	err = csvutil.Unmarshal(b, &cands)
	if err != nil {
		log.Printf("Unmarshal csv failed. err=%v", err)
		return
	}
	for _, c := range cands {
		if c.SocialMediaRaw != "" {
			err := json.Unmarshal([]byte(c.SocialMediaRaw), c.SocialMedia)
			if err != nil {
				log.Printf("Json decode %s failed. err=%v", c.SocialMediaRaw, err)
			}
		}
		offchainInfos[c.MainnetAccount] = c
	}
	log.Printf("%+v", offchainInfos)
}

type Candidate struct {
	MainnetAccount string  `json:"mainnet_account"`
	IsProducer     bool    `json:"is_producer"`
	NetID          string  `json:"net_id"`
	Pubkey         string  `json:"pubkey"`
	Status         string  `json:"status"`
	Online         bool    `json:"online"`
	Votes          float64 `json:"votes"`

	*CandidateOffchainInfo
}

type CandidateOffchainInfo struct {
	Name           string            `json:"name" csv:"name"`
	NameEN         string            `json:"name_en" csv:"name_en"`
	Logo           string            `json:"logo" csv:"logo"`
	Homepage       string            `json:"homepage" csv:"team_page"`
	Location       string            `json:"location" csv:"location"`
	LocationEN     string            `json:"location_en" csv:"location_en"`
	Type           string            `json:"type" csv:"type"`
	TypeEN         string            `json:"type_en" csv:"type_en"`
	Statement      string            `json:"statement" csv:"statement"`
	StatementEN    string            `json:"statement_en" csv:"statement_en"`
	Description    string            `json:"description" csv:"description"`
	DescriptionEN  string            `json:"description_en" csv:"description_en"`
	SocialMediaRaw string            `json:"-" csv:"social_media"`
	SocialMedia    map[string]string `json:"social_media"`
	MainnetAccount string            `json:"-" csv:"mainnet_account"`
}

func GetCandidates(page, size int) ([]*Candidate, error) {
	start := (page - 1) * size
	cands, err := db.GetCandidates(start, size)
	if err != nil {
		return nil, err
	}

	ret := make([]*Candidate, 0, len(cands))
	for _, cand := range cands {
		c := &Candidate{
			MainnetAccount: cand.Name,
			IsProducer:     cand.CandidateInfo.IsProducer,
			NetID:          cand.CandidateInfo.NetId,
			Pubkey:         cand.CandidateInfo.Pubkey,
			Status:         cand.CandidateInfo.Status,
			Online:         cand.CandidateInfo.Online,
			Votes:          cand.CandidateInfo.Votes,
			CandidateOffchainInfo: emptyOffchainInfo,
		}
		off := offchainInfos[cand.Name]
		if off != nil {
			c.CandidateOffchainInfo = off
		}
		ret = append(ret, c)
	}
	return ret, nil
}
