package queue

type Queue struct {
	queue []int
}

func Constructor() Queue {
	return Queue{
		queue: []int{},
	}
}

func (q *Queue) Push(x int) {
	q.queue = append(q.queue, x)
}

func (q *Queue) Pop() (int, bool) {
	if q.Empty() {
		return 0, false
	}
	elem := q.queue[0]
	q.queue = q.queue[1:]
	return elem, true
}

func (q *Queue) Peek() (int, bool) {
	if q.Empty() {
		return 0, false
	}
	return q.queue[0], true
}

func (q *Queue) Empty() bool {
	return len(q.queue) == 0
}
