import multiprocessing

from lottery import Bet, Lottery


class SafeLottery:
    def __init__(self, storage_path):
        self._lock = multiprocessing.Lock()
        self._lottery = Lottery(storage_path)

    def store_bets(self, bets: list[Bet]) -> None:
        with self._lock:
            self._lottery.store_bets(bets)

    def winners_for(self, agency_id) -> list[Bet]:
        with self._lock:
            winners = []
            for bet in self._lottery.load_bets():
                if self._lottery.has_won(bet) and bet.agency_id == agency_id:
                    winners.append(bet)

        return winners