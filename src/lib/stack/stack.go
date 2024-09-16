package stack

type StackError struct {
	msg string
}

func NewStackError(msg string) StackError {
	return StackError{
		msg: msg,
	}
}

func (se StackError) Error() string {
	return se.msg
}

/*
Push
Pop
Peek
*/
type Stack[T any] struct {
	elems []T
}

func NewStack[T any]() *Stack[T] {
	return &Stack[T]{elems: []T{}}
}

func (s *Stack[T]) Push(elem T) {
	s.elems = append(s.elems, elem)
}

func (s *Stack[T]) Pop() (T, error) {
	size := s.Size()
	if size < 1 {
		var empty T
		return empty, NewStackError("Attempt to Pop() an empty stack")
	}
	res := s.elems[size-1]
	s.elems = s.elems[:size-1]
	return res, nil
}

func (s *Stack[T]) Size() int {
	return len(s.elems)
}
