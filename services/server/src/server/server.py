import socket
import logger
import safe_socket
from protocol import Protocol
from lottery import Lottery

_ECHO_SERVER_MESSAGE_SIZE = 1024


class Server:
    def __init__(self, server_host: str, server_port: int) -> None:
        self.server_host = server_host
        self.server_port = server_port

    def _handle_client(self, client_socket):
        protocol = Protocol(client_socket)
        action = "handle-client"
        try:
            #por ahora lee una sola apuesta pero quiero primero poder testear bien el protocolo
            logger.info(action, logger.LogResult.in_progress)
            total = 0
            while True:
                bets = protocol.recv_batch()
                if not bets:
                    logger.info(
                        action,
                        logger.LogResult.success,
                        "messages-amount",
                    )
                    break
                total += len(bets)
                logger.info(action, logger.LogResult.success, "bets-received", total)
        except Exception as e:
            logger.error(
                action, logger.LogResult.fail, "messages-amount"
            )
            raise e

    def run(self):
        action = "accept-connection"
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as server_socket:
            server_socket.bind((self.server_host, self.server_port))
            server_socket.listen()
            while True:
                try:
                    logger.info(action, logger.LogResult.in_progress)
                    client_socket, _ = server_socket.accept()
                except Exception as e:
                    logger.error(action, logger.LogResult.fail)
                    raise e
                logger.info(action, logger.LogResult.success)

                self._handle_client(client_socket)
