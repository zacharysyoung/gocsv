package subcmd

import (
	"testing"
)

func TestList(t *testing.T) {
	lst := List[Bar]{}

	lst.Push(1.1)
	lst.Push(2.2)
	lst.Push(3.3)

	t.Error(lst.head)
	t.Error(lst.head.next)
	t.Error(lst.head.next.next)
}

func TestValCountList(t *testing.T) {
	lst := ValCountList[Bar]{}
	lst = append(lst, ValCount[Bar]{1.1, 2})
	lst = append(lst, ValCount[Bar]{2.2, 3})
	lst = append(lst, ValCount[Bar]{3.3, 9})
	t.Error(lst)
}
