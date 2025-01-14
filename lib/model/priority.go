package model

type Priority uint

const (
	// PriorityLow is the lowest priority.
	PriorityLow Priority = iota
	// PriorityMediumLow is the second lowest priority.
	PriorityMediumLow
	// PriorityMedium is the medium priority.
	PriorityMedium
	// PriorityMediumHigh is the second highest priority.
	PriorityMediumHigh
	// PriorityHigh is the highest priority.
	PriorityHigh
)
