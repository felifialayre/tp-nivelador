package protocol

import (
	"encoding/binary"
	
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/lottery"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

const BIRTHDATE_LEN = 10
const UINT32_SIZE = 4
const UINT8_SIZE = 1

const BATCH = 0
const END = 1

type ClientProtocol struct {
	socket   safe_socket.Socket
	agencyId int
}

func CrearProtocolo(socket safe_socket.Socket, agencyId int) *ClientProtocol {
	return &ClientProtocol{socket: socket, agencyId: agencyId}
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

	return p.sendFramed(payload, BATCH)
}

func (p *ClientProtocol) SendHello() error {
	var frame []byte

	frame = append(frame, byte(p.agencyId))
	return p.socket.Send(frame)
}

func (p *ClientProtocol) sendFramed(payload []byte, is_end int8) error {
	var frame []byte

	// primero el flag de fin
	frame = append(frame, byte(is_end))
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
	return p.sendFramed(buf, END)
}

func (p *ClientProtocol) Close() error {
	return p.socket.Close()
}

func (p *ClientProtocol) RecvWinners() ([]lottery.Bet, error) {
	length, err := p.readFrame()
	if err != nil {
		return nil, err
	}

	payload, err := p.socket.Recv(length)
	if err != nil {
		return nil, err
	}

	count := binary.BigEndian.Uint32(payload[:UINT32_SIZE])
	payload = payload[UINT32_SIZE:]

	winners := make([]lottery.Bet, 0, count)	
	for i := uint32(0); i < count; i++ {
		var bet lottery.Bet
		bet, payload = p.deserializeBet(payload)
		winners = append(winners, bet)
	}

	return winners, nil
}

func (p *ClientProtocol) readFrame() (int, error) {
	length_bytes, err := p.socket.Recv(UINT32_SIZE)
	if err != nil {
		return 0, err
	}
	length := binary.BigEndian.Uint32(length_bytes)
	return int(length), nil
}

func (p *ClientProtocol) deserializeBet(payload []byte) (lottery.Bet, []byte) {
	fnLen := int(payload[0])
	payload = payload[UINT8_SIZE:]
	firstName := string(payload[:fnLen])
	payload = payload[fnLen:]

	lnLen := int(payload[0])
	payload = payload[UINT8_SIZE:]
	lastName := string(payload[:lnLen])
	payload = payload[lnLen:]

	birthdate := string(payload[:BIRTHDATE_LEN])
	payload = payload[BIRTHDATE_LEN:]

	document := int(binary.BigEndian.Uint32(payload[:UINT32_SIZE]))
	payload = payload[UINT32_SIZE:]

	number := int(binary.BigEndian.Uint32(payload[:UINT32_SIZE]))
	payload = payload[UINT32_SIZE:]

	bet := lottery.Bet{
		FirstName: firstName,
		LastName:  lastName,
		Document:  document,
		Birthdate: birthdate,
		Number:    number,
	}
	return bet, payload
}