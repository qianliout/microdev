package store

import (
	"context"
	"fmt"

	"microdev/infra/consts"
	"microdev/infra/model/filter"
	promptmodel "microdev/infra/model/prompt"
	"microdev/infra/rdb"
)

type MessageDao struct {
	db *rdb.RDB
}

func NewMessageDao(db *rdb.RDB) *MessageDao {
	return &MessageDao{db: db}
}

func (dal *MessageDao) CreateMessage(ctx context.Context, data *promptmodel.Message) error {
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

func (dal *MessageDao) SearchMessage(
	ctx context.Context,
	param *promptmodel.SearchMessageParam,
) ([]*promptmodel.Message, int64, error) {
	result := make([]*promptmodel.Message, 0)
	param = param.Serialize()
	if err := param.Check(); err != nil {
		return result, 0, err
	}
	cancelCtx, cancelFunc := context.WithTimeout(ctx, consts.DBQueryTimeout)
	defer cancelFunc()

	tb := (&promptmodel.Message{}).TableName()
	db := dal.db.Get().WithContext(cancelCtx).Table(tb)
	db = db.Where("session_id = ?", param.SessionID)

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

func (dal *MessageDao) DeleteMessage(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("id must be positive")
	}
	cancelCtx, cancelFunc := context.WithTimeout(ctx, consts.DBQueryTimeout)
	defer cancelFunc()

	db := dal.db.Get().WithContext(cancelCtx).Table((&promptmodel.Message{}).TableName())
	db = db.Where("id = ?", id)
	if err := db.Delete(&promptmodel.Message{}).Error; err != nil {
		return err
	}
	return nil
}
