package sequence

// Sequence取号器接口
type Sequence interface {
	Next() (uint64, error)
}
