import signal
from threading import BrokenBarrierError

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
        signal.signal(signal.SIGTERM, self._handle_sigterm)
        action = "handle-client"
        try:
            logger.info(action, logger.LogResult.in_progress)
            agency_id = self._protocol.recv_hello()
            if agency_id == NOT_A_ID:
                return

            total = 0
            while True:
                bets_received = self._handle_batch()
                if bets_received == 0:
                    break
                total += bets_received
            logger.info(action, logger.LogResult.success, "all-bets-received", total)

            self._barrier.wait()

            winners = self._safe_lottery.winners_for(agency_id)
            self._protocol.send_winners(winners)
        except (BrokenBarrierError, OSError):
            logger.info(action, logger.LogResult.success, "graceful-shutdown", "sigterm-received")
        except Exception as e:
            logger.error(action, logger.LogResult.fail, "err", e)
        finally:
            self._socket.close()

    def _handle_batch(self):
        bets = self._protocol.recv_batch()
        if not bets:
            return 0
        self._safe_lottery.store_bets(bets)
        self._protocol.send_ack()
        return len(bets)

    def _handle_sigterm(self, _signum, _frame):
        self._socket.close()