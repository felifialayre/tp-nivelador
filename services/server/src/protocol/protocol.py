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

    def recv_batch(self) -> list[Bet] :
        frame = self._read_frame()
        batch = []

        if self._is_last_batch(frame):
            return batch

        payload = ss.recv_all(self.socket, frame.length)

        batch_size = int.from_bytes(payload[:UINT32_SIZE], "big")
        payload = payload[UINT32_SIZE:]
        
        for _ in range (batch_size):
            bet, payload = self._deserialize_bet(payload, frame)
            batch.append(bet)

        return batch

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

    def _is_last_batch(self, frame) -> bool:
        return frame.length == 0