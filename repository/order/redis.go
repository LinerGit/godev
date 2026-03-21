package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"godev/model"

	"github.com/redis/go-redis/v9"
)

type RedisRepo struct {
	Client *redis.Client
}

func orderIDKey(id uint) string {
	return fmt.Sprintf("order:%d", id)
}

func (r *RedisRepo) Insert(ctx context.Context, order model.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return err
	}

	key := orderIDKey(order.OrderID)

	txn := r.Client.TxPipeline()

	res := txn.SetNX(ctx, key, string(data), 0)

	if err := res.Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("insert order: %w", err)
	}

	if err := txn.SAdd(ctx, "orders", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("insert order: %w", err)
	}

	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("insert order: %w", err)
	}

	return nil
}

var ErrNotExist = errors.New("order does not exist")

func (r *RedisRepo) GetByID(ctx context.Context, id uint) (model.Order, error) {
	key := orderIDKey(id)
	value, err := r.Client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return model.Order{}, ErrNotExist
	} else if err != nil {
		return model.Order{}, fmt.Errorf("get order", err)
	}

	var order model.Order
	err = json.Unmarshal([]byte(value), &order)
	if err != nil {
		return model.Order{}, fmt.Errorf("unmarshal order", err)
	}

	return order, nil
}

func (r *RedisRepo) Delete(ctx context.Context, id uint) error {
	key := orderIDKey(id)

	txn := r.Client.TxPipeline()

	err := txn.Del(ctx, key).Err()
	if errors.Is(err, redis.Nil) {
		txn.Discard()
		return ErrNotExist
	} else if err != nil {
		txn.Discard()
		return fmt.Errorf("delete order", err)
	}
	if err := txn.SRem(ctx, "orders", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("delete order", err)
	}
	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("delete order", err)
	}
	return nil
}

func (r *RedisRepo) Update(ctx context.Context, order model.Order) error {
	data, err := json.Marshal(order)

	err = r.Client.SetXX(ctx, orderIDKey(order.OrderID), string(data), 0).Err()
	if errors.Is(err, redis.Nil) {
		return ErrNotExist
	} else if err != nil {
		return fmt.Errorf("update order", err)
	}
	return nil
}

type FindAllPage struct {
	Size   uint64
	Offset uint64
}

type FindResult struct {
	Orders []model.Order
	Cursor uint64
}

func (r *RedisRepo) List(ctx context.Context, page FindAllPage) (FindResult, error) {
	res := r.Client.SScan(ctx, "orders", page.Offset, "*", int64(page.Size))
	keys, cursor, err := res.Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to get order ids")
	}
	xs, err := r.Client.MGet(ctx, keys...).Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to get orders")
	}

	orders := make([]model.Order, 0, len(xs))
	for i, x := range xs {
		if x == nil {
			continue
		}
		var order model.Order
		err = json.Unmarshal([]byte(xs[i].(string)), &order)
		if err != nil {
			return FindResult{}, fmt.Errorf("failed to unmarshal order")
		}
		orders[i] = order

	}
	return FindResult{
		Orders: orders,
		Cursor: cursor,
	}, nil

}
