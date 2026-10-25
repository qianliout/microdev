package store

import (
	"context"
	"fmt"

	"microdev/infra/consts"
	"microdev/infra/model/filter"
	promptmodel "microdev/infra/model/prompt"
	"microdev/infra/rdb"
)

type SessionDao struct {
	db *rdb.RDB
}

func NewSessionDao(db *rdb.RDB) *SessionDao {
	return &SessionDao{db: db}
}

func (dal *SessionDao) CreateSession(ctx context.Context, data *promptmodel.Session) error {
	data = data.Serialize()
	if err := data.Check(); err != nil {
		return err
	}
	cancelCtx, cancelFunc := context.WithTimeout(ctx, consts.DBCreateTimeout)
	defer cancelFunc()

	db := dal.db.Get().WithContext(cancelCtx).Table(data.TableName())
	if err := db.Create(data).Error; err != nil {
		return err
	}
	return nil
}

func (dal *SessionDao) SearchSession(
	ctx context.Context,
	param *promptmodel.SearchSessionParam,
) ([]*promptmodel.Session, int64, error) {
	result := make([]*promptmodel.Session, 0)
	param = param.Serialize()
	if err := param.Check(); err != nil {
		return result, 0, err
	}
	cancelCtx, cancelFunc := context.WithTimeout(ctx, consts.DBQueryTimeout)
	defer cancelFunc()

	tb := (&promptmodel.Session{}).TableName()
	db := dal.db.Get().WithContext(cancelCtx).Table(tb)
	if param.ID > 0 {
		db = db.Where("id = ?", param.ID)
	}
	if param.UpdatedBefore > 0 {
		db = db.Where("updated_at < ?", param.UpdatedBefore)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return result, 0, err
	}
	db = filter.AddFilter(db, param.Filter)
	if err := db.Find(&result).Error; err != nil {
		return result, 0, err
	}
	for _, row := range result {
		row = row.Deserialize()
	}
	return result, total, nil
}

func (dal *SessionDao) UpdateSession(ctx context.Context, id int64, updater map[string]any) error {
	if id <= 0 {
		return fmt.Errorf("id must be positive")
	}
	if len(updater) == 0 {
		return fmt.Errorf("updater is empty")
	}
	cancelCtx, cancelFunc := context.WithTimeout(ctx, consts.DBQueryTimeout)
	defer cancelFunc()

	db := dal.db.Get().WithContext(cancelCtx).Table((&promptmodel.Session{}).TableName())
	db = db.Where("id = ?", id)
	if err := db.Updates(updater).Error; err != nil {
		return err
	}
	return nil
}

func (dal *SessionDao) DeleteSession(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("id must be positive")
	}
	cancelCtx, cancelFunc := context.WithTimeout(ctx, consts.DBQueryTimeout)
	defer cancelFunc()

	db := dal.db.Get().WithContext(cancelCtx).Table((&promptmodel.Session{}).TableName())
	db = db.Where("id = ?", id)
	if err := db.Delete(&promptmodel.Session{}).Error; err != nil {
		return err
	}
	return nil
}
