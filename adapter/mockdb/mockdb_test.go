// Copyright (c) 2012-today The upper.io/db authors. All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
// OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
// WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package mockdb

import (
	"database/sql"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/upper/db/v4"
	"github.com/upper/db/v4/internal/sqladapter"
	"testing"
	"time"
)

type mockItem struct {
	collection string

	ID        int64      `db:"id"`
	Title     string     `db:"title"`
	CreatedAt *time.Time `db:"created_at"`
}

func (i *mockItem) Store(sess db.Session) db.Store {
	return sess.Collection(i.collection)
}

func TestMockDatabase(t *testing.T) {
	settings := ConnectionURL{}

	sess, err := Open(settings)
	assert.NoError(t, err)
	assert.NotNil(t, sess.(db.Session))
	assert.NotNil(t, sess.(sqladapter.Session))

	{
		connURL := sess.ConnectionURL()
		assert.Equal(t, "mockdb://mockdb", connURL.String())
	}

	{
		name := sess.Name()
		assert.Equal(t, "mockdb", name)
		sess.Name()
	}

	{
		collections, err := sess.Collections()
		assert.NoError(t, err)
		assert.Len(t, collections, 0)
	}

	{
		items := sess.Collection("items")
		assert.NotNil(t, items)

		ok, err := items.Exists()
		assert.False(t, ok)
		assert.True(t, errors.Is(err, db.ErrCollectionDoesNotExist))

	}

	{
		Mock(sess).Collection("items")

		collections, err := sess.Collections()
		assert.NoError(t, err)
		assert.Len(t, collections, 1)

		assert.Equal(t, "items", collections[0].Name())
	}

	{
		item := mockItem{
			collection: "test_items",
		}
		err := sess.Save(&item)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, db.ErrCollectionDoesNotExist), err)
	}

	{
		item := mockItem{
			collection: "test_items",
		}

		assert.Zero(t, item.ID)
		assert.Zero(t, item.CreatedAt)

		Mock(sess).
			Collection("test_items").
			PrimaryKeys([]string{"id"}).
			Insert(func(record interface{}) (interface{}, error) {
				now := time.Now()
				item := record.(*mockItem)
				item.ID = 1
				item.CreatedAt = &now
				return item.ID, nil
			})

		err := sess.Save(&item)
		assert.NoError(t, err)

		assert.NotZero(t, item.ID)
		assert.NotZero(t, item.CreatedAt)
	}

	{
		item := mockItem{
			collection: "test_items",
		}

		assert.Zero(t, item.ID)
		assert.Zero(t, item.CreatedAt)

		errFailed := errors.New("failed")

		Mock(sess).
			Collection("test_items").
			PrimaryKeys([]string{"id"}).
			Insert(func(record interface{}) (interface{}, error) {
				return nil, errFailed
			})

		err := sess.Save(&item)
		assert.Error(t, err)
		assert.True(t, errors.Is(errFailed, err))

		assert.Zero(t, item.ID)
		assert.Zero(t, item.CreatedAt)
	}

	return

	{
		err := sess.Ping()
		assert.NoError(t, err)
	}

	{
		sess.Reset()
	}

	{
		driver := sess.Driver()
		assert.NotNil(t, driver)
		assert.NotNil(t, driver.(*sql.DB))
	}
}
