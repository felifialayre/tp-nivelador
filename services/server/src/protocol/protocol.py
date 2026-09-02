from dataclasses import dataclass

import safe_socket as ss
from lottery import Bet


FRAME_LEGNTH = 5
BIRTHDATE_LEN = 10
UINT8_SIZE = 1
UINT32_SIZE = 4

@dataclass
class Frame:
    agency_id: int
    length: int

class Protocol:
    def __init__(self, socket):
        self.socket = socket

    def recv_batch(self) -> tuple[list[Bet], int]:
        frame = self._read_frame()
        batch = []

        if self._is_last_batch(frame):
            return batch, frame.agency_id

        payload = ss.recv_all(self.socket, frame.length)

        batch_size = int.from_bytes(payload[:UINT32_SIZE], "big")
        payload = payload[UINT32_SIZE:]
        
        for _ in range (batch_size):
            bet, payload = self._deserialize_bet(payload, frame)
            batch.append(bet)

        return batch, frame.agency_id

    def _read_frame(self) -> Frame:
        frame_bytes = ss.recv_all(self.socket, FRAME_LEGNTH)

        agency_id = int.from_bytes(frame_bytes[:UINT8_SIZE], "big")
        length = int.from_bytes(frame_bytes[UINT8_SIZE:], "big")

        return Frame (
            agency_id,
            length,
        )

    def _deserialize_bet(self, payload, frame) -> Bet:
        fn_length = payload[0]
        payload = payload[UINT8_SIZE:]

        first_name =  payload[:fn_length].decode()
        payload = payload[fn_length:]        

        ln_length = int.from_bytes(payload[:UINT8_SIZE], "big")
        payload = payload[UINT8_SIZE:]

        last_name =  payload[:ln_length].decode()
        payload = payload[ln_length:]

        birthday = payload[:BIRTHDATE_LEN].decode()
        payload = payload[BIRTHDATE_LEN:]

        document = int.from_bytes(payload[:UINT32_SIZE], "big")
        payload = payload[UINT32_SIZE:]

        number = int.from_bytes(payload[:UINT32_SIZE], "big")
        payload = payload[UINT32_SIZE:]

        return Bet(
            frame.agency_id,
            first_name,
            last_name,
            document,
            birthday,
            number
        ), payload

    def send_winners(self, winners: list[Bet]) -> None:
        payload = b""

        # agregamos cantidad de ganadores al payload
        payload += len(winners).to_bytes(UINT32_SIZE, "big")
        for winner in winners:
            serialized = self._serialize_bet(winner)
            payload += serialized

        self._send_framed(payload)

    def _send_framed(self, payload):
        packet = b""

        # agregamos longitud total del paquete y payload
        packet += len(payload).to_bytes(UINT32_SIZE, "big")
        packet += payload

        ss.send_all(self.socket, packet)

    def _serialize_bet(self, bet : Bet) -> bytes:
        b_bet = b""

        fm_bytes = bet.first_name.encode()
        # 1 byte para longitud del nombre
        b_bet += len(fm_bytes).to_bytes(UINT8_SIZE, "big")
        b_bet += fm_bytes

        lm_bytes = bet.last_name.encode()
        # 1 byte para longitud del apellido
        b_bet += len(lm_bytes).to_bytes(UINT8_SIZE, "big")
        b_bet += lm_bytes

        b_bet += bet.birthdate.encode()

        b_bet += bet.document.to_bytes(UINT32_SIZE, "big")
        b_bet += bet.number.to_bytes(UINT32_SIZE, "big")

        return b_bet


    def _is_last_batch(self, frame) -> bool:
        return frame.length == 0