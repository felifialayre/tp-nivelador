package safe_socket

import "io"


func SendAll(socket io.Writer, bytes []byte) error {
	n := 0
	for n < len(bytes) {
		deltaN, err := socket.Write(bytes[n:])
		if err != nil {
			return err
		}
		n += deltaN
	}
	return nil
}

func RecvAll(socket io.Reader, size int) ([]byte, error) {
	buff := make([]byte, size)
	n := 0
	for n < size {
		deltaN, err := socket.Read(buff[n:])
		if err != nil && deltaN == 0 {
			return nil, err
		}
		n += deltaN
	}
	return buff[:n], nil
}
