package game

type PriorityPos struct {
	Pos
	Priority int
}

type pqueue []priorityPos

func (pq pqueue) push(pos Pos, priority int) pqueue {
	newNode := priorityPos{pos, priority}
	pq = append(pq, newNode)
	newNodeIndex := len(pq) - 1
	parentIndex, parent := pq.parent(newNodeIndex)

	for newNode.priority < parent.priority && parentIndex != 0 {
		pq.swap(newNodeIndex, parentIndex)
		newNodeIndex = parentIndex
		parentIndex, parent = pq.parent(newNodeIndex)

	}

	return pq
}

func (pq pqueue) pop() (pqueue, Pos) {
	result := pq[0].Pos
	pq[0] = pq[len(pq)-1]
	pq = pq[:len(pq)-1]

	if len(pq) == 0 {
		return pq, result
	}

	index := 0
	node := pq[index]

	leftChildExists, leftIndex, leftChild := pq.leftChild(index)
	rightChildExists, rightIndex, rightChild := pq.rightChild(index)

	for leftChildExists && node.priority > leftChild.priority ||
		rightChildExists && node.priority > rightChild.priority {

		if !rightChildExists || leftChild.priority <= rightChild.priority {
			pq.swap(index, leftIndex)
			index = leftIndex
		} else {
			pq.swap(index, rightIndex)
			index = rightIndex
		}

		leftChildExists, leftIndex, leftChild = pq.leftChild(index)
		rightChildExists, rightIndex, rightChild = pq.rightChild(index)
	}

	return pq, result
}

func (pq pqueue) swap(i, j int) {
	tmp := pq[i]
	pq[i] = pq[j]
	pq[j] = tmp

}

func (pq pqueue) parent(i int) (int, priorityPos) {
	index := (i-1)/2
	return index, pq[index]
}

func (pq pqueue) leftChild(i int) (bool, int, priorityPos) {
	index := i*2+1
	if index < len(pq) {
		return true, index, pq[index]
	}
	return false, 0, priorityPos{}
}

func (pq pqueue) rightChild(i int) (bool, int, priorityPos) {
	index := i*2+2
	if index < len(pq) {
		return true, index, pq[index]
	}
	return false, 0, priorityPos{}
}