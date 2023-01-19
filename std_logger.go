package tengolib

//LogInfoChainBuffer 日志缓冲区,减少并发日志丢失情况
var LogInfoChainBuffer int = 50

// logInfoChain 日志传送通道，缓冲区满后,会丢弃日志
var logInfoChain = make(chan LogInforInterface, LogInfoChainBuffer)

//GetLoggerChain 获取日志接收通道
func GetLoggerChain() (readChain <-chan LogInforInterface) {
	return logInfoChain
}

type LogInforInterface interface {
	GetName() string
	Error() error
}

func SendLogInfo(info LogInforInterface) {
	select { // 不阻塞写入,避免影响主程序
	case logInfoChain <- info:
		return
	default:
		return
	}

}
