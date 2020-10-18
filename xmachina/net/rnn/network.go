package rnn

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/drakos74/go-ex-machina/xmachina/net"

	"github.com/drakos74/go-ex-machina/xmachina/ml"
	"github.com/drakos74/go-ex-machina/xmath"
)

type Network struct {
	*RNNLayer
	net.Stats

	learn         ml.Learning
	activation    ml.SoftActivation
	neuronFactory NeuronFactory

	n, xDim, hDim int

	loss                      ml.MLoss
	predictInput, trainOutput *xmath.Window
	TmpOutput                 xmath.Vector
}

// NewRNNLayer creates a new Recurrent layer
// n : batch size e.g. rnn units
// xDim : size of trainInput/trainOutput vector
// hDim : internal hidden layer size
// rate : learning rate
func New(n, xDim, hDim int) *Network {
	return &Network{
		n:            n,
		xDim:         xDim,
		hDim:         hDim,
		predictInput: xmath.NewWindow(n),
		trainOutput:  xmath.NewWindow(n + 1),
	}
}

// WithWeights initialises the network recurrent layer and generates the starting weights.
func (net *Network) WithWeights(weights Weights) *Network {
	if net.RNNLayer != nil {
		panic("rnn layer already initialised")
	}
	net.RNNLayer = LoadRNNLayer(
		net.n,
		net.xDim,
		net.hDim,
		net.learn,
		net.neuronFactory,
		weights, 0)
	return net
}

// InitWeights initialises the network recurrent layer and generates the starting weights.
func (net *Network) InitWeights(weightGenerator xmath.ScaledVectorGenerator) *Network {
	if net.RNNLayer != nil {
		panic("rnn layer already initialised")
	}
	net.RNNLayer = NewRNNLayer(
		net.n,
		net.xDim,
		net.hDim,
		net.learn,
		net.neuronFactory,
		weightGenerator, 0)
	return net
}

func (net *Network) Rate(rate float64) *Network {
	net.learn = ml.Learn(rate)
	return net
}

func (net *Network) Activation(activation ml.Activation) *Network {
	net.neuronFactory = RNeuron(activation)
	return net
}

func (net *Network) SoftActivation(activation ml.SoftActivation) *Network {
	net.activation = activation
	return net
}

func (net *Network) Loss(loss ml.MLoss) *Network {
	net.loss = loss
	return net
}

func (net *Network) Train(data xmath.Vector) (err xmath.Vector, weights Weights) {
	// add our trainInput & trainOutput to the batch
	var batchIsReady bool
	batchIsReady = net.trainOutput.Push(data)
	// be ready for predictions ... from the start
	net.predictInput.Push(data)
	loss := xmath.Vec(len(data))
	net.TmpOutput = xmath.Vec(len(data))
	if batchIsReady {
		// we can actually train now ...
		batch := net.trainOutput.Batch()

		inp := xmath.Inp(batch)
		outp := xmath.Outp(batch)

		// forward pass
		out := net.Forward(inp)

		// keep the last data as the standard data
		net.TmpOutput = out[len(out)-1]

		// add the cross entropy loss for each of the vectors
		loss = net.loss(outp, out)

		// backward pass
		net.Backward(outp)
		// update stats
		net.Iteration++
		net.Stats.Add(loss.Sum())
		// log progress
		if net.Iteration%1000 == 0 {
			log.Info().
				Int("epoch", net.Iteration).
				Float64("err", loss.Sum()).
				Str("mean-err", fmt.Sprintf("%+v", net.Stats.Bucket)).
				Msg("training iteration")
		}
	}

	return loss, net.RNNLayer.Weights()

}

func (net *Network) Predict(input xmath.Vector) xmath.Vector {

	batchIsReady := net.predictInput.Push(input)

	if batchIsReady {
		batch := net.predictInput.Batch()
		out := net.Forward(batch)
		return out[len(out)-1]
	}

	return xmath.Vec(len(input))
}
