package proto

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

const (
	HeaderLength = 17
	headerMagic  = 0x12345678
)

const (
	AddrSchemeTestnet = 0x54
	ArrcSchemeMainnet = 0x54
)

const (
	GenesisTransactionType = iota + 1
	PaymentTransactionType
	IssueTransactionType
	TransferTransactionType
	ReissueTransactionType
	BurnTransactionType
	ExchangeTransactionType
	LeaseTransactionType
	LeaseCancelTransactionType
	CreateAliasTransactionType
	MassTransferTransactionType
	DataTransactionType
	SetScriptTransactionType
	SponsorFeeTransactionType
)
const (
	ContentIDGetPeers      = 0x1
	ContentIDPeers         = 0x2
	ContentIDGetSignatures = 0x14
	ContentIDSignatures    = 0x15
	ContentIDGetBlock      = 0x16
	ContentIDBlock         = 0x17
	ContentIDScore         = 0x18
	ContentIDTransaction   = 0x19
	ContentIDCheckpoint    = 0x64
)

type Address struct {
	Version       uint8
	AddrScheme    uint8
	PublicKeyHash [20]byte
	CheckSum      [4]byte
}

type Alias struct {
	Version       uint8
	AddrScheme    uint8
	AliasBytesLen uint16
	Alias         []byte
}

type Proof struct {
	Size  uint16
	Proof []byte
}

type BlockSignature [64]byte

type Block struct {
	Version                 uint8
	Timestamp               uint64
	ParentBlockSignature    BlockSignature
	ConsensusBlockLength    uint32
	BaseTarget              uint64
	GenerationSignature     [32]byte
	TransactionsBlockLength uint32
	//Transactions            []Transaction
	BlockSignature BlockSignature
}

type GenesisTransaction struct {
	Type      uint8
	Timestamp uint64
	Amount    uint64
	Recepient [26]byte
}

type IssueTransaction struct {
	Type           uint8
	Signature      BlockSignature
	Type2          uint8
	SenderKey      [32]byte
	NameLength     uint16
	NameBytes      []byte
	DescrLength    uint16
	DescrBytes     []byte
	Quantity       uint64
	Decimals       uint8
	ReissuableFlag uint8
	Fee            uint64
	Timestamp      uint64
}

type ReissueTransaction struct {
	Type           uint8
	Signature      BlockSignature
	Type2          uint8
	SenderKey      [32]byte
	AssetID        [32]byte
	Quantity       uint64
	ReissuableFlag uint8
	Fee            uint64
	Timestamp      uint64
}

type TransferTransaction struct {
	Type                    uint8
	Signature               BlockSignature
	Type2                   uint8
	SenderKey               [32]byte
	AssetFlag               uint8
	AssetID                 [32]byte
	FeeAssetFlag            uint8
	FeeAssetID              [32]byte
	Timestamp               uint64
	Amount                  uint64
	Fee                     uint64
	RecepientAddressOrAlias []byte
	AttachmentLength        uint16
	Attachment              []byte
}

type VersionedTransferTransaction struct {
	Reserved                uint8
	Type                    uint8
	Version                 uint8
	SenderKey               [32]byte
	AssetFlag               uint8
	AssetID                 [32]byte
	Timestamp               uint64
	Amount                  uint64
	Fee                     uint64
	RecepientAddressOrAlias []byte
	AttachmentLength        uint16
	AttachmentBytes         []byte
	ProofVersion            uint8
	ProofNumber             uint16
	Proofs                  []byte
}

type BurnTransaction struct {
	Type      uint8
	SenderKey [32]byte
	AssetID   [32]byte
	Amount    uint64
	Fee       uint64
	Timestamp uint64
	Signature BlockSignature
}

type ExchangeTransaction struct {
	Type                  uint8
	BuyOrderObjectLength  uint32
	SellOrderObjectLength uint32
	BuyOrderObjectBytes   []byte
	SellOrderObjectBytes  []byte
	Price                 uint64
	Amount                uint64
	BuyMatcherFee         uint64
	SellMatcherFee        uint64
	Fee                   uint64
	Timestamp             uint64
	Signature             BlockSignature
}

