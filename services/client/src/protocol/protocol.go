package protocol

import (
	"encoding/binary"
	
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/lottery"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

const BIRTHDATE_LEN = 10

type ClientProtocol struct {
	socket   safe_socket.Socket
	agencyId int
}

func CrearProtocolo(socket safe_socket.Socket, agencyId int) *ClientProtocol {
	return &ClientProtocol{socket: socket, agencyId: agencyId}
}

func (p *ClientProtocol) SendBet(bet lottery.Bet) error {
	payload, err := p.serializeBet(bet)
	if err != nil {
		return err
	}
	return p.sendFramed(payload)
}

func (p *ClientProtocol) SendBatch(bets []lottery.Bet) error {
	var payload []byte 
	// agregamos al principio del batch la cantidad de bets dentro del batch
	payload = binary.BigEndian.AppendUint32(payload, uint32(len(bets)))

	for _, bet := range bets {
		serialized_bet, err := p.serializeBet(bet)
		if err != nil {
			return err
		}
		payload = append(payload, serialized_bet...)
	}

	return p.sendFramed(payload)
}

func (p *ClientProtocol) sendFramed(payload []byte) error {
	var frame []byte

	// primero el agencyId
	frame = append(frame, byte(p.agencyId))
	// luego el largo
	frame = binary.BigEndian.AppendUint32(frame, uint32(len(payload)))
	// finalmente el payload
	frame = append(frame, payload...)

	err := p.socket.Send(frame)
	if err != nil {
		return err
	}
	return nil
}

func (p *ClientProtocol) serializeBet(bet lottery.Bet) ([]byte, error) {
	var buf []byte

	// se agrega al buffer 1 byte de longitud + string para nombre
	buf = append(buf, byte(len(bet.FirstName)))
	buf = append(buf, []byte(bet.FirstName)...)

	// se agrega al buffer 1 byte de longitud + string para apellido
	buf = append(buf, byte(len(bet.LastName)))
	buf = append(buf, []byte(bet.LastName)...)

	// se agrega al buffer BirthdateLen para la fecha (tamaño fijo para fechas)
	buf = append(buf, []byte(bet.Birthdate)...)

	buf = binary.BigEndian.AppendUint32(buf, uint32(bet.Document))
	buf = binary.BigEndian.AppendUint32(buf, uint32(bet.Number))

	return buf, nil
}

func (p *ClientProtocol) SendEnd() error{
	var buf []byte
	return p.sendFramed(buf)
}

func (p *ClientProtocol) Close() error {
	return p.socket.Close()
}

func (p *ClientProtocol) RecvWinners(data []byte) ([]lottery.Bet, error) {
	return nil, nil
}
