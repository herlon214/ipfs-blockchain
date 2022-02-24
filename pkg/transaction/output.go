package transaction

import "fmt"

type Output struct {
	Value  int
	PubKey string
}

func (out *Output) String() string {
	return fmt.Sprintf("Output: %d to %s", out.Value, out.PubKey)
}

func (out *Output) CanBeUnlocked(data string) bool {
	return out.PubKey == data
}
