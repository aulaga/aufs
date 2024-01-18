package internal

import (
	"fmt"
	aufs "github.com/aulaga/aufs/src"
	"io/fs"
)

type EventFile struct {
	file       aufs.File
	path       string
	propagator *EventPropagator
	changed    bool
}

func (e *EventFile) Path() string {
	return e.file.Path()
}

func (e *EventFile) Storage() aufs.Storage {
	return e.file.Storage()
}

func (e *EventFile) Close() error {
	if e.changed {
		e.propagator.AddEvent(ChangedEvent(e.path))
	}

	return e.file.Close()
}

func (e *EventFile) Read(p []byte) (n int, err error) {
	return e.file.Read(p)
}

func (e *EventFile) Seek(offset int64, whence int) (int64, error) {
	return e.file.Seek(offset, whence)
}

func (e *EventFile) Readdir(count int) ([]fs.FileInfo, error) {
	return e.file.Readdir(count)
}

func (e *EventFile) Stat() (fs.FileInfo, error) {
	return e.file.Stat()
}

func (e *EventFile) Write(p []byte) (n int, err error) {
	if len(p) > 0 {
		e.changed = true
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
	events    []Event
}

func (e *EventPropagator) Publish() {
	fmt.Println("Propagator publishing events...", len(e.events))
	for _, event := range e.events {
		for _, listener := range e.listeners {
			event.Publish(listener)
		}
	}

	e.events = nil
}

func (e *EventPropagator) AddEventListener(listener aufs.EventListener) {
	e.listeners = append(e.listeners, listener)
}

func (e *EventPropagator) AddEvent(event Event) {
	e.events = append(e.events, event)
}
