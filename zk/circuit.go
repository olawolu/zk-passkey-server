package zk

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash/mimc"
)

// Circuit defines a zk-SNARK circuit
type Circuit struct {
	// private inputs (witnesses)
	PreImage frontend.Variable `gnark:",private"`

	// public inputs
	Hash frontend.Variable `gnark:",public"`
}

// Define declares the circuit's constraints
func (circuit *Circuit) Define(api frontend.API) error {
	// prove we know a preimage to a hash
	mimc, err := mimc.NewMiMC(api)
	if err != nil {
		return err
	}

	// compute MIMC hash
	mimc.Write(circuit.PreImage)
	api.AssertIsEqual(mimc.Sum(), circuit.Hash)
	return nil
}

// NewCircuit returns a new circuit
func NewCircuit() *Circuit {
	return &Circuit{}
}