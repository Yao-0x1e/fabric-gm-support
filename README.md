# Hyperledger Fabric 国密改造

## 0. Proxy

Golang

```bash
# 临时设置
export GO111MODULE=on
export GOPROXY=https://goproxy.cn

# 永久设置
echo "export GO111MODULE=on" >> /etc/profile
echo "export GOPROXY=https://goproxy.cn" >> /etc/profile
source /etc/profile
```

Docker

```bash
cat > /etc/docker/daemon.json <<EOF
{
  "registry-mirrors": [
        "https://registry.docker-cn.com",
        "https://docker.mirrors.ustc.edu.cn",
        "http://hub-mirror.c.163.com",
        "https://cr.console.aliyun.com/"
  ]
}
EOF
```



## 1. hyperledger/fabric

下载fabric项目并切换分支至release-2.2：

```bash
mkdir -p $GOPATH/src/github.com/hyperledger
cd $GOPATH/src/github.com/hyperledger
git clone https://gitee.com/Yao-0x1e/fabric.git
cd fabric
git checkout release-2.2
# 注意该命令只执行一次
export GO111MODULE=on
go mod vendor
```

使用VS Code远程编辑Linux上的项目（在Windows下编辑会有文件权限的问题），关键词搜索替换标准库依赖为国密：

```
"crypto"\n -> "github.com/gxnublockchain/gmsupport/crypto"\n
"crypto/sha256"\n -> "github.com/gxnublockchain/gmsupport/crypto/sha256"\n
"crypto/ecdsa"\n -> "github.com/gxnublockchain/gmsupport/crypto/ecdsa"\n
"crypto/aes"\n -> "github.com/gxnublockchain/gmsupport/crypto/aes"\n
"crypto/x509"\n -> "github.com/gxnublockchain/gmsupport/crypto/x509"\n
"crypto/elliptic"\n -> "github.com/gxnublockchain/gmsupport/crypto/elliptic"\n
"crypto/tls"\n -> "github.com/gxnublockchain/gmsupport/crypto/tls"\n
"crypto/rsa"\n -> "github.com/gxnublockchain/gmsupport/crypto/rsa"\n
"crypto/ed25519"\n -> "github.com/gxnublockchain/gmsupport/crypto/ed25519"\n

"net/textproto"\n -> "github.com/gxnublockchain/gmsupport/net/textproto"\n
"net/http"\n -> "github.com/gxnublockchain/gmsupport/net/http"\n
"net/http/httptest"\n -> "github.com/gxnublockchain/gmsupport/net/http/httptest"\n
"net/http/httptrace"\n -> "github.com/gxnublockchain/gmsupport/net/http/httptrace"\n
"net/http/httputil"\n -> "github.com/gxnublockchain/gmsupport/net/http/httputil"\n
"net/http/pprof"\n -> "github.com/gxnublockchain/gmsupport/net/http/pprof"\n
```

或者在Linux下使用以下的Shell脚本通过sed命令来全局替换：

```shell
#!/bin/bash

sed -i 's#"crypto"$#"github.com/gxnublockchain/gmsupport/crypto"#g' `grep -rl '"crypto"' ./`
sed -i 's#"crypto/sha256"$#"github.com/gxnublockchain/gmsupport/crypto/sha256"#g' `grep -rl '"crypto/sha256"' ./`
sed -i 's#"crypto/ecdsa"$#"github.com/gxnublockchain/gmsupport/crypto/ecdsa"#g' `grep -rl '"crypto/ecdsa"' ./`
sed -i 's#"crypto/aes"$#"github.com/gxnublockchain/gmsupport/crypto/aes"#g' `grep -rl '"crypto/aes"' ./`
sed -i 's#"crypto/x509"$#"github.com/gxnublockchain/gmsupport/crypto/x509"#g' `grep -rl '"crypto/x509"' ./`
sed -i 's#"crypto/elliptic"$#"github.com/gxnublockchain/gmsupport/crypto/elliptic"#g' `grep -rl '"crypto/elliptic"' ./`
sed -i 's#"crypto/tls"$#"github.com/gxnublockchain/gmsupport/crypto/tls"#g' `grep -rl '"crypto/tls"' ./`
sed -i 's#"crypto/rsa"$#"github.com/gxnublockchain/gmsupport/crypto/rsa"#g' `grep -rl '"crypto/rsa"' ./`
sed -i 's#"crypto/ed25519"$#"github.com/gxnublockchain/gmsupport/crypto/ed25519"#g' `grep -rl '"crypto/ed25519"' ./`

sed -i 's#"net/textproto"$#"github.com/gxnublockchain/gmsupport/net/textproto"#g' `grep -rl '"net/textproto"' ./`
sed -i 's#"net/http"$#"github.com/gxnublockchain/gmsupport/net/http"#g' `grep -rl '"net/http"' ./`
sed -i 's#"net/http/httptest"$#"github.com/gxnublockchain/gmsupport/net/http/httptest"#g' `grep -rl '"net/http/httptest"' ./`
sed -i 's#"net/http/httptrace"$#"github.com/gxnublockchain/gmsupport/net/http/httptrace"#g' `grep -rl '"net/http/httptrace"' ./`
sed -i 's#"net/http/httputil"$#"github.com/gxnublockchain/gmsupport/net/http/httputil"#g' `grep -rl '"net/http/httputil"' ./`
sed -i 's#"net/http/pprof"$#"github.com/gxnublockchain/gmsupport/net/http/pprof"#g' `grep -rl '"net/http/pprof"' ./`
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
# 下载国密依赖并重命名
git clone https://gitee.com/Yao-0x1e/fabric-gm-support.git
rm -rf fabric-gm-support/.git
mv fabric-gm-support vendor/github.com/gxnublockchain/gmsupport

# 依赖github.com/gxnublockchain/gmsupport/net/textproto不被fabric所需要
rm -rf vendor/github.com/gxnublockchain/gmsupport/net/textproto
sed -i 's#"github.com/gxnublockchain/gmsupport/net/textproto"$#"net/textproto"#g' `grep -rl '"github.com/gxnublockchain/gmsupport/net/textproto"' ./`
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
# 编译命名和镜像
make release
cp release/linux-amd64/bin/* /usr/local/bin
make docker
# 清除编译时产生的临时镜像
docker rmi `docker images | grep none` -f
```



