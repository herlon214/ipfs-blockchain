package transaction

import "fmt"

type Input struct {
	ID  []byte
	Out int
	Sig string
}

func (in *Input) String() string {
	return fmt.Sprintf("Input: %x %d coins from %s", in.ID, in.Out, in.Sig)
}

func (in *Input) CanUnlock(data string) bool {
	return in.Sig == data
}
