// 配置使用zerolog，配置内容形式如下：
/*
  zerolog:
    globalLevel: info # trace, debug, info, warn, error, fatal, panic
    durationFieldUnit: ms # default is milliseconds, "ms", "ns", "us" is applicable
    # when rotation is enabled, logs will be written to files instead of stdout
    enableRotation: false # shall log rotation be enabled? default is false
      # see https://github.com/natefinch/lumberjack for details
      filename: "/var/log/foo/server.log"
      maxsize: 100 # unit: megabytes, default is 100
      maxage: 30 # unit: day, default is 0, never deleted
      maxbackups: 0
      localtime: false # default is false, use UTC time
      compress: false # default is false
*/
package zerologstarter
