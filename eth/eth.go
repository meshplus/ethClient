package eth

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"ethClient/internal/repo"
	"ethClient/internal/solidity"
	"fmt"
	"github.com/Rican7/retry"
	"github.com/Rican7/retry/strategy"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/compiler"
	"github.com/ethereum/go-ethereum/common/hexutil"
	types1 "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// EthRPC - Ethereum rpc client
type EthRPC struct {
	url        string
	client     httpClient
	log        logger
	Debug      bool
	privateKey *ecdsa.PrivateKey
	cid        *big.Int
}

type Ethereum struct {
	etherCli   *ethclient.Client
	privateKey *ecdsa.PrivateKey
	cid        *big.Int
}

type Config struct {
	EtherAddr    string
	KeyPath      string
	PasswordPath string
}

// New create new rpc client with given url
func New(url string) *EthRPC {
	rpc := &EthRPC{
		url:    url,
		client: http.DefaultClient,
		log:    log.New(os.Stderr, "", log.LstdFlags),
	}
	return rpc
}

func (rpc *EthRPC) call(method string, target interface{}, params ...interface{}) error {
	result, err := rpc.Call(method, params...)
	if err != nil {
		return err
	}

	if target == nil {
		return nil
	}

	return json.Unmarshal(result, target)
}

// URL returns client url
func (rpc *EthRPC) URL() string {
	return rpc.url
}

// Call returns raw response of method call
func (rpc *EthRPC) Call(method string, params ...interface{}) (json.RawMessage, error) {
	request := ethRequest{
		ID:      1,
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}

	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	response, err := rpc.client.Post(rpc.url, "application/json", bytes.NewBuffer(body))
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if rpc.Debug {
		rpc.log.Println(fmt.Sprintf("%s\nRequest: %s\nResponse: %s\n", method, body, data))
	}

	resp := new(ethResponse)
	if err := json.Unmarshal(data, resp); err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, *resp.Error
	}

	return resp.Result, nil

}

func (rpc *EthRPC) CompileContract(code string) (*CompileResult, error) {
	data, err := rpc.Call("contract_"+"compileContract", code)
	if err != nil {
		return nil, err
	}

	var cr CompileResult
	if sysErr := json.Unmarshal(data, &cr); sysErr != nil {
		return nil, sysErr
	}
	return &cr, nil
}

// Compile compiles all given Solidity source files.
func (rpc *EthRPC) Compile(codePath string, local bool) (*CompileResult, error) {
	if !local {
		return rpc.CompileContract(codePath)
	}
	codePaths := strings.Split(codePath, ",")
	contracts, err := compiler.CompileSolidity("", codePaths...)
	if err != nil {
		return nil, fmt.Errorf("compile contract: %w", err)
	}

	var (
		abis  []string
		bins  []string
		types []string
	)
	for name, contract := range contracts {
		Abi, err := json.Marshal(contract.Info.AbiDefinition) // Flatten the compiler parse
		if err != nil {
			return nil, fmt.Errorf("failed to parse ABIs from compiler output: %w", err)
		}
		abis = append(abis, string(Abi))
		bins = append(bins, contract.Code)
		types = append(types, name)
	}

	result := &CompileResult{
		Abi:   abis,
		Bin:   bins,
		Types: types,
	}
	return result, nil
}

func NewEther(config Config, repoRoot string) (*Ethereum, error) {
	configPath := filepath.Join(repoRoot, "ethereum")
	var keyPath string
	if len(config.KeyPath) == 0 {
		keyPath = filepath.Join(configPath, "account.key")
	} else {
		keyPath = config.KeyPath
	}

	etherCli, err := ethclient.Dial(config.EtherAddr)
	if err != nil {
		return nil, err
	}

	keyByte, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	var password string
	if len(config.PasswordPath) == 0 {
		psdPath := filepath.Join(configPath, "password")
		psd, err := ioutil.ReadFile(psdPath)
		if err != nil {
			return nil, err
		}
		password = strings.TrimSpace(string(psd))
	} else {
		psd, err := ioutil.ReadFile(config.PasswordPath)
		if err != nil {
			return nil, err
		}
		password = strings.TrimSpace(string(psd))
	}

	unlockedKey, err := keystore.DecryptKey(keyByte, password)
	if err != nil {
		return nil, err
	}

	Cid, err := etherCli.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	return &Ethereum{
		etherCli:   etherCli,
		privateKey: unlockedKey.PrivateKey,
		cid:        Cid,
	}, nil
}

