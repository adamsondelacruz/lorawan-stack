// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package store

import (
	"context"

	"github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// GetGatewayStore returns an GatewayStore on the given db (or transaction).
func GetGatewayStore(db *gorm.DB) GatewayStore {
	return &gatewayStore{db: db}
}

type gatewayStore struct {
	db *gorm.DB
}

// selectGatewayFields selects relevant fields (based on fieldMask) and preloads details if needed.
func selectGatewayFields(query *gorm.DB, fieldMask *types.FieldMask) *gorm.DB {
	if fieldMask == nil || len(fieldMask.Paths) == 0 {
		return query
	}
	var gatewayColumns []string
	for _, path := range fieldMask.Paths {
		if column, ok := gatewayColumnNames[path]; ok {
			gatewayColumns = append(gatewayColumns, column)
		} else {
			gatewayColumns = append(gatewayColumns, path)
		}
	}
	return query.Select(append(append(modelColumns, "gateway_id"), gatewayColumns...)) // TODO: remove possible duplicate gateway_id
}

func (s *gatewayStore) CreateGateway(ctx context.Context, gtw *ttnpb.Gateway) (*ttnpb.Gateway, error) {
	gtwModel := Gateway{
		GatewayID: gtw.GatewayID, // The ID is not mutated by fromPB.
	}
	gtwModel.fromPB(gtw, nil)
	gtwModel.SetContext(ctx)
	query := s.db.Create(&gtwModel)
	if query.Error != nil {
		return nil, query.Error
	}
	var gtwProto ttnpb.Gateway
	gtwModel.toPB(&gtwProto, nil)
	return &gtwProto, nil
}

func (s *gatewayStore) FindGateways(ctx context.Context, ids []*ttnpb.GatewayIdentifiers, fieldMask *types.FieldMask) ([]*ttnpb.Gateway, error) {
	idStrings := make([]string, len(ids))
	for i, id := range ids {
		idStrings[i] = id.GetGatewayID()
	}
	query := s.db.Scopes(withContext(ctx), withGatewayID(idStrings...))
	query = selectGatewayFields(query, fieldMask)
	if limit, offset := limitAndOffsetFromContext(ctx); limit != 0 {
		countTotal(ctx, query.Model(&Gateway{}))
		query = query.Limit(limit).Offset(offset)
	}
	var gtwModels []Gateway
	query = query.Find(&gtwModels)
	setTotal(ctx, uint64(len(gtwModels)))
	if query.Error != nil {
		return nil, query.Error
	}
	gtwProtos := make([]*ttnpb.Gateway, len(gtwModels))
	for i, gtwModel := range gtwModels {
		gtwProto := new(ttnpb.Gateway)
		gtwModel.toPB(gtwProto, nil)
		gtwProtos[i] = gtwProto
	}
	return gtwProtos, nil
}

func (s *gatewayStore) GetGateway(ctx context.Context, id *ttnpb.GatewayIdentifiers, fieldMask *types.FieldMask) (*ttnpb.Gateway, error) {
	query := s.db.Scopes(withContext(ctx), withGatewayID(id.GetGatewayID()))
	if id.EUI != nil {
		query = query.Scopes(withGatewayEUI(EUI64(*id.EUI)))
	}
	query = selectGatewayFields(query, fieldMask)
	var gtwModel Gateway
	err := query.First(&gtwModel).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errNotFoundForID(id.EntityIdentifiers())
		}
		return nil, err
	}
	gtwProto := new(ttnpb.Gateway)
	gtwModel.toPB(gtwProto, nil)
	return gtwProto, nil
}

func (s *gatewayStore) UpdateGateway(ctx context.Context, gtw *ttnpb.Gateway, fieldMask *types.FieldMask) (updated *ttnpb.Gateway, err error) {
	query := s.db.Scopes(withContext(ctx), withGatewayID(gtw.GetGatewayID()))
	query = selectGatewayFields(query, fieldMask)
	var gtwModel Gateway
	err = query.First(&gtwModel).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errNotFoundForID(gtw.GatewayIdentifiers.EntityIdentifiers())
		}
		return nil, err
	}
	if !gtw.UpdatedAt.IsZero() && gtw.UpdatedAt != gtwModel.UpdatedAt {
		return nil, errConcurrentWrite
	}
	if err := ctx.Err(); err != nil { // Early exit if context canceled
		return nil, err
	}
	columns := gtwModel.fromPB(gtw, fieldMask)
	if len(columns) > 0 {
		query = s.db.Model(&gtwModel).Select(columns).Updates(&gtwModel)
		if query.Error != nil {
			return nil, query.Error
		}
	}
	updated = new(ttnpb.Gateway)
	gtwModel.toPB(updated, nil)
	return updated, nil
}

func (s *gatewayStore) DeleteGateway(ctx context.Context, id *ttnpb.GatewayIdentifiers) error {
	return deleteEntity(ctx, s.db, id.EntityIdentifiers())
}