package abi

import (
	"bytes"
	"math/big"

	"github.com/pkg/errors"
	"github.com/sigurn/crc16"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

const getMethodsDictKeySz = 19

func MethodNameHash(name string) uint64 {
	// https://github.com/ton-blockchain/ton/blob/24dc184a2ea67f9c47042b4104bbb4d82289fac1/crypto/smc-envelope/SmartContract.h#L75
	return uint64(crc16.Checksum([]byte(name), crc16.MakeTable(crc16.CRC16_XMODEM))) | 0x10000
}

func getMethodsDict(code *cell.Cell) (*cell.Dictionary, error) {
	codeSlice := code.BeginParse()

	hdr, err := codeSlice.LoadSlice(56)
	if err != nil {
		return nil, errors.Wrap(err, "load slice")
	}

	// header contains methods dictionary
	// SETCP0
	// 19 DICTPUSHCONST
	// DICTIGETJMPZ
	if !bytes.Equal(hdr, []byte{0xFF, 0x00, 0xF4, 0xA4, 0x13, 0xF4, 0xBC}) {
		return nil, errors.New("cannot find methods dictionary header")
	}

	ref, err := codeSlice.LoadRef()
	if err != nil {
		return nil, errors.Wrap(err, "load ref")
	}

	dict, err := ref.ToDict(getMethodsDictKeySz)
	if err != nil {
		return nil, errors.Wrap(err, "ref to dict")
	}

	return dict, nil
}

func HasGetMethod(code *cell.Cell, getMethodName string) bool {
	var hash int64
	switch getMethodName {
	// reserved names cannot be used for get methods
	case "recv_internal", "main", "recv_external", "run_ticktock":
		return false
	default:
		hash = int64(MethodNameHash(getMethodName))
	}

	dict, err := getMethodsDict(code)
	if err != nil {
		return false
	}

	if dict.GetByIntKey(big.NewInt(hash)) != nil {
		return true
	}
	return false
}

func GetMethodHashes(code *cell.Cell) ([]uint64, error) {
	var ret []uint64

	dict, err := getMethodsDict(code)
	if err != nil {
		return nil, errors.Wrap(err, "get methods dict")
	}

	for _, kv := range dict.All() {
		i, err := kv.Key.BeginParse().LoadUInt(getMethodsDictKeySz)
		if err != nil {
			return nil, errors.Wrap(err, "load uint")
		}

		switch i {
		case 0, 1, 2, 3:
			continue
		}
		ret = append(ret, i)
	}

	return ret, nil
}