// Copyright 2016 DeepFabric, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package raftstore

import (
	"github.com/deepfabric/elasticell/pkg/pb/raftcmdpb"
	"github.com/deepfabric/elasticell/pkg/pool"
	"github.com/deepfabric/elasticell/pkg/redis"
	"github.com/deepfabric/elasticell/pkg/util"
)

func (s *Store) execLIndex(id uint64, req *raftcmdpb.Request) *raftcmdpb.Response {
	cmd := redis.Command(req.Cmd)
	args := cmd.Args()

	if len(args) != 2 {
		rsp := pool.AcquireResponse()
		rsp.ErrorResult = redis.ErrInvalidCommandResp

		return rsp
	}

	index, err := util.StrInt64(args[1])
	if err != nil {
		rsp := pool.AcquireResponse()
		rsp.ErrorResult = util.StringToSlice(err.Error())
		return rsp
	}

	value, err := s.getListEngine(id).LIndex(args[0], index)
	if err != nil {
		rsp := pool.AcquireResponse()
		rsp.ErrorResult = util.StringToSlice(err.Error())
		return rsp
	}

	rsp := pool.AcquireResponse()
	rsp.BulkResult = value
	rsp.HasEmptyBulkResult = len(value) == 0
	return rsp
}

func (s *Store) execLLEN(id uint64, req *raftcmdpb.Request) *raftcmdpb.Response {
	cmd := redis.Command(req.Cmd)
	args := cmd.Args()

	if len(args) != 1 {
		rsp := pool.AcquireResponse()
		rsp.ErrorResult = redis.ErrInvalidCommandResp

		return rsp
	}

	value, err := s.getListEngine(id).LLen(args[0])
	if err != nil {
		rsp := pool.AcquireResponse()
		rsp.ErrorResult = util.StringToSlice(err.Error())
		return rsp
	}

	rsp := pool.AcquireResponse()
	rsp.IntegerResult = &value
	return rsp
}

func (s *Store) execLRange(id uint64, req *raftcmdpb.Request) *raftcmdpb.Response {
	cmd := redis.Command(req.Cmd)
	args := cmd.Args()

	if len(args) != 3 {
		rsp := pool.AcquireResponse()
		rsp.ErrorResult = redis.ErrInvalidCommandResp

		return rsp
	}

	start, err := util.StrInt64(args[1])
	if err != nil {
		rsp := pool.AcquireResponse()
		rsp.ErrorResult = util.StringToSlice(err.Error())
		return rsp
	}

	end, err := util.StrInt64(args[2])
	if err != nil {
		rsp := pool.AcquireResponse()
		rsp.ErrorResult = util.StringToSlice(err.Error())
		return rsp
	}

	value, err := s.getListEngine(id).LRange(args[0], start, end)
	if err != nil {
		rsp := pool.AcquireResponse()
		rsp.ErrorResult = util.StringToSlice(err.Error())
		return rsp
	}

	rsp := pool.AcquireResponse()
	rsp.SliceArrayResult = value
	rsp.HasEmptySliceArrayResult = len(value) == 0
	return rsp
}

func (s *Store) execLInsert(ctx *applyContext, req *raftcmdpb.Request) *raftcmdpb.Response {
	cmd := redis.Command(req.Cmd)
	args := cmd.Args()

	if len(args) != 4 {
		rsp := pool.AcquireResponse()
		rsp.ErrorResult = redis.ErrInvalidCommandResp

		return rsp
	}

	pos, err := util.StrInt64(args[1])
	if err != nil {
		rsp := pool.AcquireResponse()
		rsp.ErrorResult = util.StringToSlice(err.Error())
		return rsp
	}

	value, err := s.getListEngine(ctx.req.Header.CellId).LInsert(args[0], int(pos), args[2], args[3])
	if err != nil {
		rsp := pool.AcquireResponse()
		rsp.ErrorResult = util.StringToSlice(err.Error())
		return rsp
	}

	if value > 0 {
		size := uint64(len(args[3]))

		if value == 0 {
			size += uint64(len(args[0]))
			ctx.metrics.writtenKeys++
		}

		ctx.metrics.writtenBytes += size
		ctx.metrics.sizeDiffHint += size
	}

	rsp := pool.AcquireResponse()
	rsp.IntegerResult = &value
	return rsp
}

