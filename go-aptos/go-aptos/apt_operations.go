package main

import (
	"context"
	"fmt"
	"math/big"

	"github.com/aptos-labs/aptos-go-sdk"
)

// 检查APT余额
func checkAPTBalance(ctx context.Context, client *aptos.Client, addressStr string) (*big.Int, error) {
	// 将字符串地址转换为AccountAddress类型
	address := aptos.AccountAddress{}
	err := address.ParseStringRelaxed(addressStr)
	if err != nil {
		return nil, fmt.Errorf("解析地址失败: %v", err)
	}

	// 获取账户APT余额
	balance, err := client.AccountAPTBalance(address)
	if err != nil {
		return nil, fmt.Errorf("获取APT余额失败: %v", err)
	}
	
	// 转换为big.Int
	balanceBigInt := new(big.Int).SetUint64(balance)
	
	return balanceBigInt, nil
}

// 发送APT
func sendAPT(ctx context.Context, client *aptos.Client, senderAccount *aptos.Account, recipientAddressStr string, amount *big.Int) (string, error) {
	// 将接收方地址字符串转换为AccountAddress类型
	recipientAddress := aptos.AccountAddress{}
	err := recipientAddress.ParseStringRelaxed(recipientAddressStr)
	if err != nil {
		return "", fmt.Errorf("解析接收方地址失败: %v", err)
	}

	// 创建转账payload
	payload, err := aptos.CoinTransferPayload(nil, recipientAddress, amount.Uint64())
	if err != nil {
		return "", fmt.Errorf("创建转账payload失败: %v", err)
	}

	// 构建、签名并提交交易
	resp, err := client.BuildSignAndSubmitTransaction(senderAccount, aptos.TransactionPayload{Payload: payload})
	if err != nil {
		return "", fmt.Errorf("构建、签名并提交交易失败: %v", err)
	}

	// 等待交易确认
	_, err = client.WaitForTransaction(resp.Hash)
	if err != nil {
		return "", fmt.Errorf("等待交易确认失败: %v", err)
	}
	// 验证交易是否成功
	txnInfo, err := client.TransactionByHash(resp.Hash)
	if err != nil {
		return "", fmt.Errorf("获取交易信息失败: %v", err)
	}
	// 检查交易状态
	userTxn, err := txnInfo.UserTransaction()
	if err != nil {
		return "", fmt.Errorf("解析用户交易信息失败: %v", err)
	}
	
	// fmt.Printf("交易详情:\n状态: %v\nVM状态: %s\n哈希: %s\n版本: %d\n", 
	// 	userTxn.Success, userTxn.Hash, userTxn.Version)
	if userTxn.Success {
		return resp.Hash, nil
	} else {
		return "", fmt.Errorf("交易执行失败: %s", userTxn.VmStatus)
	}
}

// 打印APT余额的可读格式
func printAPTBalance(address string, balance *big.Int) {
	// 将余额转换为可读格式 (1 APT = 10^8 Octas)
	decimalBalance := new(big.Float).SetInt(balance)
	decimalBalance = decimalBalance.Quo(decimalBalance, big.NewFloat(100000000))

	fmt.Printf("地址 %s 的APT余额: %s APT (%s Octas)\n", 
	           address, decimalBalance.Text('f', 8), balance.String())
}
