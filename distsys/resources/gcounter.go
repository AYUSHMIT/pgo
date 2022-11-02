package resources

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/UBC-NSS/pgo/distsys/tla"
	"github.com/benbjohnson/immutable"
)

type GCounter struct {
	*immutable.Map
}

var _ CRDTValue = new(GCounter)

func (c GCounter) Init() CRDTValue {
	return GCounter{immutable.NewMap(tla.ValueHasher{})}
}

func (c GCounter) Read() tla.Value {
	var value int32 = 0
	it := c.Iterator()
	for !it.Done() {
		_, v := it.Next()
		value += v.(int32)
	}
	return tla.MakeNumber(value)
}

func (c GCounter) Write(id tla.Value, value tla.Value) CRDTValue {
	oldValue, ok := c.Get(id)
	if !ok {
		oldValue = int32(0)
	}
	newValue := oldValue.(int32) + value.AsNumber()
	return GCounter{c.Set(id, newValue)}
}

// Merge current state value with other by taking the greater of each node's partial counts.
func (c GCounter) Merge(other CRDTValue) CRDTValue {
	it := other.(GCounter).Iterator()
	for !it.Done() {
		id, val := it.Next()
		if v, ok := c.Get(id); !ok || v.(int32) < val.(int32) {
			c = GCounter{c.Set(id, val)}
		}
	}
	return c
}

type GCounterKeyVal struct {
	K tla.Value
	V int32
}

func (c GCounter) GobEncode() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	it := c.Iterator()
	for !it.Done() {
		k, v := it.Next()
		pair := GCounterKeyVal{K: k.(tla.Value), V: v.(int32)}
		err := encoder.Encode(&pair)
		if err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func (c *GCounter) GobDecode(input []byte) error {
	buf := bytes.NewBuffer(input)
	decoder := gob.NewDecoder(buf)
	b := immutable.NewMapBuilder(tla.ValueHasher{})
	for {
		var pair GCounterKeyVal
		err := decoder.Decode(&pair)
		if err != nil {
			if errors.Is(err, io.EOF) {
				c.Map = b.Map()
				return nil
			} else {
				return err
			}
		}
		b.Set(pair.K, pair.V)
	}
}

func (c GCounter) String() string {
	it := c.Iterator()
	b := strings.Builder{}
	b.WriteString("map[")
	first := true
	for !it.Done() {
		if first {
			first = false
		} else {
			b.WriteString(" ")
		}
		k, v := it.Next()
		b.WriteString(k.(tla.Value).String())
		b.WriteString(":")
		b.WriteString(fmt.Sprint(v))
	}
	b.WriteString("]")
	return b.String()
}

func init() {
	gob.Register(GCounter{})
}