func (s *Store) execLPop(ctx *applyContext, req *raftcmdpb.Request) *raftcmdpb.Response {
	cmd := redis.Command(req.Cmd)
	args := cmd.Args()

	if len(args) != 1 {
		rsp := pool.AcquireResponse()
		rsp.ErrorResult = redis.ErrInvalidCommandResp

		return rsp
	}

	value, err := s.getListEngine(ctx.req.Header.CellId).LPop(args[0])
	if err != nil {
		rsp := pool.AcquireResponse()
		rsp.ErrorResult = util.StringToSlice(err.Error())
		return rsp
	}

	size := uint64(len(value))
	ctx.metrics.sizeDiffHint -= size

	rsp := pool.AcquireResponse()
	rsp.BulkResult = value
	rsp.HasEmptyBulkResult = len(value) == 0
	return rsp
}

func (s *Store) execLPush(ctx *applyContext, req *raftcmdpb.Request) *raftcmdpb.Response {
	cmd := redis.Command(req.Cmd)
	args := cmd.Args()

	if len(args) < 2 {
		rsp := pool.AcquireResponse()
		rsp.ErrorResult = redis.ErrInvalidCommandResp

		return rsp
	}

	value, err := s.getListEngine(ctx.req.Header.CellId).LPush(args[0], args[1:]...)
	if err != nil {
		rsp := pool.AcquireResponse()
		rsp.ErrorResult = util.StringToSlice(err.Error())
		return rsp
	}

	if value > 0 {
		var size uint64
		for _, arg := range args[1:] {
			size += uint64(len(arg))
		}

		if value == 1 {
			size += uint64(len(args[0]))
			ctx.metrics.writtenKeys++
		}

		ctx.metrics.writtenBytes += size
		ctx.metrics.sizeDiffHint += size
	}

	rsp := pool.AcquireResponse()
	rsp.IntegerResult = &value
	return rsp
}

func (s *Store) execLPushX(ctx *applyContext, req *raftcmdpb.Request) *raftcmdpb.Response {
	cmd := redis.Command(req.Cmd)
	args := cmd.Args()

	if len(args) != 2 {
		rsp := pool.AcquireResponse()
		rsp.ErrorResult = redis.ErrInvalidCommandResp

		return rsp
	}

	value, err := s.getListEngine(ctx.req.Header.CellId).LPushX(args[0], args[1])
	if err != nil {
		rsp := pool.AcquireResponse()
		rsp.ErrorResult = util.StringToSlice(err.Error())
		return rsp
	}

	if value > 0 {
		var size uint64

		size += uint64(len(args[1]))
		if value == 1 {
			size += uint64(len(args[0]))
			ctx.metrics.writtenKeys++
		}

		ctx.metrics.writtenBytes += size
		ctx.metrics.sizeDiffHint += size
	}

	rsp := pool.AcquireResponse()
	rsp.IntegerResult = &value
	return rsp
}

func (s *Store) execLRem(ctx *applyContext, req *raftcmdpb.Request) *raftcmdpb.Response {
	cmd := redis.Command(req.Cmd)
	args := cmd.Args()

	if len(args) != 3 {
		rsp := pool.AcquireResponse()
		rsp.ErrorResult = redis.ErrInvalidCommandResp

		return rsp
	}

	count, err := util.StrInt64(args[1])
	if err != nil {
		rsp := pool.AcquireResponse()
		rsp.ErrorResult = util.StringToSlice(err.Error())
		return rsp
	}

	value, err := s.getListEngine(ctx.req.Header.CellId).LRem(args[0], count, args[2])
	if err != nil {
		rsp := pool.AcquireResponse()
		rsp.ErrorResult = util.StringToSlice(err.Error())
		return rsp
	}

	if value > 0 {
		size := uint64(len(args[2])) * uint64(value)
		ctx.metrics.sizeDiffHint += size
	}

	rsp := pool.AcquireResponse()
	rsp.IntegerResult = &value
	return rsp
}

func (s *Store) execLSet(ctx *applyContext, req *raftcmdpb.Request) *raftcmdpb.Response {
	cmd := redis.Command(req.Cmd)
	args := cmd.Args()

	if len(args) != 3 {
		rsp := pool.AcquireResponse()
		rsp.ErrorResult = redis.ErrInvalidCommandResp

		return rsp
	}

	index, err := util.StrInt64(args[1])
	if err != nil {
		rsp := pool.AcquireResponse()
		rsp.ErrorResult = util.StringToSlice(err.Error())
		return rsp
	}

	err = s.getListEngine(ctx.req.Header.CellId).LSet(args[0], index, args[2])
	if err != nil {
		rsp := pool.AcquireResponse()
		rsp.ErrorResult = util.StringToSlice(err.Error())
		return rsp
	}

	rsp := pool.AcquireResponse()
	rsp.StatusResult = redis.OKStatusResp

	return rsp
}

