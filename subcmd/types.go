package subcmd

type (
	Foo int64
	Bar float64

	Baz interface{ Foo | Bar }
)

type List[T Baz] struct {
	head, tail *element[T]
}

type element[T Baz] struct {
	next *element[T]
	val  T
}

func (lst *List[T]) Push(v T) {
	if lst.tail == nil {
		lst.head = &element[T]{val: v}
		lst.tail = lst.head
	} else {
		lst.tail.next = &element[T]{val: v}
		lst.tail = lst.tail.next
	}
}

func (lst *List[T]) AllElements() []T {
	var elems []T
	for e := lst.head; e != nil; e = e.next {
		elems = append(elems, e.val)
	}
	return elems
}

type ValCount[T Baz] struct {
	val   T
	count int
}

type ValCountList[T Baz] []ValCount[T]

func (lst ValCountList[T]) Sort() {

}
