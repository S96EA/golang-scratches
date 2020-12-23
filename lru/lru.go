package main

type Data struct {
	id   int64
	data interface{}
}

type ListEntry struct {
	data *Data
	next *ListEntry
	prev *ListEntry
}

type List struct {
	Head *ListEntry
	Tail *ListEntry
}

func NewList() *List {
	return &List{}
}

func (list *List) AddTail(i *Data) {
	entry := &ListEntry{
		data: i,
		next: nil,
		prev: nil,
	}
	if list.Head == nil {
		list.Head = entry
		list.Tail = entry
		return
	}
	list.Tail.next = entry
	entry.prev = list.Tail
	list.Tail = entry
}

func (list *List) RemoveTail() *Data {
	tailEntry := list.Tail
	if list.Head == list.Tail {
		list.Head = nil
		list.Tail = nil
		return tailEntry.data
	}
	list.Tail = list.Tail.prev
	list.Tail.next = nil
	tailEntry.prev = nil
	return tailEntry.data
}

func (list *List) Remove(entry *ListEntry) {
	prev := entry.prev
	next := entry.next
	if prev == nil {
		list.Head = next
	} else {
		prev.next = next
	}

	if next == nil {
		list.Tail = nil
	} else {
		next.prev = prev
	}
	entry.prev = nil
	entry.next = nil
}

func (list *List) AddHead(i *Data) {
	entry := &ListEntry{
		data: i,
		next: nil,
		prev: nil,
	}
	if list.Head == nil {
		list.Head = entry
		list.Tail = entry
		return
	}
	list.Head.prev = entry
	entry.next = list.Head
	list.Head = entry
	return
}

type Lru struct {
	list    *List
	data    map[int64]*ListEntry
	maxSize int
}

func NewLRU() *Lru {
	data := make(map[int64]*ListEntry)
	list := NewList()
	maxSize := 10
	lru := &Lru{
		list:    list,
		data:    data,
		maxSize: maxSize,
	}
	return lru
}

func (lru *Lru) GetData(id int64) *Data {
	entry, ok := lru.data[id]
	if !ok || entry == nil {
		return nil
	}
	lru.list.Remove(entry)
	lru.list.AddHead(entry.data)
	lru.data[id] = lru.list.Head
	return entry.data
}

func (lru *Lru) DelData(id int64) *Data {
	entry, ok := lru.data[id]
	if !ok || entry == nil {
		return nil
	}
	lru.list.Remove(entry)
	delete(lru.data, id)
	return entry.data
}

func (lru *Lru) AddData(i *Data) {
	id := i.id
	_, ok := lru.data[id]
	if ok {
		lru.DelData(id)
	}
	lru.list.AddHead(i)
	lru.data[id] = lru.list.Head

	if len(lru.data) > lru.maxSize {
		data := lru.list.RemoveTail()
		delete(lru.data, data.id)
	}
}

func (lru *Lru) PrintStat() {
	list := lru.list
	for iter := list.Head; iter != nil; iter = iter.next {
		print(iter.data.id, " ")
	}
	println()
}

func main() {
	lru := NewLRU()
	for i := 0; i < 120; i++ {
		lru.AddData(&Data{
			id: int64(i),
			data: nil,
		})
		lru.PrintStat()
	}

	lru.GetData(110)
	lru.PrintStat()

	lru.GetData(117)
	lru.PrintStat()

}
