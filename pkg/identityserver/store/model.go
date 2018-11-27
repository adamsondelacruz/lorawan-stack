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
	"time"

	"github.com/jinzhu/gorm"
)

func cleanTime(t time.Time) time.Time {
	return t.UTC().Truncate(time.Millisecond)
}

func init() {
	gorm.NowFunc = func() time.Time {
		return cleanTime(time.Now())
	}
}

type modelInterface interface {
	PrimaryKey() string
	SetContext(ctx context.Context)
}

// Model is the base of database models.
type Model struct {
	ctx context.Context

	ID        string `gorm:"type:UUID;primary_key;default:gen_random_uuid()"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// PrimaryKey returns the primary key of the model.
func (m Model) PrimaryKey() string { return m.ID }

// SetContext needs to be called before creating models.
func (m *Model) SetContext(ctx context.Context) {
	m.ctx = ctx
}

var modelColumns = []string{"id", "created_at", "updated_at"}

// SoftDelete makes a Delete operation set a DeletedAt instead of actually deleting the model.
type SoftDelete struct {
	DeletedAt *time.Time `gorm:"index"`
}