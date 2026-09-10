package protocol

import (
	"encoding/binary"
	"fmt"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/lottery"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

const BIRTHDATE_LEN = 10
const UINT32_SIZE = 4
const UINT8_SIZE = 1
const FRAME_SIZE = 5

type Opcode byte

const (
	OpBatch Opcode = iota
	OpEnd
	OpACK
	OpWinners
	OpHello
)

type ClientProtocol struct {
	socket   safe_socket.Socket
	agencyId int
}

type Frame struct {
	opcode Opcode
	length int
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

	return p.sendFramed(payload, OpBatch)
}

func (p *ClientProtocol) SendHello() error {
	var payload []byte

	payload = append(payload, byte(p.agencyId))
	return p.sendFramed(payload, OpHello)
}

func (p *ClientProtocol) sendFramed(payload []byte, opcode Opcode) error {
	var frame []byte

	// primero el flag de fin
	frame = append(frame, byte(opcode))
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
	return p.sendFramed(buf, OpEnd)
}

func (p *ClientProtocol) Close() error {
	return p.socket.Close()
}

func (p *ClientProtocol) RecvWinners() ([]lottery.Bet, error) {
	frame, err := p.readFrame()
	if err != nil {
		return nil, err
	}

	payload, err := p.socket.Recv(frame.length)
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

func (p *ClientProtocol) readFrame() (Frame, error) {
	frame_bytes, err := p.socket.Recv(FRAME_SIZE)
	if err != nil {
		return Frame{}, err
	}
	opcode := Opcode(frame_bytes[0])
	frame_bytes = frame_bytes[UINT8_SIZE:]
	length := int(binary.BigEndian.Uint32(frame_bytes[:UINT32_SIZE]))

	return Frame{opcode, length}, nil
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

func (p *ClientProtocol) RecvACK() (Opcode, error){
	frame, err := p.readFrame()
	if err != nil {
		return OpEnd, err
	}
	if frame.opcode != OpACK { return OpEnd, fmt.Errorf("ack inesperado: se recibió opcode %d", frame.opcode) }
	return frame.opcode, nil
}