/*
简单封装gorm，目的是统一配置方式，支持多数据源配置，示例如下：

	  gorm:
		db0:
		  dbType: mysql
		  dsn: root:root@tcp(127.0.0.1:3307)/sakila?charset=utf8mb4&parseTime=True&loc=Local
		  connMaxIdleTime: 10m # Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
		  connMaxLifetime: 20m
		  maxIdleConns: 5
		  maxOpenConns: 20
		  logger:
		    logMode: info
		    ignoreErrRecordNotFound: true
		    slowThresholdMS: 200
*/
package gormstarter
