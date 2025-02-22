package main

import "io"

type Device struct {
	TargetStop    string
	TargetService string
	Connection    io.ReadWriteCloser
}
