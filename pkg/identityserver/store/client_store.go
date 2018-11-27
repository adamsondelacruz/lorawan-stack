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

// GetClientStore returns an ClientStore on the given db (or transaction).
func GetClientStore(db *gorm.DB) ClientStore {
	return &clientStore{db: db}
}

type clientStore struct {
	db *gorm.DB
}

// selectClientFields selects relevant fields (based on fieldMask) and preloads details if needed.
func selectClientFields(query *gorm.DB, fieldMask *types.FieldMask) *gorm.DB {
	if fieldMask == nil || len(fieldMask.Paths) == 0 {
		return query
	}
	var clientColumns []string
	for _, path := range fieldMask.Paths {
		if column, ok := clientColumnNames[path]; ok {
			clientColumns = append(clientColumns, column)
		} else {
			clientColumns = append(clientColumns, path)
		}
	}
	return query.Select(append(append(modelColumns, "client_id"), clientColumns...)) // TODO: remove possible duplicate client_id
}

func (s *clientStore) CreateClient(ctx context.Context, cli *ttnpb.Client) (*ttnpb.Client, error) {
	cliModel := Client{
		ClientID: cli.ClientID, // The ID is not mutated by fromPB.
	}
	cliModel.fromPB(cli, nil)
	cliModel.SetContext(ctx)
	query := s.db.Create(&cliModel)
	if query.Error != nil {
		return nil, query.Error
	}
	var cliProto ttnpb.Client
	cliModel.toPB(&cliProto, nil)
	return &cliProto, nil
}

func (s *clientStore) FindClients(ctx context.Context, ids []*ttnpb.ClientIdentifiers, fieldMask *types.FieldMask) ([]*ttnpb.Client, error) {
	idStrings := make([]string, len(ids))
	for i, id := range ids {
		idStrings[i] = id.GetClientID()
	}
	query := s.db.Scopes(withContext(ctx), withClientID(idStrings...))
	query = selectClientFields(query, fieldMask)
	if limit, offset := limitAndOffsetFromContext(ctx); limit != 0 {
		countTotal(ctx, query.Model(&Client{}))
		query = query.Limit(limit).Offset(offset)
	}
	var cliModels []Client
	query = query.Find(&cliModels)
	setTotal(ctx, uint64(len(cliModels)))
	if query.Error != nil {
		return nil, query.Error
	}
	cliProtos := make([]*ttnpb.Client, len(cliModels))
	for i, cliModel := range cliModels {
		cliProto := new(ttnpb.Client)
		cliModel.toPB(cliProto, nil)
		cliProtos[i] = cliProto
	}
	return cliProtos, nil
}

func (s *clientStore) GetClient(ctx context.Context, id *ttnpb.ClientIdentifiers, fieldMask *types.FieldMask) (*ttnpb.Client, error) {
	query := s.db.Scopes(withContext(ctx), withClientID(id.GetClientID()))
	query = selectClientFields(query, fieldMask)
	var cliModel Client
	err := query.First(&cliModel).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errNotFoundForID(id.EntityIdentifiers())
		}
		return nil, err
	}
	cliProto := new(ttnpb.Client)
	cliModel.toPB(cliProto, nil)
	return cliProto, nil
}

func (s *clientStore) UpdateClient(ctx context.Context, cli *ttnpb.Client, fieldMask *types.FieldMask) (updated *ttnpb.Client, err error) {
	query := s.db.Scopes(withContext(ctx), withClientID(cli.GetClientID()))
	query = selectClientFields(query, fieldMask)
	var cliModel Client
	err = query.First(&cliModel).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errNotFoundForID(cli.ClientIdentifiers.EntityIdentifiers())
		}
		return nil, err
	}
	if !cli.UpdatedAt.IsZero() && cli.UpdatedAt != cliModel.UpdatedAt {
		return nil, errConcurrentWrite
	}
	if err := ctx.Err(); err != nil { // Early exit if context canceled
		return nil, err
	}
	columns := cliModel.fromPB(cli, fieldMask)
	if len(columns) > 0 {
		query = s.db.Model(&cliModel).Select(columns).Updates(&cliModel)
		if query.Error != nil {
			return nil, query.Error
		}
	}
	updated = new(ttnpb.Client)
	cliModel.toPB(updated, nil)
	return updated, nil
}

func (s *clientStore) DeleteClient(ctx context.Context, id *ttnpb.ClientIdentifiers) error {
	return deleteEntity(ctx, s.db, id.EntityIdentifiers())
}