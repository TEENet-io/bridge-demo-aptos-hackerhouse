package main

import (
	"fmt"
	"math/big"

	"github.com/aptos-labs/aptos-go-sdk"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
)
// CheckTWBTCBalance checks the TWBTC token balance for an account
func CheckTWBTCBalance(client *aptos.Client, address aptos.AccountAddress, moduleAddress string) (*big.Int, error) {
	resources, err := client.AccountResources(address)
	if err != nil {
		return nil, fmt.Errorf("获取账户资源失败: %v", err)
	}

	moduleAddr := aptos.AccountAddress{}
	err = moduleAddr.ParseStringRelaxed(moduleAddress)
	if err != nil {
		return nil, fmt.Errorf("解析模块地址失败: %v", err)
	}

	resourceType := fmt.Sprintf("0x1::coin::CoinStore<%s::btc_tokenv3::BTC>", moduleAddr.String())
	
	var balance *big.Int = big.NewInt(0)
	for _, resource := range resources {
		if resource.Type == resourceType {
			if coinData, ok := resource.Data["coin"]; ok {
				if coinMap, ok := coinData.(map[string]interface{}); ok {
					if valueStr, ok := coinMap["value"].(string); ok {
						balance, _ = new(big.Int).SetString(valueStr, 10)
					}
				}
			}
			break
		}
	}

	return balance, nil
}

// SendTWBTC sends TWBTC tokens to another account
func SendTWBTC(client *aptos.Client, senderAccount aptos.TransactionSigner, receiverAddress aptos.AccountAddress, amount uint64, moduleAddress string) (string, error) {
	// toodo
	return "", nil
}

// RegisterTWBTC registers the TWBTC token for an account
func RegisterTWBTC(client *aptos.Client, account aptos.TransactionSigner, moduleAddress string) (string, error) {
	address := aptos.AccountAddress{}
	err := address.ParseStringRelaxed(moduleAddress)
	if err != nil {
		return "", fmt.Errorf("解析地址失败: %v", err)
	}

	rawTxn, err := client.BuildTransaction(account.AccountAddress(), aptos.TransactionPayload{
		Payload: &aptos.EntryFunction{
			Module: aptos.ModuleId{
				Address: address,
				Name:    "btc_tokenv2",
			},
			Function: "register",
			ArgTypes: []aptos.TypeTag{},
			Args: [][]byte{},
		},
	},
	)
	if err != nil {
		panic("Failed to build transaction:" + err.Error())
	}
	signedTxn, err := rawTxn.SignedTransaction(account)
	if err != nil {
		panic("Failed to sign transaction:" + err.Error())
	}

	// 4. Submit transaction
	submitResult, err := client.SubmitTransaction(signedTxn)
	if err != nil {
		panic("Failed to submit transaction:" + err.Error())
	}
	txnHash := submitResult.Hash

	_, err = client.WaitForTransaction(txnHash)
	if err != nil {
		panic("Failed to wait for transaction:" + err.Error())
	}

	return txnHash, nil
}



func initTWBTC(client *aptos.Client, account aptos.TransactionSigner, moduleAddress string) (string, error) {
	address := aptos.AccountAddress{}
	err := address.ParseStringRelaxed(moduleAddress)
	if err != nil {
		return "", fmt.Errorf("解析地址失败: %v", err)
	}

	rawTxn, err := client.BuildTransaction(account.AccountAddress(), aptos.TransactionPayload{
		Payload: &aptos.EntryFunction{
			Module: aptos.ModuleId{
				Address: address,
				Name:    "btc_tokenv3",
			},
			Function: "initialize_module",
			ArgTypes: []aptos.TypeTag{},
			Args: [][]byte{},
		},
	},
	)
	if err != nil {
		panic("Failed to build transaction:" + err.Error())
	}
	signedTxn, err := rawTxn.SignedTransaction(account)
	if err != nil {
		panic("Failed to sign transaction:" + err.Error())
	}

	submitResult, err := client.SubmitTransaction(signedTxn)
	if err != nil {
		panic("Failed to submit transaction:" + err.Error())
	}
	txnHash := submitResult.Hash

	_, err = client.WaitForTransaction(txnHash)
	if err != nil {
		panic("Failed to wait for transaction:" + err.Error())
	}

	return txnHash, nil
}


func initBridge(client *aptos.Client, account aptos.TransactionSigner, moduleAddress string, feeAccount aptos.AccountAddress, fee uint64) (string, error) {
	address := aptos.AccountAddress{}
	err := address.ParseStringRelaxed(moduleAddress)
	if err != nil {
		return "", fmt.Errorf("解析地址失败: %v", err)
	}
	// admin: &signer,
	// // pk: vector<u8>,
	// fee_account: address,
	// fee: u64
	if err != nil {
		return "", fmt.Errorf("解析地址失败: %v", err)
	}
	
	// Convert feeAccount.Address to []byte
	feeAccountBytes, err := bcs.Serialize(&feeAccount)
	if err != nil {
		return "", fmt.Errorf("序列化费用账户地址失败: %v", err)
	}
	
	// Convert fee to []byte
	feeBytes, err := bcs.SerializeU64(fee)
	if err != nil {
		return "", fmt.Errorf("序列化费用失败: %v", err)
	}
	
	rawTxn, err := client.BuildTransaction(account.AccountAddress(), aptos.TransactionPayload{
		Payload: &aptos.EntryFunction{
			Module: aptos.ModuleId{
				Address: address,
				Name:    "btc_bridgev3",
			},
			Function: "initialize",
			ArgTypes: []aptos.TypeTag{},
			Args:     [][]byte{feeAccountBytes, feeBytes},
		},
	},
	)
	if err != nil {
		panic("Failed to build transaction:" + err.Error())
	}
	signedTxn, err := rawTxn.SignedTransaction(account)
	if err != nil {
		panic("Failed to sign transaction:" + err.Error())
	}

	submitResult, err := client.SubmitTransaction(signedTxn)
	if err != nil {
		panic("Failed to submit transaction:" + err.Error())
	}
	txnHash := submitResult.Hash

	_, err = client.WaitForTransaction(txnHash)
	if err != nil {
		panic("Failed to wait for transaction:" + err.Error())
	}

	return txnHash, nil

}