type LeaseTransaction struct {
	Type                    uint8
	SenderKey               [32]byte
	RecepientAddressOrAlias []byte
	Amount                  uint64
	Fee                     uint64
	Timestamp               uint64
	Signature               BlockSignature
}

type LeaseCancelTransaction struct {
	Version   uint8
	ChainByte uint8
	LeaseId   uint8
	Fee       uint64
	SenderKey [32]byte
	Timestamp uint64
}

type CreateAliasTransaction struct {
	Type             uint8
	SenderKey        [32]byte
	AliasObjectLen   uint16
	AliasObjectBytes []byte
	Fee              uint64
	Timestamp        uint64
	Signature        BlockSignature
}

type MassTransferTransaction struct {
	Type              uint8
	Version           uint8
	SenderKey         [32]byte
	AssetFlag         uint8
	AssetId           [32]byte
	NumberOfTransfers uint16
	Transfers         []byte
	Timestamp         uint8
	Fee               uint8
	AttachmenetLen    uint16
	AttachmentBytes   []byte
	ProofsVersion     uint8
	ProofCount        uint16
	Proofs            []byte
}

type DataEntry struct {
	Key1  string
	Value []byte
}

type DataTransaction struct {
	Reserved       uint8
	Type           uint8
	Version        uint8
	SenderKey      [32]byte
	NumDataEntries uint16

	DataEntries   []DataEntry
	Timestamp     uint64
	Fee           uint64
	ProofsVErsion uint8
	ProofCount    uint8
	SignatureLen  uint16
	Signature     BlockSignature
}

type SponsoredFeeTransaction struct {
	Type               uint8
	Version            uint8
	SenderKey          [32]byte
	AssetID            [32]byte
	MinimalFeeInAssets uint64
	Fee                uint64
	Timestamp          uint64
	Proofs             [64]byte
}

type SetScriptTransaction struct {
	Type              uint8
	Version           uint8
	ChainId           uint8
	SenderKey         [32]byte
	ScriptNotNull     uint8
	ScriptObjectLen   uint16
	ScriptObjectBytes []byte
	Fee               uint64
	Timestamp         uint64
}

type Order struct {
}

type Header struct {
	Length        uint32
	Magic         uint32
	ContentID     uint8
	PayloadLength uint32
	PayloadCsum   uint32
}

func (h *Header) MarshalBinary() ([]byte, error) {
	data := make([]byte, 17)

	binary.BigEndian.PutUint32(data[0:4], h.Length)
	binary.BigEndian.PutUint32(data[4:8], headerMagic)
	data[8] = h.ContentID
	binary.BigEndian.PutUint32(data[9:13], h.PayloadLength)
	binary.BigEndian.PutUint32(data[13:17], h.PayloadCsum)

	return data, nil
}

func (h *Header) UnmarshalBinary(data []byte) error {
	h.Length = binary.BigEndian.Uint32(data[0:4])
	h.Magic = binary.BigEndian.Uint32(data[4:8])
	if h.Magic != headerMagic {
		return fmt.Errorf("received wrong magic: want %x, have %x", headerMagic, h.Magic)
	}
	h.ContentID = data[8]
	h.PayloadLength = binary.BigEndian.Uint32(data[9:13])
	h.PayloadCsum = binary.BigEndian.Uint32(data[13:17])

	return nil
}

type Handshake struct {
	NameLength         uint8
	Name               string
	VersionMajor       uint32
	VersionMinor       uint32
	VersionPatch       uint32
	NodeNameLength     uint8
	NodeName           string
	NodeNonce          uint64
	DeclaredAddrLength uint32
	DeclaredAddrBytes  []byte
	Timestamp          uint64
}

type GetPeersMessage struct {
	Header
}

func (m *GetPeersMessage) MarshalBinary() ([]byte, error) {
	m.ContentID = ContentIDGetPeers
	header, err := m.Header.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return header, nil
}

func (m *GetPeersMessage) UnmarshalBinary(b []byte) error {
	err := m.Header.UnmarshalBinary(b)
	return err
}

type PeerInfo struct {
	addr net.IP
	port uint16
}

func (m *PeerInfo) MarshalBinary() ([]byte, error) {
	buffer := make([]byte, 6)

	copy(buffer[0:4], m.addr.To4())
	binary.BigEndian.PutUint16(buffer[4:6], m.port)

	return buffer, nil
}