func (rpc *EthRPC) Deploy(config Config, codePath, argContract string, local bool) (string, *CompileResult, error) {
	repoRoot, err := repo.PathRoot()
	if err != nil {
		return "", nil, err
	}
	ether, err := NewEther(config, repoRoot)
	if err != nil {
		return "", nil, err
	}
	// compile solidity first
	compileResult, err := rpc.Compile(codePath, local)
	if err != nil {
		return "", nil, err
	}

	var addr common.Address

	if len(compileResult.Abi) == 0 || len(compileResult.Bin) == 0 || len(compileResult.Types) == 0 {
		return "", nil, fmt.Errorf("empty contract")
	}

	auth, err := bind.NewKeyedTransactorWithChainID(ether.privateKey, ether.cid)
	if err != nil {
		return "", nil, err
	}

	for i, bin := range compileResult.Bin {
		if bin == "0x" {
			continue
		}
		parsed, err := abi.JSON(strings.NewReader(compileResult.Abi[i]))
		if err != nil {
			return "", nil, err
		}
		code := strings.TrimPrefix(strings.TrimSpace(bin), "0x")

		// prepare for constructor parameters
		var argx []interface{}
		if len(argContract) != 0 {
			argSplits := strings.Split(argContract, "^")
			var argArr []interface{}
			for _, arg := range argSplits {
				if strings.Index(arg, "[") == 0 && strings.LastIndex(arg, "]") == len(arg)-1 {
					if len(arg) == 2 {
						argArr = append(argArr, make([]string, 0))
						continue
					}
					// deal with slice
					argSp := strings.Split(arg[1:len(arg)-1], ",")
					argArr = append(argArr, argSp)
					continue
				}
				argArr = append(argArr, arg)
			}
			argx, err = solidity.Encode(parsed, "", argArr...)
			if err != nil {
				return "", nil, err
			}
		}

		addr1, tx, _, err := bind.DeployContract(auth, parsed, common.FromHex(code), ether.etherCli, argx...)
		addr = addr1
		if err != nil {
			return "", nil, err
		}
		var r *types1.Receipt
		if err := retry.Retry(func(attempt uint) error {
			r, err = ether.etherCli.TransactionReceipt(context.Background(), tx.Hash())
			if err != nil {
				return err
			}

			return nil
		}, strategy.Wait(1*time.Second)); err != nil {
			return "", nil, err
		}

		if r.Status == types1.ReceiptStatusFailed {
			return "", nil, fmt.Errorf("deploy contract failed, tx hash is: %s", r.TxHash.Hex())
		}
		//write abi file
		dir := filepath.Dir(compileResult.Types[i])
		base := filepath.Base(compileResult.Types[i])
		ext := filepath.Ext(compileResult.Types[i])
		f := strings.TrimSuffix(base, ext)
		filename := fmt.Sprintf("%s.abi", f)
		p := filepath.Join(dir, filename)
		err = ioutil.WriteFile(p, []byte(compileResult.Abi[i]), 0644)
		if err != nil {
			return "", nil, err
		}
	}
	return addr.Hex(), compileResult, nil
}

//// EthSendTransaction creates new message call transaction or a contract creation, if the data field contains code.
//func (rpc *EthRPC) EthSendTransaction(data hexutil.Bytes) (common.Hash, error) {
//	var hash common.Hash
//
//	err := rpc.call("eth_sendRawTransaction", &hash, data)
//	return hash, err
//}

// EthSendTransaction creates new message call transaction or a contract creation, if the data field contains code.

func (rpc *EthRPC) EthSendTransactionWithReceipt() {

}

// EthGetTransactionReceipt returns the receipt of a transaction by transaction hash.
// Note That the receipt is not available for pending transactions.
func (rpc *EthRPC) EthGetTransactionReceipt(hash common.Hash) (*TransactionReceipt, error) {
	transactionReceipt := new(TransactionReceipt)

	err := rpc.call("eth_getTransactionReceipt", transactionReceipt, hash)
	if err != nil {
		return nil, err
	}

	return transactionReceipt, nil
}

// EthSendRawTransaction creates new message call transaction or a contract creation for signed transactions.
func (rpc *EthRPC) EthSendRawTransaction(data hexutil.Bytes) (common.Hash, error) {
	var hash common.Hash

	err := rpc.call("eth_sendRawTransaction", &hash, data)
	return hash, err
}

//Eth Invoke method
func (rpc *EthRPC) InvokeEthContract(method string, args ...interface{}) (json.RawMessage, error) {
	return rpc.Call(method, args)
}

// Eth1 returns 1 ethereum value (10^18 wei)
func Eth1() *big.Int {
	return big.NewInt(1000000000000000000)
}
