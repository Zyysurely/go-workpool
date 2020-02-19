package channelAndMutex


// 在设计pool（其实大部分的风格都是受启发于nsq的），采用select + channel的方式
// 但是在做benchmark时却没有得到优秀的结果，根据测试发现mutex的性能比channel高得多。。

import "testing"
func BenchmarkUseMutex(b *testing.B) {
	for n := 0; n < b.N; n++ {
		UseMutex()
	}
}
func BenchmarkUseChan(b *testing.B) {
	for n := 0; n < b.N; n++ {
		UseChan()
	}
}