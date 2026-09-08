import multiprocessing
import socket

import logger
from client_handler import ClientHandler
from safe_lottery import SafeLottery


class Server:
    def __init__(self, server_host: str, server_port: int, storage_path: str, agency_quorum_min: int) -> None:
        self.server_host = server_host
        self.server_port = server_port
        self._safe_lottery = SafeLottery(storage_path)
        self._barrier = multiprocessing.Barrier(agency_quorum_min)
        self._handlers = []

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

                new_handler = ClientHandler(self._safe_lottery, client_socket, self._barrier)
                handler_process = multiprocessing.Process(target=new_handler.handle_client)
                self._handlers.append(handler_process)
                handler_process.start()

                client_socket.close()
                # esta línea es necesaria porque python forkea el proceso así que se 
                # creó una copia de los fd (cierro solamente la copia q tiene el server)

