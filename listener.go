package stoppableListener

import (
	"errors"
	"net"
	"sync"
	"time"
)

type StoppableListener struct {
	*net.TCPListener          //Wrapped listener
	stop             chan int //Channel used only to indicate listener should shutdown
	mutex            sync.Mutex
}

func New(l net.Listener) (*StoppableListener, error) {
	tcpL, ok := l.(*net.TCPListener)

	if !ok {
		return nil, errors.New("Cannot wrap listener")
	}

	retval := &StoppableListener{}
	retval.TCPListener = tcpL
	retval.stop = make(chan int)

	return retval, nil
}

var StoppedError = errors.New("Listener stopped")

func (sl *StoppableListener) Accept() (net.Conn, error) {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()
	for {
		//Wait up to one second for a new connection
		sl.SetDeadline(time.Now().Add(time.Second))

		newConn, err := sl.TCPListener.Accept()

		//Check for the channel being closed
		select {
		case <-sl.stop:
			if err == nil {
				newConn.Close()
			}
			return nil, StoppedError
		default:
			//If the channel is still open, continue as normal
		}

		if err != nil {
			netErr, ok := err.(net.Error)

			//If this is a timeout, then continue to wait for
			//new connections
			if ok && netErr.Timeout() && netErr.Temporary() {
				continue
			}
		}

		return newConn, err
	}
}

func (sl *StoppableListener) AcceptTCP() (*net.TCPConn, error) {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()
	for {
		//Wait up to one second for a new connection
		sl.SetDeadline(time.Now().Add(time.Second))

		newConn, err := sl.TCPListener.AcceptTCP()

		//Check for the channel being closed
		select {
		case <-sl.stop:
			if err == nil {
				newConn.Close()
			}
			return nil, StoppedError
		default:
			//If the channel is still open, continue as normal
		}

		if err != nil {
			netErr, ok := err.(net.Error)

			//If this is a timeout, then continue to wait for
			//new connections
			if ok && netErr.Timeout() && netErr.Temporary() {
				continue
			}
		}

		return newConn, err
	}
}

func (sl *StoppableListener) Stop() {
	sl.stop <- 0
	go func() {
		sl.mutex.Lock()
	}()
}

func (sl *StoppableListener) Start() {
	go func() {
		sl.mutex.Unlock()
	}()
}
