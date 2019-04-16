package db

import "github.com/iost-official/iost-api/model/blockchain/rpcpb"

type Candidate struct {
	Name          string                             `bson:"name" json:"name"`
	CandidateInfo *rpcpb.GetProducerVoteInfoResponse `bson:"candidateInfo" json:"candidate_info"`
}
