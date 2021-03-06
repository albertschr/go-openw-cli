# openw-cli
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)

## 概述

[TOC]

## Build development environment

The requirements to build OpenWallet are:

- Golang version 1.10 or later
- Properly configured Go language environment
- Golang supported operating system

## 源码工具


```shell

# 如果不适用gomod，需要放在进入$GOPATH/src/github.com/blocktree目录下。
$ cd $GOPATH/src/github.com/blocktree/go-openw-cli

# 编译工具
$ go build


```
---

## openw-cli介绍

openw-cli是一款命令行工具，通过调用openw-server钱包服务API实现主机客户端下的钱包管理。

---

## 功能详细设计 `openw-cli`

### 配置文件

使用openw-cli需要依赖配置文件，样例如下：

```ini

# Remote Server
remoteserver = "api.openwallet.cn"

# API Version
version = "1.0.0"

# App ID
appid = "1234qwer"

# App key
appkey = "qwer1234"

# Log file path
logdir = "/usr/logs/"

# Data directory, store keys, databases, backups
datadir = "/usr/data/"

# Wallet Summary Period
summaryperiod = "1h"

# The custom name of local node
localname = "blocktree"

# Be trusted client server
trustedserver = "client.blocktree.top"

# Enable client server request local transfer
enablerequesttransfer = false

# Enable client server execute summary task
enableexecutesummarytask = false

# Enable client server edit wallet summary settings
enableeditsummarysettings = false

# Enable key agreement on local node communicate with client server
enablekeyagreement = true

# Enable https or wss
enablessl = false

# Network request timeout, unit: second
requesttimeout = 60

# Terminal print log of debug 
logdebug = false

```

我们提供命令行工具openw-cli，以下功能点作为管理资产的【子命令】，附加以下参数变量。

### 全局可选参数

| 参数变量    | 描述                                                  |
|-------------|-----------------------------------------------------|
| -s, -symbol | 币种标识符，其后带值[symbol]，如btc，ltc，eth，ada，btm，sc  |
| -c, -conf   | 工具配置文件路径。                                     |
| -i, -init   | 是否初始化，应用于配置功能时候，是否需要执行初始化流程。 |
| -p, -path   | 指定文件目录。                                         |
| -f, -file   | 指定加载的文件。                                       |
| -debug      | 是否打印debug日志信息。                                |
| -logdir     | 指定日志输出目录。                                     |

### 文件目录结构

使用openw-cli管理区块链资产时会创建以下文件目录，存储一些钱包相关的数据。目录用途说明如下：

| 参数变量                  | 描述                                                                         |
|---------------------------|----------------------------------------------------------------------------|
| {datadir}/key/               | 钱包keystore文件目录，文件命名 [alias]-[WalletID].key                         |
| {datadir}/db/                | 钱包数据库缓存目录，文件命名 [alias]-[WalletID].db                              |
| {datadir}/backup/            | 钱包备份文件导出目录，以文件夹归档备份，文件夹命名 [alias]-[WalletID]-yyyyMMddHHmmss |

> 命令输入结构: openw-cli [配置文件] [子命令] [可选参数...]
> 如：openw-cli -c ./node.ini newwallet -s btc

### 命令示例

