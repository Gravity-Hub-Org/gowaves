package ng

import (
	"sync"

	"github.com/wavesplatform/gowaves/pkg/proto"
	"github.com/wavesplatform/gowaves/pkg/services"
	"github.com/wavesplatform/gowaves/pkg/state"
	"github.com/wavesplatform/gowaves/pkg/types"
	"go.uber.org/zap"
)

type State struct {
	storage        *storage
	prevAddedBlock *proto.Block
	applier        types.BlocksApplier
	state          state.State
	mu             sync.Mutex
	knownBlocks    knownBlocks
}

func NewState(services services.Services) *State {
	return &State{
		mu:          sync.Mutex{},
		storage:     newStorage(services.Scheme),
		applier:     services.BlocksApplier,
		state:       services.State,
		knownBlocks: knownBlocks{},
	}
}

func (a *State) AddBlock(block *proto.Block) {
	a.mu.Lock()
	defer a.mu.Unlock()

	added := a.knownBlocks.add(block)
	if !added { // already tried
		return
	}
	// same block
	if a.prevAddedBlock != nil && a.prevAddedBlock.BlockID() == block.BlockID() {
		return
	}

	err := a.storage.PushBlock(block)
	if err != nil {
		zap.S().Debugf("NG State: %v", err)
		return
	}

	mu := a.state.Mutex()
	locked := mu.Lock()
	err = a.state.RollbackTo(block.Parent)
	locked.Unlock()

	if err != nil {
		if state.IsNotFound(err) {
			zap.S().Debugf("NG State: not found block to rollback")
			if a.storage.ContainsID(block.Parent) {
				zap.S().Debugf("NG State: sig contains %s", block.Parent.String())
				prevBlock, err := a.storage.PreviousBlock()
				if err != nil {
					zap.S().Debug(err)
					return
				}
				locked := mu.Lock()
				height, err := a.state.Height()
				if err != nil {
					locked.Unlock()
					zap.S().Debug(err)
					return
				}
				err = a.state.RollbackToHeight(height - 1)
				if err != nil {
					locked.Unlock()
					zap.S().Debug(err)
					return
				}
				_, err = a.state.AddDeserializedBlock(prevBlock)
				if err != nil {
					locked.Unlock()
					zap.S().Debug(err)
					return
				}
				locked.Unlock()
			}
		} else {
			zap.S().Infof("NG State: can't rollback to id %s, initiator id %s: %v", block.Parent.String(), block.BlockID().String(), err)
			a.storage.Pop()
			return
		}
	}

	err = a.applier.Apply([]*proto.Block{block})
	if err != nil {
		a.storage.Pop()

		// return prev block, if possible
		if a.prevAddedBlock != nil {
			err := a.applier.Apply([]*proto.Block{a.prevAddedBlock})
			if err != nil {
				zap.S().Error("NG: can't apply previous added block, maybe broken ngState ", err)
			}
		}
		return
	}
	a.prevAddedBlock = block
}

func (a *State) AddMicroblock(micro *proto.MicroBlock) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.prevAddedBlock == nil {
		zap.S().Debug("prevAddedBlock is not set")
		return
	}

	err := a.storage.PushMicro(micro)
	if err != nil {
		zap.S().Debugf("Failed to push micro: %v", err)
		return
	}

	block, err := a.storage.Block()
	if err != nil {
		zap.S().Error(err)
		return
	}

	if a.prevAddedBlock.Parent != block.Parent {
		zap.S().Errorf("NG State: parents not equal, expected %q actual %q", a.prevAddedBlock.Parent, block.Parent)
		return
	}

	curHeight, err := a.state.Height()
	if err != nil {
		zap.S().Error(err)
		return
	}

	curBlock, err := a.state.BlockByHeight(curHeight)
	if err != nil {
		zap.S().Error(err)
		return
	}

	if curBlock.Parent != block.Parent {
		zap.S().Errorf("NG State: current block parent not equal prev block %q actual %q", curBlock.Parent, block.Parent)
		return
	}

	lock := a.state.Mutex()
	zap.S().Debug("Before the rollback lock()")
	locked := lock.Lock()
	zap.S().Debug("After the rollback lock()")
	err = a.state.RollbackTo(curBlock.Parent)
	if err != nil {
		zap.S().Errorf("NG State: failed to rollback to sig %s: %v", curBlock.Parent, err)
		locked.Unlock()
		return
	}
	zap.S().Debug("Successfully rollbacked")
	locked.Unlock()
	zap.S().Debug("After the rollback unlock")

	err = a.applier.Apply([]*proto.Block{block})
	if err != nil {
		zap.S().Error(err)
		// remove prev added micro
		a.storage.Pop()
		return
	}

	a.prevAddedBlock = block
}

func (a *State) BlockApplied() {
	h, err := a.state.Height()
	if err != nil {
		zap.S().Debug(err)
		return
	}
	block, err := a.state.BlockByHeight(h)
	if err != nil {
		zap.S().Debug(err)
		return
	}
	a.blockApplied(block)
}

// notify method
func (a *State) blockApplied(block *proto.Block) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.prevAddedBlock == nil {
		a.prevAddedBlock = block
		a.storage = a.storage.newFromBlock(block)
		return
	}

	if a.prevAddedBlock.BlockID() == block.BlockID() {
		return
	}

	a.prevAddedBlock = block
	a.storage = a.storage.newFromBlock(block)
}