func (s *Store) execLTrim(ctx *applyContext, req *raftcmdpb.Request) *raftcmdpb.Response {
	cmd := redis.Command(req.Cmd)
	args := cmd.Args()

	if len(args) != 3 {
		rsp := pool.AcquireResponse()
		rsp.ErrorResult = redis.ErrInvalidCommandResp

		return rsp
	}

	begin, err := util.StrInt64(args[1])
	if err != nil {
		rsp := pool.AcquireResponse()
		rsp.ErrorResult = util.StringToSlice(err.Error())
		return rsp
	}

	end, err := util.StrInt64(args[2])
	if err != nil {
		rsp := pool.AcquireResponse()
		rsp.ErrorResult = util.StringToSlice(err.Error())
		return rsp
	}

	err = s.getListEngine(ctx.req.Header.CellId).LTrim(args[0], begin, end)
	if err != nil {
		rsp := pool.AcquireResponse()
		rsp.ErrorResult = util.StringToSlice(err.Error())
		return rsp
	}

	rsp := pool.AcquireResponse()
	rsp.StatusResult = redis.OKStatusResp

	return rsp
}

func (s *Store) execRPop(ctx *applyContext, req *raftcmdpb.Request) *raftcmdpb.Response {
	cmd := redis.Command(req.Cmd)
	args := cmd.Args()

	if len(args) != 1 {
		rsp := pool.AcquireResponse()
		rsp.ErrorResult = redis.ErrInvalidCommandResp

		return rsp
	}

	value, err := s.getListEngine(ctx.req.Header.CellId).RPop(args[0])
	if err != nil {
		rsp := pool.AcquireResponse()
		rsp.ErrorResult = util.StringToSlice(err.Error())
		return rsp
	}

	size := uint64(len(value))
	ctx.metrics.sizeDiffHint -= size

	rsp := pool.AcquireResponse()
	rsp.BulkResult = value
	rsp.HasEmptyBulkResult = len(value) == 0
	return rsp
}

func (s *Store) execRPush(ctx *applyContext, req *raftcmdpb.Request) *raftcmdpb.Response {
	cmd := redis.Command(req.Cmd)
	args := cmd.Args()

	if len(args) < 2 {
		rsp := pool.AcquireResponse()
		rsp.ErrorResult = redis.ErrInvalidCommandResp

		return rsp
	}

	value, err := s.getListEngine(ctx.req.Header.CellId).RPush(args[0], args[1:]...)
	if err != nil {
		rsp := pool.AcquireResponse()
		rsp.ErrorResult = util.StringToSlice(err.Error())
		return rsp
	}

	if value > 0 {
		var size uint64
		for _, arg := range args[1:] {
			size += uint64(len(arg))
		}

		if value == 1 {
			size += uint64(len(args[0]))
			ctx.metrics.writtenKeys++
		}

		ctx.metrics.writtenBytes += size
		ctx.metrics.sizeDiffHint += size
	}

	rsp := pool.AcquireResponse()
	rsp.IntegerResult = &value
	return rsp
}

func (s *Store) execRPushX(ctx *applyContext, req *raftcmdpb.Request) *raftcmdpb.Response {
	cmd := redis.Command(req.Cmd)
	args := cmd.Args()

	if len(args) != 2 {
		rsp := pool.AcquireResponse()
		rsp.ErrorResult = redis.ErrInvalidCommandResp

		return rsp
	}

	value, err := s.getListEngine(ctx.req.Header.CellId).RPushX(args[0], args[1])
	if err != nil {
		rsp := pool.AcquireResponse()
		rsp.ErrorResult = util.StringToSlice(err.Error())
		return rsp
	}

	if value > 0 {
		var size uint64

		size += uint64(len(args[1]))
		if value == 1 {
			size += uint64(len(args[0]))
			ctx.metrics.writtenKeys++
		}

		ctx.metrics.writtenBytes += size
		ctx.metrics.sizeDiffHint += size
	}

	rsp := pool.AcquireResponse()
	rsp.IntegerResult = &value
	return rsp
}
