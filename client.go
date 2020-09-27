package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"os"
	"time"

	"github.com/gobuffalo/packr"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/meshplus/bitxhub-model/pb"
	"github.com/meshplus/pier/pkg/plugins"
)

type Client struct {
	config *Config
	ctx    context.Context
	cancel context.CancelFunc
	eventC chan *pb.IBTP
	pierID string
	proof  []byte
}

var (
	_      plugins.Client = (*Client)(nil)
	logger                = hclog.New(&hclog.LoggerOptions{
		Name:   "client",
		Output: os.Stderr,
		Level:  hclog.Trace,
	})
)

const PackrPath = "./config"

func (c *Client) Initialize(configPath string, pierID string, extra []byte) error {
	cfg, err := UnmarshalConfig(configPath)
	if err != nil {
		logger.Error("unmarshal config for plugin: %v", err)
		return fmt.Errorf("unmarshal config for plugin :%w", err)
	}

	c.config = cfg
	c.pierID = pierID
	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.eventC = make(chan *pb.IBTP, 40960)

	box := packr.NewBox(PackrPath)
	proof, err := box.Find("proof_1.0.0_rc")
	if err != nil {
		logger.Error("find proof: %w", err.Error())
		return err
	}

	c.proof = proof

	logger.Info("fake client initialized")

	return nil
}

func (c *Client) Start() error {
	logger.Info("config is ", c.config.Fake)

	go func(ctx context.Context) {
		if c.config.Fake.Tps == 0 {
			return
		}

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		i := uint64(1)
		for {
			select {
			case <-ticker.C:
				for j := uint64(0); j < uint64(c.config.Fake.Tps); j++ {
					ibtp := mockIBTP(i, c.pierID, c.config.Fake.To, c.proof)
					c.eventC <- ibtp
					i++
				}
				logger.Info("send ibtp counter:", i)
			case <-ctx.Done():
				close(c.eventC)
				return
			}
		}
	}(c.ctx)

	logger.Info("fake client started")
	return nil
}

func mockIBTP(index uint64, from, to string, proof []byte) *pb.IBTP {
	content := &pb.Content{
		SrcContractId: "mychannel&transfer",
		DstContractId: "mychannel&transfer",
		Func:          "interchainCharge",
		Args:          [][]byte{[]byte("Alice"), []byte("Alice"), []byte("1")},
		Callback:      "interchainConfirm",
	}

	bytes, _ := content.Marshal()

	payload := &pb.Payload{
		Encrypted: false,
		Content:   bytes,
	}

	ibtppd, _ := payload.Marshal()

	return &pb.IBTP{
		From:      from,
		To:        to,
		Payload:   ibtppd,
		Index:     index,
		Type:      pb.IBTP_INTERCHAIN,
		Timestamp: time.Now().UnixNano(),
		Proof:     proof,
	}
}

func (c *Client) Stop() error {
	c.cancel()
	return nil
}

func (c *Client) Name() string {
	return c.config.Fake.Name
}

func (c *Client) Type() string {
	return "fake"
}

func (c *Client) GetIBTP() chan *pb.IBTP {
	return c.eventC
}

// SubmitIBTP submit interchain ibtp. It will unwrap the ibtp and execute
// the function inside the ibtp. If any execution results returned, pass
// them to other modules.
func (c *Client) SubmitIBTP(ibtp *pb.IBTP) (*pb.SubmitIBTPResponse, error) {
	receipt := &pb.IBTP{
		From:      ibtp.From,
		To:        ibtp.To,
		Index:     ibtp.Index,
		Type:      pb.IBTP_RECEIPT_SUCCESS,
		Timestamp: time.Now().UnixNano(),
		Version:   ibtp.Version,
	}

	return &pb.SubmitIBTPResponse{
		Status: true,
		Result: receipt,
	}, nil
}

// GetOutMessage gets crosschain tx by `to` address and index
func (c *Client) GetOutMessage(to string, idx uint64) (*pb.IBTP, error) {
	return mockIBTP(idx, c.pierID, to, c.proof), nil
}

// GetInMessage gets the execution results from contract by from-index key
func (c *Client) GetInMessage(from string, idx uint64) ([][]byte, error) {
	return nil, nil
}

// GetInMeta queries contract about how many interchain txs have been
// executed on this appchain for different source chains.
func (c *Client) GetInMeta() (map[string]uint64, error) {
	return nil, nil
}

// GetOutMeta queries contract about how many interchain txs have been
// sent out on this appchain to different destination chains.
func (c *Client) GetOutMeta() (map[string]uint64, error) {
	return nil, nil
}

// GetCallbackMeta queries contract about how many callback functions have been
// executed on this appchain from different destination chains.
func (c *Client) GetCallbackMeta() (map[string]uint64, error) {
	return nil, nil
}

func (c *Client) CommitCallback(ibtp *pb.IBTP) error {
	return nil
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: plugins.Handshake,
		Plugins: map[string]plugin.Plugin{
			plugins.PluginName: &plugins.AppchainGRPCPlugin{Impl: &Client{}},
		},
		Logger: logger,
		// A non-nil value here enables gRPC serving for this plugin...
		GRPCServer: DefaultGRPCServer,
	})

	logger.Info("Plugin server down")
}

func DefaultGRPCServer(opts []grpc.ServerOption) *grpc.Server {
	opts = append(opts, grpc.MaxConcurrentStreams(1000),
		grpc.InitialWindowSize(10*1024*1024),
		grpc.InitialConnWindowSize(100*1024*1024))
	return grpc.NewServer(opts...)
}
