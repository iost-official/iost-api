package model

import (
	"bufio"
	"log"
	"os"
	"strings"

	"github.com/iost-official/iost-api/model/db"
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
	f, err := os.Open(candidateFile)
	if err != nil {
		log.Printf("Open candidate file failed. err=%v", err)
		return
	}
	defer f.Close()

	br := bufio.NewReader(f)
	br.ReadString('\n') // skip header
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			break
		}
		arr := strings.Split(line, ",")
		if len(arr) < 14 {
			log.Printf("Invalid candidate information line: %s", line)
			continue
		}
		offchainInfos[arr[13]] = &CandidateOffchainInfo{
			Name:          arr[0],
			NameEN:        arr[1],
			Logo:          arr[2],
			Homepage:      arr[3],
			Location:      arr[4],
			Type:          arr[5],
			TypeEN:        arr[6],
			Statement:     arr[7],
			StatementEN:   arr[8],
			Description:   arr[9],
			DescriptionEN: arr[10],
			SocialMedia:   arr[11],
			SocialMediaEN: arr[12],
		}
	}

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
	Name          string `json:"name"`
	NameEN        string `json:"name_en"`
	Logo          string `json:"logo"`
	Homepage      string `json:"homepage"`
	Location      string `json:"location"`
	Type          string `json:"type"`
	TypeEN        string `json:"type_en"`
	Statement     string `json:"statement"`
	StatementEN   string `json:"statement_en"`
	Description   string `json:"description"`
	DescriptionEN string `json:"description_en"`
	SocialMedia   string `json:"social_media"`
	SocialMediaEN string `json:"social_media_en"`
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
