package qln

type nodeWithDist struct {
	Dist float64
	Node int
	Amt  int64
}

// distanceHeap is a min-distance heap that's used within our path finding
// algorithm to keep track of the "closest" node to our source node.
type distanceHeap struct {
	nodes []nodeWithDist
}

// Len returns the number of nodes in the priority queue.
//
// NOTE: This is part of the heap.Interface implementation.
func (d *distanceHeap) Len() int { return len(d.nodes) }

// Less returns whether the item in the priority queue with index i should sort
// before the item with index j.
//
// NOTE: This is part of the heap.Interface implementation.
func (d *distanceHeap) Less(i, j int) bool {
	return d.nodes[i].Dist < d.nodes[j].Dist
}

// Swap swaps the nodes at the passed indices in the priority queue.
//
// NOTE: This is part of the heap.Interface implementation.
func (d *distanceHeap) Swap(i, j int) {
	d.nodes[i], d.nodes[j] = d.nodes[j], d.nodes[i]
}

// Push pushes the passed item onto the priority queue.
//
// NOTE: This is part of the heap.Interface implementation.
func (d *distanceHeap) Push(x interface{}) {
	d.nodes = append(d.nodes, x.(nodeWithDist))
}

// Pop removes the highest priority item (according to Less) from the priority
// queue and returns it.
//
// NOTE: This is part of the heap.Interface implementation.
func (d *distanceHeap) Pop() interface{} {
	n := len(d.nodes)
	x := d.nodes[n-1]
	d.nodes = d.nodes[0 : n-1]
	return x
}
