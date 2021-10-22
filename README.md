# Hyperledger Fabric 国密改造

## 1. hyperledger/fabric

下载fabric项目并切换分支至release-2.2：

```bash
mkdir $GOPATH/src/github.com/hyperledger
cd $GOPATH/src/github.com/hyperledger
git clone https://github.com/hyperledger/fabric.git
cd fabric
git checkout release-2.2
# 注意该命令只执行一次
go mod vendor
```

使用VS Code打开项目，关键词搜索替换标准库依赖为国密：

```
"crypto"\n -> "github.com/gxnublockchain/gmsupport/crypto"\n
"crypto/sha256"\n -> "github.com/gxnublockchain/gmsupport/crypto/sha256"\n
"crypto/ecdsa"\n -> "github.com/gxnublockchain/gmsupport/crypto/ecdsa"\n
"crypto/aes"\n -> "github.com/gxnublockchain/gmsupport/crypto/aes"\n
"crypto/x509"\n -> "github.com/gxnublockchain/gmsupport/crypto/x509"\n
"crypto/elliptic"\n -> "github.com/gxnublockchain/gmsupport/crypto/elliptic"\n
"crypto/tls"\n -> "github.com/gxnublockchain/gmsupport/crypto/tls"\n
"crypto/rsa"\n -> "github.com/gxnublockchain/gmsupport/crypto/rsa"\n

"net/http"\n -> "github.com/gxnublockchain/gmsupport/net/http"\n
"net/http/httptest"\n -> "github.com/gxnublockchain/gmsupport/net/http/httptest"\n
"net/http/httptrace"\n -> "github.com/gxnublockchain/gmsupport/net/http/httptrace"\n
"net/http/httputil"\n -> "github.com/gxnublockchain/gmsupport/net/http/httputil"\n
"net/http/pprof"\n -> "github.com/gxnublockchain/gmsupport/net/http/pprof"\n
```

避免使用GoLand，注意搜索的时候打开**正则匹配**以匹配换行符，没有换行符会替换掉import之外的其它内容。

在**Makefile**文件的**go install**命令后添加参数**-mod=vendor**，保证编译的时候是使用修改后的依赖。

向**images/baseos/Dockerfile**、**images/ccenv/Dockerfile**、**images/orderer/Dockerfile**、**images/peer/Dockerfile**中所有的**RUN apk**的前一行添加以下代码替换apk的镜像源为国内的源来提高Docker编译速度。

```dockerfile
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
```

在项目根目录下执行以下命令引入国密依赖：

```bash
mkdir -p vendor/github.com/gxnublockchain/
cd vendor/github.com/gxnublockchain/
# 下载国密依赖并重命名
git clone https://gitee.com/Yao-0x1e/fabric-gm-support.git
mv fabric-gm-support gmsupport
rm -rf gmsupport/.git
# 将国密兼容包的其他依赖导入
cp -r gmsupport/vendor/github.com $GOPATH/src/github.com/hyperledger/fabric/vendor
cp -r gmsupport/vendor/golang.org $GOPATH/src/github.com/hyperledger/fabric/vendor
```

由于在构建ccenv时会从系统的gopath/src复制依赖，所以需要将hyperledger/fabric/vendor下的所有内容复制到gopath/src路径下。命令如下：

```bash
cp -r $GOPATH/src/github.com/hyperledger/fabric/vendor/* $GOPATH/src
```

最后使用以下命令完成环境搭建：

```bash
# 清除上次编译时所产生的镜像
docker rmi `docker images | grep fabric-tools` -f
docker rmi `docker images | grep fabric-peer` -f
docker rmi `docker images | grep fabric-orderer` -f
# 编译命令
chmod +x scripts/*
make release
cp release/linux-amd64/bin/* /usr/local/bin
make docker
# 清除编译时产生的临时镜像
docker rmi `docker images | grep none` -f
```



## 2. hyperledger/fabric-samples

```bash
cd $GOPATH/src/github.com/hyperledger
git clone https://github.com/hyperledger/fabric-samples.git
cd fabric-samples
git checkout release-2.2
cp -r $GOPATH/src/github.com/hyperledger/fabric/sampleconfig config
```

修改**test-network/scripts/deployCC.sh**的打包中的**go mod vendor**，避免在安装链码的过程中下载非国密的依赖。

以asset-transfer-basic合约为例进行国密改造，过程如下：

```bash
cd asset-transfer-basic/chaincode-go/
go mod vendor

# 替换国密（使用之前改造fabric时的依赖）
cp -r $GOPATH/src/github.com/hyperledger/fabric-chaincode-go vendor/github.com/hyperledger
cp -r $GOPATH/src/github.com/hyperledger/fabric-protos-go vendor/github.com/hyperledger
# 导入国密相关的依赖
mkdir -p vendor/github.com/gxnublockchain
cp -r $GOPATH/src/github.com/gxnublockchain/gmsupport vendor/github.com/gxnublockchain/
cp -r $GOPATH/src/github.com/gxnublockchain/gmsupport/vendor/github.com/* vendor/github.com
cp -r $GOPATH/src/github.com/gxnublockchain/gmsupport/vendor/golang.org/* vendor/golang.org
cp -r $GOPATH/src/google.golang.org/grpc vendor/google.golang.org
```

经过上述步骤即可完成一个合约的国密改造，同样适用于其他任意Go语言编写的合约。

然后启动网络并进行测试，过程如下：

```bash
cd test-network
./network.sh up createChannel -s couchdb
./network.sh deployCC -ccn basic -ccp ../asset-transfer-basic/chaincode-go -ccl go

export FABRIC_CFG_PATH=`pwd`/../config/
export CORE_PEER_TLS_ENABLED=true

export ORDERER_CA=`pwd`/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
export PEER0_ORG1_CA=`pwd`/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export PEER0_ORG2_CA=`pwd`/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt

export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=`pwd`/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=`pwd`/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=localhost:7051

peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C mychannel -n basic --peerAddresses localhost:7051 --tlsRootCertFiles $PEER0_ORG1_CA --peerAddresses localhost:9051 --tlsRootCertFiles $PEER0_ORG2_CA -c '{"function":"InitLedger","Args":[]}'
peer chaincode query -C mychannel -n basic -c '{"function":"ReadAsset","Args":["asset1"]}'
```



## 3. hyperledger/fabric-ca

## 4. hyperledger/fabric-sdk-go