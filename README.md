# pprofx

用于无代码入侵开启cpu监控,基于linux预留信号来开启/结束录制

## usage

### go get

```shell
go get -u github.com/luanruisong/pprofx
```

### import

```go
import (
    _ "github.com/luanruisong/pprofx"
)
```

## Single

pprofx 监控了两个linux预留用户信号

- SIGUSR1 -> golang:syscall.SIGUSR1
- SIGUSR2 -> golang:syscall.SIGUSR2


### SIGUSR1

SIGUSR1 用于手动开启/关闭录制

```shell
kill -SIGUSR1 {PID}
```

出现日志

```shell
2023-05-16 17:17:56 [pprofx] profile file created {run_path}/pprof_auto_5s_20230516171756.profile
2023-05-16 17:17:56 [pprofx] heap file created {run_path}/pprof_auto_5s_20230516171756.heap
2023-05-16 17:17:56 [pprofx] start recording
```
再次执行

```shell
kill -SIGUSR1 {PID}
```

出现日志

```shell
2023-05-16 17:18:01 [pprofx] stop recording
2023-05-16 17:18:01 [pprofx] close file connect recording
```

录制结束，接下来就可以使用golang提供的工具进行分析


### SIGUSR2

SIGUSR2 用于自动录制

```shell
kill -SIGUSR2 {PID}
```

出现日志

```shell
2023-05-16 17:17:56 [pprofx] recording -> 10m0s
2023-05-16 17:17:56 [pprofx] profile file created {run_path}/pprof_auto_5s_20230516171756.profile
2023-05-16 17:17:56 [pprofx] heap file created {run_path}/pprof_auto_5s_20230516171756.heap
2023-05-16 17:17:56 [pprofx] start recording

---> wait for 10m0s <---

2023-05-16 17:27:56 [pprofx] stop recording
2023-05-16 17:27:56 [pprofx] close file connect recording
```

## other

### 自动录制

关于自动录制，可以在代码中通过函数来设置需要录制的时间

```go
pprofx.AutoDuration(time.Second * 5)
```

