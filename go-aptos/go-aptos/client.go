package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/aptos-labs/aptos-go-sdk"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/crypto"
)

// 创建 Aptos 客户端
func createClient() (*aptos.Client, error) {
	// 从环境变量获取网络配置，默认为devnet
	networkConfig := aptos.DevnetConfig // TODO mainnet

	// 创建客户端
	client, err := aptos.NewClient(networkConfig)
	if err != nil {
		return nil, fmt.Errorf("创建客户端失败: %v", err)
	}
	return client, nil
}

// 创建账户从私钥
func createAccountFromPrivateKey(privateKeyHex string) (*aptos.Account, error) {
	// 解码私钥
	privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x")
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("解析私钥失败: %v", err)
	}

	// 创建Ed25519私钥
	key := crypto.Ed25519PrivateKey{}
	err = key.FromBytes(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("创建Ed25519私钥失败: %v", err)
	}

	// 从签名者创建账户
	account, err := aptos.NewAccountFromSigner(&key)
	if err != nil {
		return nil, fmt.Errorf("从私钥创建账户失败: %v", err)
	}

	return account, nil
}
// 构建并提交交易的通用函数
func buildAndSubmitTransaction(
	ctx context.Context,
	client *aptos.Client,
	account *aptos.Account,
	function string,
	typeArgs []string,
	args []interface{},
) (string, error) {
	// 构建模块ID和函数名
	parts := strings.Split(function, "::")
	if len(parts) != 3 {
		return "", fmt.Errorf("无效的函数格式，应为 'address::module::function'")
	}
	
	address := aptos.AccountAddress{}
	err := address.ParseStringRelaxed(parts[0])
	if err != nil {
		return "", fmt.Errorf("解析地址失败: %v", err)
	}
	
	moduleId := aptos.ModuleId{
		Address: address,
		Name:    parts[1],
	}
	
	// 转换类型参数
	var typeTags []aptos.TypeTag
	for _, typeArg := range typeArgs {
		typeTag, err := aptos.ParseTypeTag(typeArg)
		if err != nil {
			return "", fmt.Errorf("解析类型参数失败: %v", err)
		}
		typeTags = append(typeTags, *typeTag)
	}
	
	// 转换参数为字节数组
	var argsBytes [][]byte
	for _, arg := range args {
		var argBytes []byte
		var err error
		
		switch v := arg.(type) {
		case string:
			if strings.HasPrefix(v, "0x") {
				// 处理地址
				addr := aptos.AccountAddress{}
				err = addr.ParseStringRelaxed(v)
				if err != nil {
					return "", fmt.Errorf("解析地址参数失败: %v", err)
				}
				argBytes, err = bcs.Serialize(&addr)
			} else {
				// 处理普通字符串 - 不直接使用bcs.Serialize
				// 根据官方例子，字符串需要特殊处理
				argBytes = []byte(v)
			}
		case uint64:
			argBytes, err = bcs.SerializeU64(v)
		case bool:
			argBytes, err = bcs.SerializeBool(v)
		default:
			return "", fmt.Errorf("不支持的参数类型: %T", arg)
		}
		
		if err != nil {
			return "", fmt.Errorf("序列化参数失败: %v", err)
		}
		argsBytes = append(argsBytes, argBytes)
	}
	
	// 构建交易负载
	payload := aptos.TransactionPayload{
		Payload: &aptos.EntryFunction{
			Module:   moduleId,
			Function: parts[2],
			ArgTypes: typeTags,
			Args:     argsBytes,
		},
	}

	// 构建、签名并提交交易
	resp, err := client.BuildSignAndSubmitTransaction(account, payload)
	if err != nil {
		return "", fmt.Errorf("构建、签名并提交交易失败: %v", err)
	}

	// 等待交易确认
	_, err = client.WaitForTransaction(resp.Hash)
	if err != nil {
		return "", fmt.Errorf("等待交易确认失败: %v", err)
	}

	return resp.Hash, nil
}

// 日志信息输出函数，带颜色
func logInfo(message string) { 
	fmt.Printf("\x1b[36m%s\x1b[0m\n", message)
}

func logSuccess(message string) { 
	fmt.Printf("\x1b[32m%s\x1b[0m\n", message)
}

func logError(message string) { 
	fmt.Printf("\x1b[31m%s\x1b[0m\n", message)
}

func logWarning(message string) { 
	fmt.Printf("\x1b[33m%s\x1b[0m\n", message)
}

// 获取模块地址
func getModuleAddress() (string, error) {
	moduleAddress := os.Getenv("MODULE_PUBLISHER_ACCOUNT_ADDRESS")
	if moduleAddress == "" {
		return "", fmt.Errorf("缺少模块地址。请设置MODULE_PUBLISHER_ACCOUNT_ADDRESS环境变量")
	}
	return moduleAddress, nil
}
