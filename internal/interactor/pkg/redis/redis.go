package redis

//import (
//	"context"
//	"errors"
//	"time"
//
//	"github.com/go-redis/redis/v9"
//)
//
//const (
//	String = "String"
//	Hash   = "Hash"
//	List   = "List"
//	Set    = "Set"
//	Sorted = "Sorted"
//)
//
//type Config struct {
//	// redis address
//	Address *string
//	// redis port
//	Port *string
//	// redis user
//	Username *string
//	// redis password
//	Password *string
//	// redis use default DB
//	DB *int64
//}
//
//type DB interface {
//	// Create data to redis
//	Create(choose, key string, input []byte, ttl time.Duration) (err error)
//	// First is get data to redis
//	First(choose, key string) (output []byte, err error)
//}
//
//type db struct {
//	// Context
//	ctx context.Context
//	// redis database
//	redisClient *redis.Client
//}
//
//func (c *Config) Connect() (DB, error) {
//	ctx := context.Background()
//	redisConfig := &redis.Options{}
//	if c.Address != nil && c.Port != nil {
//		redisConfig.Addr = *c.Address + ":" + *c.Port
//	}
//
//	if c.DB != nil {
//		redisConfig.DB = int(*c.DB)
//	}
//
//	if c.Username != nil {
//		redisConfig.Username = *c.Username
//	}
//
//	if c.Username != nil {
//		redisConfig.Password = *c.Password
//	}
//
//	redisClient := redis.NewClient(redisConfig)
//	if redisClient == nil {
//		return nil, errors.New("redis connect error")
//	}
//
//	return &db{
//		redisClient: redisClient,
//		ctx:         ctx,
//	}, nil
//}
//
//func (d *db) Create(choose, key string, input []byte, ttl time.Duration) (err error) {
//	switch choose {
//	case String:
//		err = d.redisClient.Set(d.ctx, key, input, ttl).Err()
//	default:
//		return errors.New("no option")
//	}
//
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
//
//func (d *db) First(choose, key string) (output []byte, err error) {
//	switch choose {
//	case String:
//		err = d.redisClient.Get(d.ctx, key).Scan(&output)
//	default:
//		return nil, errors.New("no option")
//	}
//
//	if err != nil {
//		return nil, err
//	}
//
//	return output, nil
//}
