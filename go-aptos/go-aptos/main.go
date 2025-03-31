package main

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/aptos-labs/aptos-go-sdk"
)

// 打印使用帮助
func printUsage() {
	fmt.Println("用法:")
	fmt.Println("  检查APT余额: ./main check-apt [地址]")
	fmt.Println("  发送APT: ./main send-apt <接收地址> <数量>")
	fmt.Println("  检查TWBTC余额: ./main check-twbtc [地址]")
	fmt.Println("  注册TWBTC: ./main register-twbtc")
	fmt.Println("  发送TWBTC: ./main send-twbtc <接收地址> <数量>")
	fmt.Println("  初始化桥接: ./main init-bridge <费用账户地址> <费用>")
	fmt.Println("  redeem-request: ./main redeem-request <接收地址> <数量>")
	fmt.Println("  mint: ./main mint <btc_tx_id> <接收地址> <数量>")
	fmt.Println("  注册TWBTC: ./main registerTWBTC <接收地址>")
	fmt.Println("  初始化TWBTC: ./main init-twbtc")
}

// 主函数
func main() {
	// 从环境变量获取私钥
	privateKey := os.Getenv("PRIVATE_KEY")
	
	// 如果环境变量没有设置，尝试从命令行参数获取
	if privateKey == "" && len(os.Args) > 1 {
		privateKey = os.Args[1]
	}

	if privateKey == "" {
		logError("错误: 缺少私钥。请设置PRIVATE_KEY环境变量或作为第一个命令行参数提供")
		os.Exit(1)
	}

	// 获取模块地址
	moduleAddress, err := getModuleAddress()
	if err != nil {
		logError(err.Error())
		os.Exit(1)
	}

	// 创建上下文
	ctx := context.Background()

	// 创建客户端
	client, err := createClient()
	if err != nil {
		logError(fmt.Sprintf("创建客户端失败: %v", err))
		os.Exit(1)
	}

	// 创建账户
	account, err := createAccountFromPrivateKey(privateKey)
	if err != nil {
		logError(fmt.Sprintf("创建账户失败: %v", err))
		os.Exit(1)
	}

	// 如果没有足够的命令行参数，则显示帮助信息
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	// 解析命令
	command := os.Args[1]

	switch command {
	case "check-apt":
		// 检查APT余额
		var address string
		if len(os.Args) > 2 {
			address = os.Args[2]
		} else {
			address = account.Address.String()
		}

		balance, err := checkAPTBalance(ctx, client, address)
		if err != nil {
			logError(fmt.Sprintf("检查APT余额失败: %v", err))
			os.Exit(1)
		}

		printAPTBalance(address, balance)

	case "send-apt":
		// 发送APT
		if len(os.Args) < 4 {
			logError("错误: 发送APT需要指定接收地址和数量")
			fmt.Println("用法: ./main send-apt <接收地址> <数量(APT)>")
			os.Exit(1)
		}

		recipient := os.Args[2]
		amountStr := os.Args[3]

		// 将APT转换为Octas (1 APT = 10^8 Octas)
		amount, success := new(big.Float).SetString(amountStr)
		if !success {
			logError(fmt.Sprintf("错误: 无效的金额 %s", amountStr))
			os.Exit(1)
		}

		// 转换为Octas (整数)
		amount = amount.Mul(amount, big.NewFloat(100000000))
		amountInt, _ := amount.Int(nil)

		txHash, err := sendAPT(ctx, client, account, recipient, amountInt)
		if err != nil {
			logError(fmt.Sprintf("发送APT失败: %v", err))
			os.Exit(1)
		}

		logSuccess(fmt.Sprintf("成功发送 %s APT 到地址 %s", amountStr, recipient))
		logSuccess(fmt.Sprintf("交易哈希: %s", txHash))

	case "check-twbtc":
		// 检查TWBTC余额
		var addressStr string
		if len(os.Args) > 2 {
			addressStr = os.Args[2]
		} else {
			addressStr = account.Address.String()
		}
		address := aptos.AccountAddress{}
		err := address.ParseStringRelaxed(addressStr)
		if err != nil {
			logError(fmt.Sprintf("解析地址失败: %v", err))
			os.Exit(1)
		}
		
		balance, err := CheckTWBTCBalance(client, address, moduleAddress)
		if err != nil {
			logError(fmt.Sprintf("检查TWBTC余额失败: %v", err))
			os.Exit(1)
		}

		fmt.Printf("地址 %s 的TWBTC余额: %s Satoshis\n", addressStr, balance.String())

	case "register-twbtc":
		// 注册TWBTC代币
		txHash, err := RegisterTWBTC(client, account, moduleAddress)
		if err != nil {
			logError(fmt.Sprintf("注册TWBTC代币失败: %v", err))
			os.Exit(1)
		}

		logSuccess("成功注册TWBTC代币")
		logSuccess(fmt.Sprintf("交易哈希: %s", txHash))

	case "send-twbtc":
		// 发送TWBTC
		if len(os.Args) < 4 {
			logError("错误: 发送TWBTC需要指定接收地址和数量")
			fmt.Println("用法: ./main send-twbtc <接收地址> <数量(BTC)>")
			os.Exit(1)
		}

		recipientStr := os.Args[2]
		amountStr := os.Args[3]
		recipient := aptos.AccountAddress{}
		err := recipient.ParseStringRelaxed(recipientStr)
		if err != nil {
			logError(fmt.Sprintf("解析接收地址失败: %v", err))
			os.Exit(1)
		}

		// 检查接收方是否已注册TWBTC
		recipientBalance, err := CheckTWBTCBalance(client, recipient, moduleAddress)
		if err != nil {
			// 如果是资源不存在的错误，提示用户需要先注册
			if strings.Contains(err.Error(), "resource not found") {
				logWarning("接收方还未注册TWBTC代币，发送前需要先注册")
			} else {
				logWarning(fmt.Sprintf("检查接收方TWBTC余额失败: %v", err))
				// 但不退出，继续尝试发送
			}
		} else {
			logInfo(fmt.Sprintf("接收方当前TWBTC余额: %s Satoshis", recipientBalance.String()))
		}

		// 将BTC转换为Satoshis (1 BTC = 10^8 Satoshis)
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			logError(fmt.Sprintf("错误: 无效的金额 %s", amountStr))
			os.Exit(1)
		}
		
		// 转换为Satoshis (整数)
		amountSatoshis := uint64(amount * 100000000)

		txHash, err := SendTWBTC(client, account, recipient, amountSatoshis, moduleAddress)
		if err != nil {
			logError(fmt.Sprintf("发送TWBTC失败: %v", err))
			os.Exit(1)
		}

		logSuccess(fmt.Sprintf("成功发送 %s BTC 到地址 %s", amountStr, recipientStr))
		logSuccess(fmt.Sprintf("交易哈希: %s", txHash))

	case "init-twbtc":
		// 初始化TWBTC
		txHash, err := initTWBTC(client, account, moduleAddress)
		if err != nil {
			logError(fmt.Sprintf("初始化TWBTC失败: %v", err))
			os.Exit(1)
		}
		logSuccess(fmt.Sprintf("成功初始化TWBTC"))
		logSuccess(fmt.Sprintf("交易哈希: %s", txHash))

	case "init-bridge":
		// 初始化桥接
		if len(os.Args) < 4 {
			logError("错误: 初始化桥接需要指定管理员地址、费用账户地址和费用")
			fmt.Println("用法: ./main init-bridge  <费用账户地址> <费用>")
			os.Exit(1)
		}
		feeAccountAddress_str := os.Args[2]
		fee_str := os.Args[3]
		feeAccountAddress := aptos.AccountAddress{}
		err := feeAccountAddress.ParseStringRelaxed(feeAccountAddress_str)
		if err != nil {
			logError(fmt.Sprintf("解析费用账户地址失败: %v", err))
			os.Exit(1)
		}
		fee, err := strconv.ParseUint(fee_str, 10, 64)

		txHash, err := initBridge(client, account, moduleAddress, feeAccountAddress, fee)
		if err != nil {
			logError(fmt.Sprintf("初始化桥接失败: %v", err))
			os.Exit(1)
		}
		logSuccess(fmt.Sprintf("成功初始化桥接"))
		logSuccess(fmt.Sprintf("交易哈希: %s", txHash))

	case "redeem-request":
		// 赎回请求
		if len(os.Args) < 4 {
			logError("错误: 赎回请求需要指定接收地址和数量")
			fmt.Println("用法: ./main redeem-request <接收地址> <数量(BTC)>")
			os.Exit(1)
		}
		recipientStr := os.Args[2]
		amountStr := os.Args[3]

		amount, err := strconv.ParseUint(amountStr, 10, 64)
		if err != nil {
			logError(fmt.Sprintf("错误: 无效的金额 %s", amountStr))
			os.Exit(1)
		}
		txHash, err := redeemRequest(client, account, moduleAddress, recipientStr, amount)
		if err != nil {
			logError(fmt.Sprintf("赎回请求失败: %v", err))
			os.Exit(1)
		}
		logSuccess(fmt.Sprintf("成功发送 %s BTC 到地址 %s", amountStr, recipientStr))
		logSuccess(fmt.Sprintf("交易哈希: %s", txHash))

	case "registerTWBTC":
		if len(os.Args) < 3 {
			logError("错误: 注册TWBTC需要指定接收地址")
			fmt.Println("用法: ./main registerTWBTC <接收地址>")
			os.Exit(1)
		}
		receiverAddressStr := os.Args[2]
		receiverAddress := aptos.AccountAddress{}
		err := receiverAddress.ParseStringRelaxed(receiverAddressStr)
		txHash, err := registerTWBTC(client, account, moduleAddress, receiverAddress)
		if err != nil {
			logError(fmt.Sprintf("注册TWBTC失败: %v", err))
			os.Exit(1)
		}
		logSuccess(fmt.Sprintf("成功注册TWBTC"))
		logSuccess(fmt.Sprintf("交易哈希: %s", txHash))
	case "mint":
		// 赎回确认
		if len(os.Args) < 4 {
			logError("错误: 赎回确认需要指定btc_tx_id")
			fmt.Println("用法: ./main mint <btc_tx_id> <接收地址> <数量(BTC)>")
			os.Exit(1)
		}
		btc_tx_id := os.Args[2]
		recipientStr := os.Args[3]
		amountStr := os.Args[4]
		recipient := aptos.AccountAddress{}
		err := recipient.ParseStringRelaxed(recipientStr)
		if err != nil {
			logError(fmt.Sprintf("解析接收地址失败: %v", err))
			os.Exit(1)
		}
		amount, err := strconv.ParseUint(amountStr, 10, 64)
		if err != nil {
			logError(fmt.Sprintf("错误: 无效的金额 %s", amountStr))
			os.Exit(1)
		}
		txHash, err := mintTWBTC(client, account, moduleAddress, recipient, amount, btc_tx_id)
		if err != nil {
			logError(fmt.Sprintf("赎回确认失败: %v", err))
			os.Exit(1)
		}
		logSuccess(fmt.Sprintf("成功发送 %s BTC 到地址 %s", amountStr, recipientStr))
		logSuccess(fmt.Sprintf("交易哈希: %s", txHash))
		
	default:
		logError(fmt.Sprintf("未知命令: %s", command))
		fmt.Println("可用命令: check-apt, send-apt, check-twbtc, register-twbtc, send-twbtc")
		printUsage()
		os.Exit(1)
	}
}