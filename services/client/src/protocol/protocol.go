package protocol

import (
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/lottery"
	"encoding/binary"
)

const BIRTHDATE_LEN = 10

type servProtocol struct {
	socket safe_socket.Socket
}

func CrearProtocolo(socket safe_socket.Socket) *servProtocol {
	return &servProtocol{socket: socket}
}

func (p *servProtocol) SendBet(bet lottery.Bet) error {
	payload := p.serializeBet(bet)
	return p.sendFramed(payload) 
}

func (p *servProtocol) sendFramed(payload []byte) error {
	return nil
}

func (p *servProtocol) serializeBet(bet lottery.Bet) []byte {
	var buf []byte

	// se agrega al buffer 1 byte de longitud + string para nombre
	buf = append(buf, byte(len(bet.FirstName)))
	buf = append(buf, []byte(bet.FirstName)...)

	// se agrega al buffer 1 byte de longitud + string para apellido
	buf = append(buf, byte(len(bet.LastName)))
	buf = append(buf, []byte(bet.LastName)...)

	// se agrega al buffer BirthdateLen para la fecha (tamaño fijo para fechas)
	buf = append(buf, BIRTHDATE_LEN)

	buf = binary.BigEndian.AppendUint32(buf, uint32(bet.Document))
	buf = binary.BigEndian.AppendUint32(buf, uint32(bet.Number))

	return buf
}
