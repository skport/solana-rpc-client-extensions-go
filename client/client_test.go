package client

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	sdkRpc "github.com/blocto/solana-go-sdk/rpc"
)

var (
	rpc sdkRpc.RpcClient
)

func TestMain(t *testing.M) {
	rpc = sdkRpc.NewRpcClient(sdkRpc.DevnetRPCEndpoint)

	ctx := context.Background()
	r, err := rpc.GetVersion(ctx)
	if err != nil {
		log.Fatalf("failed to version info, err: %v", err)
	}
	fmt.Println("solana core version", r.Result.SolanaCore)

	os.Exit(t.Run())
}
