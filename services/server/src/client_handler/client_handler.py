import logger
from protocol import NOT_A_ID, Protocol
from safe_lottery import SafeLottery


class ClientHandler:
    def __init__(self, safe_lottery: SafeLottery, socket, barrier):
        self._safe_lottery = safe_lottery
        self._protocol = Protocol(socket)
        self._barrier = barrier
        self._socket = socket

    def handle_client(self):
        action = "handle-client"
        try:
            #modularizar
            logger.info(action, logger.LogResult.in_progress)
            agency_id = self._protocol.recv_hello()
            if agency_id == NOT_A_ID:
                return

            total = 0
            while True:
                bets = self._protocol.recv_batch()
                if not bets:
                    logger.info(
                        action,
                        logger.LogResult.success,
                        "messages-amount",
                        total
                    )
                    break
                self._safe_lottery.store_bets(bets)
                total += len(bets)
            logger.info(action, logger.LogResult.success, "all-bets-received", total)
        except Exception as e:
            logger.error(
                action, logger.LogResult.fail
            )
            raise e

        self._barrier.wait()    

        winners = self._safe_lottery.winners_for(agency_id)
        self._protocol.send_winners(winners)

        self._socket.close()
