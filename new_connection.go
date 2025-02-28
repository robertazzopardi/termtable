package main

type Action string

const (
	SUBMIT Action = "SUBMIT"
	TEST   Action = "TEST"
	CANCEL Action = "CANCEL"
)

type TestStatus string

const (
	PASSED TestStatus = "PASSED"
	FAILED TestStatus = "FAILED"
)
