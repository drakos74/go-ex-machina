package xmachina

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"testing"

	"github.com/drakos74/go-ex-machina/xmachina/ml"
	"github.com/drakos74/go-ex-machina/xmachina/net/ff"
	"github.com/drakos74/go-ex-machina/xmath"
	"github.com/stretchr/testify/assert"
)

func testNetwork_BinaryClassificationInMem_Benchmark(t *testing.T) {

	// build the network
	network := ff.New(2, 1).
		Add(2, ff.Perceptron(ml.Model(), xmath.Const(0.5))). // hidden layer
		Add(1, ff.Perceptron(ml.Model(), xmath.Const(0.5)))  // output layer

	// parse the input data
	b, err := ioutil.ReadFile("test/testdata/bin_class_input.csv")
	assert.NoError(t, err)

	reader := csv.NewReader(bytes.NewBuffer(b))

	records, err := reader.ReadAll()
	assert.NoError(t, err)

	inputSet := xmath.Mat(len(records))
	outputSet := xmath.Mat(len(records))

	for i, record := range records {
		inp := xmath.Vec(len(record) - 1)
		out := xmath.Vec(len(record) - 2)

		for j, value := range record {
			f, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
			if err != nil {
				panic(fmt.Sprintf("cannot Train with non-numeric value %v: %v", value, err))
			}
			if j < 2 {
				inp[j] = f
			} else {
				out[j-2] = f
			}
		}
		inputSet[i] = inp
		outputSet[i] = out
	}

	TrainInMem(Training(0.0001, 10000), network, inputSet, outputSet)

	// check trained network performance

	for i, input := range inputSet {
		o := network.Predict(input).Round()
		r := outputSet[i]
		assert.Equal(t, o, r)
	}

}
