import multiprocessing
import signal
import socket

import logger
from client_handler import ClientHandler
from safe_lottery import SafeLottery

SHUTDOWN_TIMEOUT_SECONDS = 2

class Server:
    def __init__(self, server_host: str, server_port: int, storage_path: str, agency_quorum_min: int) -> None:
        self._safe_lottery = SafeLottery(storage_path)
        self._barrier = multiprocessing.Barrier(agency_quorum_min)

        self._handlers = []
        self._running = True

        try:
            self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            self._server_socket.bind((server_host, server_port))
        except Exception as e:
            logger.info("init-socket", logger.LogResult.fail)
            return
        self._server_socket.listen()

        signal.signal(signal.SIGTERM, self._handle_sigterm)

    def run(self):
        action = "accept-connection"
        while self._running: 
            try:
                logger.info(action, logger.LogResult.in_progress)
                client_socket, _ = self._server_socket.accept()
            except OSError:
                logger.info(action, logger.LogResult.fail, "sigterm-received")
                break
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

        for handle_process in self._handlers:
            # propagamos sigterm a los hijos
            handle_process.terminate()

        for handle_process in self._handlers:
            handle_process.join(SHUTDOWN_TIMEOUT_SECONDS)
            if not handle_process.is_alive():
                handle_process.close()


    def _handle_sigterm(self, _signum, _frame):
        self._running = False
        try:
            self._server_socket.close()
        except Exception as e:
            logger.error("close-socket", logger.LogResult.fail)
            raise e
        self._barrier.abort()