```shell

# 临时生成随机的OWTP通信证书，不保存到本地缓存
$ ./go-openw-cli genkeychain

# 通过-c或-conf设置工具的配置文件路径
$ ./go-openw-cli -c=./node.ini

#### 节点相关 ####

# 登记到openw-server，成为应用的授权节点。
$ ./go-openw-cli -c=./node.ini noderegister

# 查看节点的信息
$ ./go-openw-cli -c=./node.ini nodeinfo

# 更新区块链资料
$ ./go-openw-cli -c=./node.ini updateinfo

#### 钱包相关 ####

# 创建钱包
$ ./go-openw-cli -c=./node.ini newwallet

# Enter wallet's name: MyWallet         //输入钱包名字
# Enter wallet password:                //输入钱包密码
# //创建成功后，显示钱包种子文件
# Wallet create successfully, key path: openw/data/key/NASSUM-W6zkTDtnWZWFd2SQPms9F62BBPfuqU2ETg.key

# 查看节点本地已创建的钱包
$ ./go-openw-cli -c=./node.ini listwallet

# 创建钱包资产账户，先选择钱包
$ ./go-openw-cli -c=./node.ini newaccount


# [Please select a wallet]
# Enter wallet No.: 0               //输入No.序号，选择本地已有钱包
# Enter wallet password:            //输入密码，解锁钱包
# Enter account's name: NASSUM      //输入需要创建资产账户名
# Enter account's symbol: NAS       //输入需要创建的币种symbol
# //创建成功，默认显示资产账户ID和默认接收地址
# create [NAS] account successfully
# new accountID: 9HqxxcNSMxdt225Dis3mdnzT18egbV7Cg3R85y6AUPx8
# new address: n1EZVYXBx5tQ41L6QRyEhpqV4TpH6NwPrPE

# 查看钱包资产账户，先选择钱包
$ ./go-openw-cli -c=./node.ini listaccount

# 创建新地址，先选择钱包，再选择资产账户
$ ./go-openw-cli -c=./node.ini newaddress


# [Please select a wallet]
# Enter wallet No.: 0                               //输入No.序号，选择本地已有钱包
# [Please select a account]
# Enter account No.: 0                              //输入No.序号，选择钱包已有的资产账户
# Enter the number of addresses you want: 100       //输入需要创建地址的数量
# create [100] addresses successfully
# //创建地址成功，把新地址导出以下文件路径。
# addresses has been exported into: openw/data/export/address/[9HqxxcNSMxdt225Dis3mdnzT18egbV7Cg3R85y6AUPx8]-20190313163227.txt

# 创建新地址，先选择钱包，再选择资产账户，输入offset和limit查询地址列表
$ ./go-openw-cli -c=./node.ini searchaddress


# 选择资产账户，发起转账交易
$ ./go-openw-cli -c=./node.ini transfer

# [Please select a wallet]                                          //输入No.序号，选择本地已有钱包
# Enter wallet No.: 0
# [Please select a account]                                         //输入No.序号，选择钱包已有的资产账户
# Enter account No.: 0
# Enter contract address:                                           //如果是代币合约转账，输入合约地址，默认空
# Enter received address: AR8LWKndC2ztfLoCobZhHEwkQCUZk1yKEsF       //输入转账的目标地址
# Enter amount to send: 2.8                                         //输入转账数量
# Enter fee rate:                                                   //输入手续费率，默认空（推荐）
# Enter wallet password:                                            //输入钱包解锁密码
# -----------------------------------------------                   //以下为转账日志信息
# [VSYS  Transfer]
# From Account: 33yYCwSeBump7V6AFX8r2KFsXqkJ7zxkg36UxyvYgy1o
# To Address: AR8LWKndC2ztfLoCobZhHEwkQCUZk1yKEsF
# Send Amount: 2.8
# Fees: 0.1
# FeeRate: 10000000
# -----------------------------------------------
# send transaction successfully.
# transaction id: GbB1oQkXQTSDudTEhKwdhyvnUvunnHKngZqGE9Xfa3tn

# 设置汇总，先选择钱包，再选择资产账户
$ ./go-openw-cli -c=./node.ini setsum


# [Please select a wallet]
# Enter wallet No.: 0                               //输入No.序号，选择本地已有钱包
# [Please select a account]
# Enter account No.: 0                                                          //输入No.序号，选择钱包已有的资产账户
# Enter account's summary address: n1EZVYXBx5tQ41L6QRyEhpqV4TpH6NwPrPE          //输入钱包汇总转账到的地址
# Enter account's summary threshold: 2                                          //输入汇总阈值，账户总余额超过此值，执行汇总交易
# Enter address's minimum transfer amount: 0.01                                 //输入地址最低转账额，地址余额超过此值才发起转账
# Enter address's retained balance: 0                                           //输入地址保留余额，地址转账时需要剩余部分余额
# Enter how many confirms can transfer: 1                                       //输入地址未花得到多少确认后才可用于转账交易
# setup summary info successfully

# 查看已有账户的汇总设置信息
$ ./go-openw-cli -c=./node.ini listsuminfo

# 启动汇总定时器
$ ./go-openw-cli -c=./node.ini startsum

# Enter summary task json file path:            //输入汇总任务json文件，如果为空，则提供选择钱包和资产账户启动汇总


# [Please select a wallet]
# Enter wallet No.: 0                               //输入No.序号，选择本地已有钱包
# [Please select a account]
# Enter account No.: 0                              //输入No.序号，选择钱包已有的资产账户

# //汇总启动成功，定时执行任务
# The timer for summary task start now. Execute by every 10 seconds.
# [Summary Task Start]------2019-03-13 16:43:33
# Summary account[9HqxxcNSMxdt225Dis3mdnzT18egbV7Cg3R85y6AUPx8] Symbol: NAS start
# Summary account[9HqxxcNSMxdt225Dis3mdnzT18egbV7Cg3R85y6AUPx8] Current Balance: 0, below threshold: 2
# Summary account[9HqxxcNSMxdt225Dis3mdnzT18egbV7Cg3R85y6AUPx8] Symbol: NAS end
# [Summary Task End]------2019-03-13 16:43:34

# 启动汇总定时器，通过文件加载需要汇总的钱包和资产账户
$ ./go-openw-cli -c=./node.ini startsum -f=/usr/to/sum.json

```

```json

`汇总样例JSON`

{
    "wallets": [
        {
            "walletID": "1234qwer",       //钱包ID
            "password": "12345678",       //钱包解锁密码
            "accounts": [                 //需要汇总的账户列表
                {
                    "accountID": "123",   //资产账户ID
                    "feeRate": "0.0001",  //交易费率，填空为推荐费率
                    "contracts":[         //汇总代币合约
                        "all",            //全部合约
                        "3qoe2ll2=",      //指定的合约ID
                    ]
                },
            ],
        },
    ]
}

```

```shell

# 查询主链列表
$ ./go-openw-cli -c=./node.ini listsymbol

# 查询主链下的合约列表
$ ./go-openw-cli -c=./node.ini listtokencontract

# 选择钱包及账户，查看账户下拥有的代币余额
$ ./go-openw-cli -c=./node.ini listtokenbalance

# 启动后台托管钱包服务
$ ./go-openw-cli -c=./node.ini trustserver

```