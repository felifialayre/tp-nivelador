import socket

import logger
from lottery import Bet, Lottery
from protocol import Protocol, NOT_A_ID


class Server:
    def __init__(self, server_host: str, server_port: int, storage_path: str) -> None:
        self.server_host = server_host
        self.server_port = server_port
        self.lottery = Lottery(storage_path)

    def _handle_client(self, client_socket):
        protocol = Protocol(client_socket)
        action = "handle-client"
        try:
            #modularizar
            logger.info(action, logger.LogResult.in_progress)
            agency_id = protocol.recv_hello()
            if agency_id == NOT_A_ID:
                return

            total = 0
            while True:
                bets = protocol.recv_batch()
                if not bets:
                    logger.info(
                        action,
                        logger.LogResult.success,
                        "messages-amount",
                        total
                    )
                    break
                self.lottery.store_bets(bets)
                total += len(bets)
            logger.info(action, logger.LogResult.success, "all-bets-received", total)
        except Exception as e:
            logger.error(
                action, logger.LogResult.fail
            )
            raise e

        winners = self._find_winners(agency_id)
        protocol.send_winners(winners)
        
        client_socket.close()

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

    def _find_winners(self, agency_id) -> list[Bet]:
        winners = []
        for bet in self.lottery.load_bets():
            if self.lottery.has_won(bet) and bet.agency_id == agency_id:
                winners.append(bet)
                
        return winners