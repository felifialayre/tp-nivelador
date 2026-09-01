import safe_socket as ss
from lottery import Bet

FRAME_LEGNTH = 5
BIRTHDATE_LEN = 10
UINT8_SIZE = 1
UINT32_SIZE = 4

class Protocol:
    def __init__(self, socket):
        self.socket = socket

    def recv_bet(self) -> Bet:
        agency_id, packet_length = self._read_frame()
        bet = self._deserialize_bet(packet_length, agency_id)

        return bet


    def _read_frame(self) -> tuple[int, int]:
        frame_bytes = ss.recv_all(self.socket, FRAME_LEGNTH)

        agency_id = int.from_bytes(frame_bytes[:UINT8_SIZE], "big")
        length = int.from_bytes(frame_bytes[UINT8_SIZE:], "big")

        return agency_id, length

    def _deserialize_bet(self, packet_length, agency_id) -> Bet:
        packet_bytes = ss.recv_all(self.socket, packet_length)
        offset = 0

        fn_length = int.from_bytes(packet_bytes[:UINT8_SIZE], "big")
        offset += UINT8_SIZE
        first_name =  packet_bytes[offset:offset + fn_length].decode()
        offset += fn_length

        ln_length = int.from_bytes(packet_bytes[offset:offset+1], "big")
        offset += UINT8_SIZE
        last_name =  packet_bytes[offset:offset+ln_length].decode()
        offset += ln_length

        birthday = packet_bytes[offset:offset + BIRTHDATE_LEN].decode()
        offset += BIRTHDATE_LEN

        document = int.from_bytes(packet_bytes[offset:offset + UINT32_SIZE], "big")
        offset += UINT32_SIZE

        number = int.from_bytes(packet_bytes[offset:offset + UINT32_SIZE], "big")

        return Bet(
            agency_id,
            first_name,
            last_name,
            document,
            birthday,
            number
        )