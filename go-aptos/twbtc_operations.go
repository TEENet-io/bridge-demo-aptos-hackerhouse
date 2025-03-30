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

	// 验证交易是否成功
	txnInfo, err := client.TransactionByHash(txnHash)
	if err != nil {
		return "", fmt.Errorf("获取交易信息失败: %v", err)
	}
	// 检查交易状态
	userTxn, err := txnInfo.UserTransaction()
	if err != nil {
		return "", fmt.Errorf("解析用户交易信息失败: %v", err)
	}
		
	if userTxn.Success {
		return txnHash, nil
	} else {
		return "", fmt.Errorf("交易执行失败: %s", userTxn.VmStatus)
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

	// 验证交易是否成功
	txnInfo, err := client.TransactionByHash(txnHash)
	if err != nil {
		return "", fmt.Errorf("获取交易信息失败: %v", err)
	}
	// 检查交易状态
	userTxn, err := txnInfo.UserTransaction()
	if err != nil {
		return "", fmt.Errorf("解析用户交易信息失败: %v", err)
	}
	if userTxn.Success {
		return txnHash, nil
	} else {
		return "", fmt.Errorf("交易执行失败: %s", userTxn.VmStatus)
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

	// 验证交易是否成功	
	txnInfo, err := client.TransactionByHash(txnHash)
	if err != nil {
		return "", fmt.Errorf("获取交易信息失败: %v", err)
	}
	userTxn, err := txnInfo.UserTransaction()
	if err != nil {
		return "", fmt.Errorf("解析用户交易信息失败: %v", err)	
	}
	if userTxn.Success {
		return txnHash, nil
	} else {
		return "", fmt.Errorf("交易执行失败: %s", userTxn.VmStatus)
	}


	return txnHash, nil

}

func registerTWBTC(client *aptos.Client, account aptos.TransactionSigner, moduleAddress string, receiverAddress aptos.AccountAddress) (string, error) {
	address := aptos.AccountAddress{}
	err := address.ParseStringRelaxed(moduleAddress)
	if err != nil {
		return "", fmt.Errorf("解析地址失败: %v", err)
	}

	receiverAddressBytes, err := bcs.Serialize(&receiverAddress)
	if err != nil {
		return "", fmt.Errorf("序列化接收方地址失败: %v", err)
	}

	rawTxn, err := client.BuildTransaction(account.AccountAddress(), aptos.TransactionPayload{
		Payload: &aptos.EntryFunction{
			Module: aptos.ModuleId{
				Address: address,
				Name:    "btc_tokenv3",
			},
			Function: "registerv2",
			ArgTypes: []aptos.TypeTag{},
			Args: [][]byte{receiverAddressBytes},
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

	// 验证交易是否成功
	txnInfo, err := client.TransactionByHash(txnHash)
	if err != nil {
		return "", fmt.Errorf("获取交易信息失败: %v", err)
	}
	userTxn, err := txnInfo.UserTransaction()
	if err != nil {
		return "", fmt.Errorf("解析用户交易信息失败: %v", err)
	}
	if userTxn.Success {
		return txnHash, nil
	} else {
		return "", fmt.Errorf("交易执行失败: %s", userTxn.VmStatus)
	}

	return txnHash, nil


}


func mintTWBTC(client *aptos.Client, account aptos.TransactionSigner, moduleAddress string, receiverAddress aptos.AccountAddress, amount uint64, btc_tx_id string) (string, error) {
	address := aptos.AccountAddress{}
	err := address.ParseStringRelaxed(moduleAddress)
	if err != nil {
		return "", fmt.Errorf("解析地址失败: %v", err)
	}

	// btc_tx_id: String,
	// receiver: address,
	// amount: u64,


	// 正确的BCS编码方式 - 首先需要编码字符串长度，然后是字符串内容
	btc_tx_id_len := len(btc_tx_id)
	btc_tx_id_bytes := make([]byte, 0)

	// 添加字符串长度(使用LEB128编码)
	if btc_tx_id_len < 128 {
		btc_tx_id_bytes = append(btc_tx_id_bytes, byte(btc_tx_id_len))
	} else {
		// 处理更长的字符串...
		// 这里简化处理，实际上应该使用LEB128编码
		encodedLen := byte(btc_tx_id_len) | 0x80
		btc_tx_id_bytes = append(btc_tx_id_bytes, encodedLen, byte(btc_tx_id_len>>7))
	}

	// 添加字符串内容
	btc_tx_id_bytes = append(btc_tx_id_bytes, []byte(btc_tx_id)...)


	receiverAddressBytes, err := bcs.Serialize(&receiverAddress)
	if err != nil {	
		return "", fmt.Errorf("序列化接收方地址失败: %v", err)
	}
	amountBytes, err := bcs.SerializeU64(amount)
	if err != nil {
		return "", fmt.Errorf("序列化金额失败: %v", err)
	}

	rawTxn, err := client.BuildTransaction(account.AccountAddress(), aptos.TransactionPayload{
		Payload: &aptos.EntryFunction{
			Module: aptos.ModuleId{
				Address: address,
				Name:    "btc_bridgev3",
			},
			Function: "mint",
			ArgTypes: []aptos.TypeTag{},
			Args: [][]byte{btc_tx_id_bytes, receiverAddressBytes, amountBytes},
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

	// 验证交易是否成功
	txnInfo, err := client.TransactionByHash(txnHash)
	if err != nil {
		return "", fmt.Errorf("获取交易信息失败: %v", err)
	}
	userTxn, err := txnInfo.UserTransaction()
	if err != nil {
		return "", fmt.Errorf("解析用户交易信息失败: %v", err)
	}
	if userTxn.Success {
		return txnHash, nil
	} else {
		return "", fmt.Errorf("交易执行失败: %s", userTxn.VmStatus)
	}

	return txnHash, nil
}


func redeemRequest(client *aptos.Client, account aptos.TransactionSigner, moduleAddress string, receiverAddress string, amount uint64) (string, error) {
	address := aptos.AccountAddress{}
	err := address.ParseStringRelaxed(moduleAddress)
	if err != nil {
		return "", fmt.Errorf("解析地址失败: %v", err)
	}
	amountBytes, err := bcs.SerializeU64(amount)
	if err != nil {
		return "", fmt.Errorf("序列化金额失败: %v", err)
	}
	

	// 正确的BCS编码方式 - 首先需要编码字符串长度，然后是字符串内容
	receiverStrLen := len(receiverAddress)
	receiverBytes := make([]byte, 0)

	// 添加字符串长度(使用LEB128编码)
	if receiverStrLen < 128 {
		receiverBytes = append(receiverBytes, byte(receiverStrLen))
	} else {
		// 处理更长的字符串...
		// 这里简化处理，实际上应该使用LEB128编码
		encodedLen := byte(receiverStrLen) | 0x80
		receiverBytes = append(receiverBytes, encodedLen, byte(receiverStrLen>>7))
	}

	// 添加字符串内容
	receiverBytes = append(receiverBytes, []byte(receiverAddress)...)

	// 使用正确编码的String类型调用合约
	rawTxn, err := client.BuildTransaction(account.AccountAddress(), aptos.TransactionPayload{
		Payload: &aptos.EntryFunction{
			Module: aptos.ModuleId{
				Address: address,
				Name:    "btc_bridgev3",
			},
			Function: "redeem_request",
			ArgTypes: []aptos.TypeTag{},
			Args:     [][]byte{amountBytes, receiverBytes},
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
	// 验证交易是否成功
	txnInfo, err := client.TransactionByHash(txnHash)
	if err != nil {
		return "", fmt.Errorf("获取交易信息失败: %v", err)
	}
	userTxn, err := txnInfo.UserTransaction()
	if err != nil {
		return "", fmt.Errorf("解析用户交易信息失败: %v", err)
	}
	if userTxn.Success {
		return txnHash, nil
	} else {
		return "", fmt.Errorf("交易执行失败: %s", userTxn.VmStatus)
	}



	return txnHash, nil
}



