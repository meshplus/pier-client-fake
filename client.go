package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gobuffalo/packr"
	"github.com/hashicorp/go-hclog"
	"github.com/meshplus/bitxhub-model/pb"
	"github.com/meshplus/pier/pkg/plugins"
)

type Client struct {
	config *Config
	ctx    context.Context
	cancel context.CancelFunc
	eventC chan *pb.IBTP
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

func (c *Client) Initialize(configPath string, extra []byte) error {
	cfg, err := UnmarshalConfig(configPath)
	if err != nil {
		logger.Error("unmarshal config for plugin: %v", err)
		return fmt.Errorf("unmarshal config for plugin :%w", err)
	}

	c.config = cfg
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

func (c *Client) GetIBTPCh() chan *pb.IBTP {
	return c.eventC
}

func (c *Client) GetUpdateMeta() chan *pb.UpdateMeta {
	//TODO implement me
	panic("implement me")
}

func (c *Client) SubmitIBTP(from string, index uint64, serviceID string, ibtpType pb.IBTP_Type, content *pb.Content, proof *pb.BxhProof, isEncrypted bool) (*pb.SubmitIBTPResponse, error) {
	ret := &pb.SubmitIBTPResponse{Status: true}
	return ret, nil
}

func (c *Client) SubmitReceipt(to string, index uint64, serviceID string, ibtpType pb.IBTP_Type, result *pb.Result, proof *pb.BxhProof) (*pb.SubmitIBTPResponse, error) {
	ret := &pb.SubmitIBTPResponse{Status: true}
	return ret, nil
}

func (c *Client) GetReceiptMessage(servicePair string, idx uint64) (*pb.IBTP, error) {
	return nil, fmt.Errorf("not found")
}

func (c *Client) GetDstRollbackMeta() (map[string]uint64, error) {
	//TODO implement me
	panic("implement me")
}

func (c *Client) GetServices() ([]string, error) {
	return []string{}, nil
}

func (c *Client) GetChainID() (string, string, error) {
	return "", c.config.Fake.Name, nil
}

func (c *Client) GetAppchainInfo(chainID string) (string, []byte, string, error) {
	return "", nil, "", nil
}

const PackrPath = "./config"

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

					ibtp := c.mockIBTP(i, c.proof)
					//ibtp2 := c.mockIBTP2(i, c.proof)
					c.eventC <- ibtp
					//c.eventC <- ibtp2
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

func (c *Client) mockIBTP2(index uint64, proof []byte) *pb.IBTP {
	content := &pb.Content{
		//SrcContractId: "mychannel&transfer",
		//DstContractId: "mychannel&transfer",
		Func: "interchainCharge",
		Args: [][]byte{[]byte("Alice"), []byte("Alice"), []byte("1")},
		//Callback:      "interchainConfirm",
	}

	bytes, _ := content.Marshal()

	payload := &pb.Payload{
		Encrypted: false,
		Content:   bytes,
	}

	ibtppd, _ := payload.Marshal()

	return &pb.IBTP{
		From:    ":" + c.config.Fake.Name + ":mychannel&transfer2",
		To:      ":" + c.config.Fake.To + ":mychannel&transfer2",
		Payload: ibtppd,
		Index:   index,
		Type:    pb.IBTP_INTERCHAIN,
		Proof:   proof,
	}
}

func (c *Client) mockIBTP(index uint64, proof []byte) *pb.IBTP {
	content := &pb.Content{
		//SrcContractId: "mychannel&transfer",
		//DstContractId: "mychannel&transfer",
		Func: "interchainCharge",
		Args: [][]byte{[]byte("Alice"), []byte("Alice"), []byte("1")},
		//Callback:      "interchainConfirm",
	}

	bytes, _ := content.Marshal()

	payload := &pb.Payload{
		Encrypted: false,
		Content:   bytes,
	}

	ibtppd, _ := payload.Marshal()

	return &pb.IBTP{
		From:    ":" + c.config.Fake.Name + ":mychannel&transfer",
		To:      ":" + c.config.Fake.To + ":mychannel&transfer",
		Payload: ibtppd,
		Index:   index,
		Type:    pb.IBTP_INTERCHAIN,
		Proof:   proof,
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

//// SubmitIBTP submit interchain ibtp. It will unwrap the ibtp and execute
//// the function inside the ibtp. If any execution results returned, pass
//// them to other modules.
//func (c *Client) SubmitIBTP(ibtp *pb.IBTP) (*pb.SubmitIBTPResponse, error) {
//	receipt := &pb.IBTP{
//		From:      ibtp.From,
//		To:        ibtp.To,
//		Index:     ibtp.Index,
//		Type:      pb.IBTP_RECEIPT_SUCCESS,
//		Timestamp: time.Now().UnixNano(),
//		Version:   ibtp.Version,
//	}
//
//	return &pb.SubmitIBTPResponse{
//		Status: true,
//		Result: receipt,
//	}, nil
//}

// GetOutMessage gets crosschain tx by `to` address and index
func (c *Client) GetOutMessage(to string, idx uint64) (*pb.IBTP, error) {
	return c.mockIBTP(idx, c.proof), nil
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

//func main() {
//	plugin.Serve(&plugin.ServeConfig{
//		HandshakeConfig: plugins.Handshake,
//		Plugins: map[string]plugin.Plugin{
//			plugins.PluginName: &plugins.AppchainGRPCPlugin{Impl: &Client{}},
//		},
//		Logger: logger,
//		// A non-nil value here enables gRPC serving for this plugin...
//		GRPCServer: DefaultGRPCServer,
//	})
//
//	logger.Info("Plugin server down")
//}
//
//func DefaultGRPCServer(opts []grpc.ServerOption) *grpc.Server {
//	opts = append(opts, grpc.MaxConcurrentStreams(1000),
//		grpc.MaxRecvMsgSize(100*1024*1024),
//		grpc.InitialWindowSize(10*1024*1024),
//		grpc.InitialConnWindowSize(100*1024*1024))
//	return grpc.NewServer(opts...)
//}
