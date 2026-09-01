package safe_socket

import "io"

type Socket interface {
	Send(data []byte) error
	Recv(n int) ([]byte, error)
}
type safeSocket struct {
	socket io.ReadWriter
}

func CrearSafeSocket(conn io.ReadWriter) Socket {
	return &safeSocket{socket: conn}
}

func (s *safeSocket) Recv(n int) ([]byte, error) {
	return RecvAll(s.socket, n)
}

func (s *safeSocket) Send(data []byte) error {
	return SendAll(s.socket, data)
}