# solana-rpc-client-extensions-go
[![](https://img.shields.io/github/go-mod/go-version/golang/go/release-branch.go1.19?filename=src%2Fgo.mod&label=GO%20VERSION&style=for-the-badge&logo=appveyor)](https://github.com/golang/go/releases/tag/go1.19)

Go code to perform Solana RPC's GetStakeActivation client-side.
This code was implemented with reference to the following repository.

- [solana-rpc-client-extensions](https://github.com/anza-xyz/solana-rpc-client-extensions)

## Dependencies
- [solana-go-sdk](https://github.com/blocto/solana-go-sdk)

## Motivation
The GetStakeActivation RPC code is removed, but users may still need to get access to stake activation data.
- https://solana.com/ja/docs/rpc/deprecated/getstakeactivation

The RPC method was removed because it's possible to get calculate the status of a stake account on the client-side.

This repo contains go code for mimicking GetStakeActivation on the client-side. See the examples/ in each repo to see how to use them, or read the source code!

## Usage

### Installing

```shell
go get -v github.com/skport/solana-rpc-client-extensions-go
```