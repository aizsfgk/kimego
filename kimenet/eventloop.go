package main

import (
	"fmt"
	"syscall"
)

type EventLoop struct {
	serv *Server
	EpFd int
}

type Event int

const (
	INVALID_EVNET Event = 0
	READ_EVENT    Event = 1
	WRITE_EVNET   Event = 2
)


func NewEventLoop() (el *EventLoop, err error) {
	el = new(EventLoop)
	el.EpFd, err = syscall.EpollCreate1(syscall.EPOLL_CLOEXEC)
	if err != nil {
		return
	}
	//syscall.CloseOnExec(el.EpFd)
	return
}

func (el *EventLoop) Poll(callback func(fd int, event Event) error) (err error) {

	var (
		defaultSize = 1024
		epollEvent []syscall.EpollEvent
		num int
	)

	for {
		if el.serv.State == Stop {
			fmt.Println("服务器已经停止")
			return
		}

		epollEvent = make([]syscall.EpollEvent, defaultSize)

		num, err = syscall.EpollWait(el.EpFd, epollEvent, 1000)
		if err !=nil && err != syscall.EAGAIN && err != syscall.EINTR {
			fmt.Println("EpollWait err: ", err.Error())
			continue
		}

		//fmt.Println("epollWait return num: ", num)
		for i := 0; i < num; i++ {
			fd := int(epollEvent[i].Fd)
			ev := INVALID_EVNET
			if (epollEvent[i].Events & syscall.EPOLLIN ) != 0 {
				ev |= READ_EVENT
			}

			if (epollEvent[i].Events & syscall.EPOLLRDHUP ) != 0 {
				ev |= READ_EVENT
			}

			if (epollEvent[i].Events & syscall.EPOLLOUT) != 0 {
				ev |= WRITE_EVNET
			}

			if (epollEvent[i].Events & syscall.EPOLLERR) != 0 {
				ev |= WRITE_EVNET
			}

			if (epollEvent[i].Events) & syscall.EPOLLHUP != 0 {
				ev |= WRITE_EVNET
			}
			err = callback(fd, ev)
		}

		if num == defaultSize {
			defaultSize <<= 1
		}
	}

}

func (el EventLoop) CloseFd(fd int) (err error) {
	return el.Remove(fd)
}

func (el *EventLoop) Close() (err error) {
	err = syscall.Close(el.EpFd)
	return
}

func (el *EventLoop) AddRead(fd int) (err error) {
	return el.addEvent(fd, syscall.EPOLLIN)
}

func (el *EventLoop) AddWrite(fd int) (err error) {
	return el.addEvent(fd, syscall.EPOLLOUT)
}

func (el *EventLoop) ModRead(fd int) (err error) {
	return el.modEvent(fd, syscall.EPOLLIN)
}

func (el *EventLoop) ModWrite(fd int) (err error) {
	return el.modEvent(fd, syscall.EPOLLOUT)
}

func (el *EventLoop) ModReadWrite(fd int) (err error) {
	return el.modEvent(fd, syscall.EPOLLIN|syscall.EPOLLOUT)
}

func (el *EventLoop) Remove(fd int) (err error) {
	return el.removeEvent(fd)
}

// ************* epoll wrapper *********** //
func (el *EventLoop) addEvent(fd int, events uint32) (err error) {
	err = syscall.EpollCtl(el.EpFd, syscall.EPOLL_CTL_ADD, fd, &syscall.EpollEvent{
		Events: events,
		Fd:     int32(fd),
	})
	return
}

func  (el *EventLoop) modEvent(fd int, events uint32) (err error) {
	err = syscall.EpollCtl(el.EpFd, syscall.EPOLL_CTL_MOD, fd, &syscall.EpollEvent{
		Events: events,
		Fd:     int32(fd),
	})
	return
}

func  (el *EventLoop) removeEvent(fd int) (err error) {
	err = syscall.EpollCtl(el.EpFd, syscall.EPOLL_CTL_DEL, fd, nil)
	return
}
