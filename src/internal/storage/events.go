package storage

import (
	"fmt"
	aufs "github.com/aulaga/cloud/src/filesystem"
	"io/fs"
)

type EventFile struct {
	file       aufs.File
	path       string
	propagator *EventPropagator
}

func (e EventFile) Path() string {
	return e.file.Path()
}

func (e EventFile) Storage() aufs.Storage {
	return e.file.Storage()
}

func (e EventFile) Close() error {
	return e.file.Close()
}

func (e EventFile) Read(p []byte) (n int, err error) {
	return e.file.Read(p)
}

func (e EventFile) Seek(offset int64, whence int) (int64, error) {
	return e.file.Seek(offset, whence)
}

func (e EventFile) Readdir(count int) ([]fs.FileInfo, error) {
	return e.file.Readdir(count)
}

func (e EventFile) Stat() (fs.FileInfo, error) {
	return e.file.Stat()
}

func (e EventFile) Write(p []byte) (n int, err error) {
	fmt.Println("[EventFile] Writing...", len(p))
	if len(p) > 0 {
		e.propagator.AddEvent(ChangedEvent(e.path))
	}
	return e.file.Write(p)
}

type Event interface {
	Publish(listener aufs.EventListener)
}

type simpleEvent struct {
	eventAction func(listener aufs.EventListener)
}

func (e *simpleEvent) Publish(listener aufs.EventListener) {
	e.eventAction(listener)
}

func MovedEvent(src string, dst string) Event {
	return &simpleEvent{eventAction: func(listener aufs.EventListener) {
		listener.Moved(src, dst)
	}}
}

func ChangedEvent(path string) Event {
	return &simpleEvent{eventAction: func(listener aufs.EventListener) {
		listener.Changed(path)
	}}
}

func DeletedEvent(path string) Event {
	return &simpleEvent{eventAction: func(listener aufs.EventListener) {
		listener.Deleted(path)
	}}
}

type EventPropagator struct {
	listeners []aufs.EventListener
	events    map[Event]bool
}

func (e *EventPropagator) Publish() {
	fmt.Println("Propagator publishing events...", len(e.events))
	for event := range e.events {
		for _, listener := range e.listeners {
			event.Publish(listener)
		}
	}

	e.events = map[Event]bool{}
}

func (e *EventPropagator) AddEventListener(listener aufs.EventListener) {
	e.listeners = append(e.listeners, listener)
}

func (e *EventPropagator) AddEvent(event Event) {
	e.events[event] = true
}