func (m *PeerInfo) UnmarshalBinary(data []byte) error {
	if len(data) < 6 {
		return errors.New("too short")
	}

	m.addr = net.IPv4(data[0], data[1], data[2], data[3])
	m.port = binary.BigEndian.Uint16(data[4:6])

	return nil
}

type PeersMessage struct {
	Header
	PeersCount uint32
	Peers      []PeerInfo
}

func (m *PeersMessage) MarshalBinary() ([]byte, error) {
	m.ContentID = ContentIDPeers
	header, err := m.Header.MarshalBinary()
	if err != nil {
		return nil, err
	}

	body := make([]byte, 4)

	binary.BigEndian.PutUint32(body[0:4], m.PeersCount)

	for _, k := range m.Peers {
		peer, err := k.MarshalBinary()
		if err != nil {
			return nil, err
		}
		body = append(body, peer...)
	}

	header = append(header, body...)
	return header, nil
}

func (m *PeersMessage) UnmarshalBinary(data []byte) error {
	if err := m.Header.UnmarshalBinary(data[:17]); err != nil {
		return err
	}

	for i := uint32(0); i < m.Header.PayloadLength; i += 6 {
		var peer PeerInfo
		if err := peer.UnmarshalBinary(data[i : i+6]); err != nil {
			return err
		}
		m.Peers = append(m.Peers, peer)
	}

	return nil
}

type BlockID [64]byte

type GetSignaturesMessage struct {
	Header
	BlockIdsCount uint32
	Blocks        []BlockID
}

type SignaturesMessage struct {
	Header
	BlockSignaturesCount uint32
	BlockSignature       []BlockSignature
}

type GetBlockMessage struct {
	Header
	BlockID BlockID
}

type BlockMessage struct {
	Header
	BlockBytes []byte
}

type ScoreMessage struct {
	Header
	Score []byte
}

type TransactionMessage struct {
	Header
	Transaction []byte
}

type CheckpointItem struct {
	Height    uint64
	Signature BlockSignature
}

type CheckPointMessage struct {
	Header
	CheckpointItemsCount uint32
}

func (h *Handshake) marshalBinaryName() ([]byte, error) {
	data := make([]byte, h.NameLength+1)
	data[0] = h.NameLength
	copy(data[1:1+h.NameLength], h.Name)

	return data, nil
}

func (h *Handshake) marshalBinaryVersion() ([]byte, error) {
	data := make([]byte, 12)

	binary.BigEndian.PutUint32(data[0:4], h.VersionMajor)
	binary.BigEndian.PutUint32(data[4:8], h.VersionMinor)
	binary.BigEndian.PutUint32(data[8:12], h.VersionPatch)

	return data, nil
}

func (h *Handshake) marshalBinaryNodeName() ([]byte, error) {
	data := make([]byte, h.NodeNameLength+1)

	data[0] = h.NodeNameLength
	copy(data[1:1+h.NodeNameLength], h.NodeName)

	return data, nil
}

func (h *Handshake) marshalBinaryAddr() ([]byte, error) {
	data := make([]byte, 20+h.DeclaredAddrLength)

	binary.BigEndian.PutUint64(data[0:8], h.NodeNonce)
	binary.BigEndian.PutUint32(data[8:12], h.DeclaredAddrLength)

	copy(data[12:12+h.DeclaredAddrLength], h.DeclaredAddrBytes)
	binary.BigEndian.PutUint64(data[12+h.DeclaredAddrLength:20+h.DeclaredAddrLength], h.Timestamp)

	return data, nil
}

func (h *Handshake) MarshalBinary() ([]byte, error) {
	data1, err := h.marshalBinaryName()
	if err != nil {
		return nil, err
	}
	data2, err := h.marshalBinaryVersion()
	if err != nil {
		return nil, err
	}
	data3, err := h.marshalBinaryNodeName()
	if err != nil {
		return nil, err
	}
	data4, err := h.marshalBinaryAddr()
	if err != nil {
		return nil, err
	}

	data1 = append(data1, data2...)
	data1 = append(data1, data3...)
	data1 = append(data1, data4...)
	return data1, nil
}
