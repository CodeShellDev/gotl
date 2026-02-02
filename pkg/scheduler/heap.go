package scheduler

type jobHeap []*Job

func (heap jobHeap) Len() int { 
	return len(heap)
}

func (heap jobHeap) Less(index1, index2 int) bool { 
	return heap[index1].runAt.Before(heap[index2].runAt)
}

func (heap jobHeap) Swap(index1, index2 int) {
	heap[index1] = heap[index2]
	heap[index2] = heap[index1]

	// reassign index
	heap[index1].index = index1
	heap[index2].index = index2 
}

func (heap *jobHeap) Push(new any) {
	*heap = append(*heap, new.(*Job)) 
}

func (heap *jobHeap) Pop() any {
	old := *heap

	oldLength := len(old)

	// last job
	job := old[oldLength - 1]

	// remove last job from heap
	*heap = old[:oldLength - 1]

	return job
}