## 2. hyperledger/fabric-ca

下载fabric-ca并切换到release-1.4分支：

```bash
cd $GOPATH/src/github.com/hyperledger
git clone https://gitee.com/Yao-0x1e/fabric-ca.git
cd fabric-ca
git checkout release-1.4
# 注意这一步很重要否则无法编译
export GO111MODULE=off
```

先进行与fabric相同的依赖替换操作，然后一样地将国密包加入vendor中即可。然后通过以下命令进行部署即可：

```bash
make fabric-ca-client
make fabric-ca-server
mv bin/* /usr/local/bin

export FABRIC_CA_DYNAMIC_LINK=true
make docker
```



## 3. hyperledger/fabric-samples

下载fabric-samples并切换到release-2.2分支：

```bash
cd $GOPATH/src/github.com/hyperledger
git clone https://gitee.com/Yao-0x1e/fabric-samples.git
cd fabric-samples
git checkout release-2.2
cp -r $GOPATH/src/github.com/hyperledger/fabric/sampleconfig config
```

### 3.1 asset-transfer-basic/chaincode-go

删除**test-network/scripts/deployCC.sh**脚本中的**go mod vendor**，避免在安装链码的过程中下载非国密的依赖。

以asset-transfer-basic中的合约为例进行国密改造，过程如下：

```bash
cd asset-transfer-basic/chaincode-go/
export GO111MODULE=on
go mod vendor
```

然后对替换asset-transfer-basic/chaincode-go文件夹里的密码学依赖为国密，同样为关键词搜索替换。替换完成后导入国密依赖即可：

```bash
# 导入国密相关的依赖
mkdir -p vendor/github.com/gxnublockchain
git clone https://gitee.com/Yao-0x1e/fabric-gm-support.gi
rm -rf fabric-gm-support/.git
mv fabric-gm-support vendor/github.com/gxnublockchain/gmsupport
```

经过上述步骤即可完成一个合约的国密改造，同样适用于其他任意Go语言编写的合约。

然后启动网络并进行测试，过程如下：

```bash
cd test-network
./network.sh up createChannel -s couchdb -ca
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



### 3.2 asset-transfer-basic/application-go

以asset-transfer-basic中的SDK客户端为例进行国密改造，过程如下：

```bash
cd asset-transfer-basic/application-go/
export GO111MODULE=on
go mod vendor
```

然后对替换asset-transfer-basic/application-go文件夹里的密码学依赖为国密，同样为关键词搜索替换。替换完成后导入国密依赖即可：

```bash
# 导入国密相关的依赖
mkdir -p vendor/github.com/gxnublockchain
git clone https://gitee.com/Yao-0x1e/fabric-gm-support.git
rm -rf fabric-gm-support/.git
mv fabric-gm-support vendor/github.com/gxnublockchain/gmsupport
```

经过上述步骤即可完成一个SDK客户端的国密改造，同样适用于其他任意Go语言编写的SDK客户端。

然后在test-network已经启动了的情况下使用以下命令测试即可：

```bash
# 删除现有的密钥和钱包
rm -rf keystore
rm -rf wallet

go run ./assetTransfer.go
go build
./asset-transfer-basic
```

