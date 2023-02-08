package tengologger

import "sync"

type LogInforInterface interface {
	GetName() string
	Error() error
}

//LogInfoChainBuffer 日志缓冲区,减少并发日志丢失情况
var LogInfoChainBuffer int = 50

// logInfoChain 日志传送通道，缓冲区满后,会丢弃日志
var logInfoChain = make(chan LogInforInterface, LogInfoChainBuffer)

var setLoggerWrite sync.Once

// SetLoggerWriter 设置日志输出逻辑
func SetLoggerWriter(fn func(logInfo LogInforInterface, typeName string, err error)) {
	if fn == nil {
		return
	}
	setLoggerWrite.Do(func() {
		go func() {
			defer func() {
				recover() // 此处由错误，直接丢弃，无法输出，可探讨是否可以输出到标准输出
			}()
			for logInfo := range logInfoChain {
				fn(logInfo, logInfo.GetName(), logInfo.Error())
			}
		}()
	})

}

func SendLogInfo(info LogInforInterface) {
	select { // 不阻塞写入,避免影响主程序
	case logInfoChain <- info:
		return
	default:
		return
	}
}
