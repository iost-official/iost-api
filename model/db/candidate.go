package db

import (
	"github.com/globalsign/mgo/bson"
	"github.com/iost-official/iost-api/model/blockchain/rpcpb"
)

type Candidate struct {
	Name          string                             `bson:"name" json:"name"`
	CandidateInfo *rpcpb.GetProducerVoteInfoResponse `bson:"candidateInfo" json:"candidate_info"`
}

func GetCandidates(start, limit int) ([]*Candidate, error) {
	c := GetCollection(CollectionCandidate)
	var ret []*Candidate
	err := c.Find(bson.M{}).Sort("-candidateInfo.votes").Skip(start).Limit(limit).All(&ret)
	return ret, err
}

func GetCandidateCount() (int, error) {
	c := GetCollection(CollectionCandidate)
	return c.Find(bson.M{}).Count()
}
