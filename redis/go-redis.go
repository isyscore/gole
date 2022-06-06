package redis

import (
	goredis "github.com/go-redis/redis/v8"
	"github.com/isyscore/gole/config"
	goleTime "github.com/isyscore/gole/time"
	"time"
)

type ConfigError struct {
	ErrMsg string
}

func (error *ConfigError) Error() string {
	return error.ErrMsg
}

func init() {
	config.LoadConfig()

	if config.GetValueBoolDefault("base.redis.enable", false) {
		err := config.GetValueObject("base.redis", &config.RedisCfg)
		if err != nil {
			return
		}
	}
}

func GetClient() (goredis.UniversalClient, error) {
	if config.RedisCfg.Sentinel.Master != "" {
		return goredis.NewFailoverClient(getSentinelConfig()), nil
	} else if len(config.RedisCfg.Cluster.Addrs) != 0 {
		return goredis.NewClusterClient(getClusterConfig()), nil
	} else {
		return goredis.NewClient(getStandaloneConfig()), nil
	}
}

func getStandaloneConfig() *goredis.Options {
	addr := "127.0.0.1:6379"
	if config.RedisCfg.Standalone.Addr != "" {
		addr = config.RedisCfg.Standalone.Addr
	}

	redisConfig := &goredis.Options{
		Addr: addr,

		DB:       config.RedisCfg.Standalone.Database,
		Network:  config.RedisCfg.Standalone.Network,
		Username: config.RedisCfg.Username,
		Password: config.RedisCfg.Password,

		MaxRetries:      config.RedisCfg.MaxRetries,
		MinRetryBackoff: goleTime.NumToTimeDuration(config.RedisCfg.MinRetryBackoff, time.Millisecond),
		MaxRetryBackoff: goleTime.NumToTimeDuration(config.RedisCfg.MaxRetryBackoff, time.Millisecond),

		DialTimeout:  goleTime.NumToTimeDuration(config.RedisCfg.DialTimeout, time.Millisecond),
		ReadTimeout:  goleTime.NumToTimeDuration(config.RedisCfg.ReadTimeout, time.Millisecond),
		WriteTimeout: goleTime.NumToTimeDuration(config.RedisCfg.WriteTimeout, time.Millisecond),

		PoolFIFO:           config.RedisCfg.PoolFIFO,
		PoolSize:           config.RedisCfg.PoolSize,
		MinIdleConns:       config.RedisCfg.MinIdleConns,
		MaxConnAge:         goleTime.NumToTimeDuration(config.RedisCfg.MaxConnAge, time.Millisecond),
		PoolTimeout:        goleTime.NumToTimeDuration(config.RedisCfg.PoolTimeout, time.Millisecond),
		IdleTimeout:        goleTime.NumToTimeDuration(config.RedisCfg.IdleTimeout, time.Millisecond),
		IdleCheckFrequency: goleTime.NumToTimeDuration(config.RedisCfg.IdleCheckFrequency, time.Millisecond),
	}
	return redisConfig
}

func getSentinelConfig() *goredis.FailoverOptions {
	redisConfig := &goredis.FailoverOptions{
		SentinelAddrs: config.RedisCfg.Sentinel.Addrs,
		MasterName:    config.RedisCfg.Sentinel.Master,

		DB:               config.RedisCfg.Sentinel.Database,
		Username:         config.RedisCfg.Username,
		Password:         config.RedisCfg.Password,
		SentinelUsername: config.RedisCfg.Sentinel.SentinelUser,
		SentinelPassword: config.RedisCfg.Sentinel.SentinelPassword,

		MaxRetries:      config.RedisCfg.MaxRetries,
		MinRetryBackoff: goleTime.NumToTimeDuration(config.RedisCfg.MinRetryBackoff, time.Millisecond),
		MaxRetryBackoff: goleTime.NumToTimeDuration(config.RedisCfg.MaxRetryBackoff, time.Millisecond),

		DialTimeout:  goleTime.NumToTimeDuration(config.RedisCfg.DialTimeout, time.Millisecond),
		ReadTimeout:  goleTime.NumToTimeDuration(config.RedisCfg.ReadTimeout, time.Millisecond),
		WriteTimeout: goleTime.NumToTimeDuration(config.RedisCfg.WriteTimeout, time.Millisecond),

		PoolFIFO:           config.RedisCfg.PoolFIFO,
		PoolSize:           config.RedisCfg.PoolSize,
		MinIdleConns:       config.RedisCfg.MinIdleConns,
		MaxConnAge:         goleTime.NumToTimeDuration(config.RedisCfg.MaxConnAge, time.Millisecond),
		PoolTimeout:        goleTime.NumToTimeDuration(config.RedisCfg.PoolTimeout, time.Millisecond),
		IdleTimeout:        goleTime.NumToTimeDuration(config.RedisCfg.IdleTimeout, time.Millisecond),
		IdleCheckFrequency: goleTime.NumToTimeDuration(config.RedisCfg.IdleCheckFrequency, time.Millisecond),
	}

	return redisConfig
}

func getClusterConfig() *goredis.ClusterOptions {
	if len(config.RedisCfg.Cluster.Addrs) == 0 {
		config.RedisCfg.Cluster.Addrs = []string{"127.0.0.1:6379"}
	}

	redisConfig := &goredis.ClusterOptions{
		Addrs: config.RedisCfg.Cluster.Addrs,

		Username: config.RedisCfg.Username,
		Password: config.RedisCfg.Password,

		MaxRedirects:   config.RedisCfg.Cluster.MaxRedirects,
		ReadOnly:       config.RedisCfg.Cluster.ReadOnly,
		RouteByLatency: config.RedisCfg.Cluster.RouteByLatency,
		RouteRandomly:  config.RedisCfg.Cluster.RouteRandomly,

		MaxRetries:      config.RedisCfg.MaxRetries,
		MinRetryBackoff: goleTime.NumToTimeDuration(config.RedisCfg.MinRetryBackoff, time.Millisecond),
		MaxRetryBackoff: goleTime.NumToTimeDuration(config.RedisCfg.MaxRetryBackoff, time.Millisecond),

		DialTimeout:  goleTime.NumToTimeDuration(config.RedisCfg.DialTimeout, time.Millisecond),
		ReadTimeout:  goleTime.NumToTimeDuration(config.RedisCfg.ReadTimeout, time.Millisecond),
		WriteTimeout: goleTime.NumToTimeDuration(config.RedisCfg.WriteTimeout, time.Millisecond),
		PoolFIFO:     config.RedisCfg.PoolFIFO,
		PoolSize:     config.RedisCfg.PoolSize,
		MinIdleConns: config.RedisCfg.MinIdleConns,

		MaxConnAge:         goleTime.NumToTimeDuration(config.RedisCfg.MaxConnAge, time.Millisecond),
		PoolTimeout:        goleTime.NumToTimeDuration(config.RedisCfg.PoolTimeout, time.Millisecond),
		IdleTimeout:        goleTime.NumToTimeDuration(config.RedisCfg.IdleTimeout, time.Millisecond),
		IdleCheckFrequency: goleTime.NumToTimeDuration(config.RedisCfg.IdleCheckFrequency, time.Millisecond),
	}
	return redisConfig
}
