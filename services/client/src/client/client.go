package client

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/lottery"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/protocol"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

const CONNECTION_ATTEMPTS_MAX = 3
const CONNECTION_ATTEMPS_DELAY_MS = 200

type ClientConfig struct {
	ServerHost string
	ServerPort string
	AgencyId   string
	InputFile  string
	OutputFile string
	BatchSize  int
}

type Client struct {
	protocol protocol.ClientProtocol
	config   ClientConfig
}

func NewClient(config ClientConfig) (*Client, error) {
	const action = "creating-client"
	conn, err := connectToServer(config.ServerHost, config.ServerPort)
	if err != nil {
		logger.Warn("connect-to-server", logger.Fail)
		return nil, err
	}
	socket := safe_socket.CrearSafeSocket(conn)

	agencyId, err := strconv.Atoi(config.AgencyId)
	if err != nil {
		logger.Warn(action, logger.Fail, "err", err)
		return nil, err
	}

	client := &Client{protocol: *protocol.CrearProtocolo(socket, agencyId), config: config}
	return client, nil
}

func connectToServer(host, port string) (net.Conn, error) {
	const action = "connect-to-server"
	var err error
	var conn net.Conn

	logger.Info(action, logger.InProgress)
	for i := range CONNECTION_ATTEMPTS_MAX {
		conn, err = net.Dial("tcp", host+":"+port)
		if err != nil {
			logger.Warn(action, logger.Fail, "attempt", i)
			time.Sleep(CONNECTION_ATTEMPS_DELAY_MS * time.Millisecond)
			continue
		}

		logger.Info(action, logger.Success)
		break
	}

	return conn, err
}

func (client *Client) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		<-ctx.Done()
		client.protocol.Close()
	}()

	err := client.run()
	if err != nil && ctx.Err() != nil {
		logger.Info("client-run", logger.Success, "reason", "sigterm")
		return nil
	}
	return err
}

func (client *Client) run() error {
	defer client.protocol.Close()

	if err := client.protocol.SendHello(); err != nil {
		logger.Warn("send-hello", logger.Fail, "err", err)
		return err
	}

	if err := client.sendBets(); err != nil {
		return err
	}

	return client.getWinners()
}

func (client *Client) sendBets() error {
	const mainAction = "get-winners"

	input_file, err := os.Open(client.config.InputFile)
	if err != nil {
		logger.Warn("open-input-file", logger.Fail, "err", err)
		return err
	}

	defer input_file.Close()

	input := bufio.NewScanner(input_file)

	for {
		batch, err := client.readBatch(input)

		if err != nil {
			logger.Warn(mainAction, logger.Fail, "err", err)
			return err
		}
		if len(batch) == 0 {
			err := client.protocol.SendEnd()
			if err != nil {
				logger.Error("send-end", logger.Fail, "err", err)
				return err
			}
			break
		}
		if err := client.protocol.SendBatch(batch); err != nil {
			logger.Error("send-batch", logger.Fail, "err", err)
			return err
		}
		if _, err := client.protocol.RecvACK(); err != nil {
			return err
		}
	}

	if err := input.Err(); err != nil {
		logger.Warn("read-file", logger.Fail, "err", err)
		return err
	}
	logger.Info(mainAction, logger.Success, "agency-id", client.config.AgencyId)

	return nil
}

func (client *Client) getWinners() error {
	winners, err := client.protocol.RecvWinners()
	if err != nil {
		logger.Warn("get-winners", logger.Fail, "err", err)
		return err
	}

	return client.writeWinners(winners)
}

func (client *Client) writeWinners(winners []lottery.Bet) error {
	output_file, err := os.Create(client.config.OutputFile)
	if err != nil {
		logger.Warn("create-output-file", logger.Fail, "err", err)
		return err
	}
	defer output_file.Close()

	for _, bet := range winners {
		line := fmt.Sprintf("%s,%s,%d,%s,%d\n",
			bet.FirstName, bet.LastName, bet.Document, bet.Birthdate, bet.Number)

		n, err := output_file.WriteString(line)
		if err != nil {
			logger.Error("write-winner", logger.Fail, "err", err)
			return err
		}
		if n != len(line) {
			return fmt.Errorf("short write: %d de %d bytes", n, len(line))
		}
	}

	return nil
}

func (client *Client) readBatch(input *bufio.Scanner) ([]lottery.Bet, error) {
	const action = "read-batch"
	b_size := client.config.BatchSize
	batch := make([]lottery.Bet, 0, b_size)

	for i := 0; i < b_size && input.Scan(); i++ {
		bet, err := parseBet(input.Text())
		if err != nil {
			logger.Warn(action, logger.Fail, "err", err)
			return nil, err
		}

		batch = append(batch, bet)
	}

	return batch, nil
}

func parseBet(line string) (lottery.Bet, error) {
	fields := strings.Split(line, ",")
	if len(fields) != 5 {
		return lottery.Bet{}, fmt.Errorf("line with %d fields, 5 were expected: %q", len(fields), line)
	}

	document, err := strconv.Atoi(fields[2])
	if err != nil {
		return lottery.Bet{}, fmt.Errorf("invalid document: %w", err)
	}

	number, err := strconv.Atoi(fields[4])
	if err != nil {
		return lottery.Bet{}, fmt.Errorf("invalid number: %w", err)
	}

	return lottery.Bet{
		FirstName: fields[0],
		LastName: fields[1],
		Document: document,
		Birthdate: fields[3],
		Number: number,
	}, nil
}
