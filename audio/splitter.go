package audio

//
// import "sync"
//
// type sharedStream struct {
// 	lock           *sync.Mutex
// 	childPositions []int
// 	buffer         []byte
// 	stream         Stream
// }
//
// type splitStream struct {
// 	id           int
// 	sharedStream *sharedStream
// }
//
// // NewSplitter creates a new audio splitter that allows multiple
// // devices to read from simutaneously.
// func NewSplitter(stream Stream, num int) []Stream {
// 	sstream := &sharedStream{
// 		lock:           new(sync.Mutex),
// 		childPositions: make([]int, num),
// 		stream:         stream,
// 	}
//
// 	var children []Stream
//
// 	for i := 0; i < num; i++ {
// 		children = append(children, splitStream{
// 			id:           i,
// 			sharedStream: sstream,
// 		})
// 	}
//
// 	return children
// }
//
// func (s splitStream) SampleRate() int {
// 	return s.sharedStream.stream.SampleRate()
// }
//
// func (s splitStream) Read(dst interface{}) (int, error) {
//
// }
