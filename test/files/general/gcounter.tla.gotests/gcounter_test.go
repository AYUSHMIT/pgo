package gcounter

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/UBC-NSS/pgo/distsys"
	"github.com/UBC-NSS/pgo/distsys/resources"
	"github.com/UBC-NSS/pgo/distsys/tla"
)

func getNodeMapCtx(self tla.TLAValue, nodeAddrMap map[tla.TLAValue]string, constants []distsys.MPCalContextConfigFn) *distsys.MPCalContext {
	ctx := distsys.NewMPCalContext(self, ANode, append(constants,
		distsys.EnsureArchetypeRefParam("cntr", resources.IncrementalMapMaker(func(index tla.TLAValue) distsys.ArchetypeResourceMaker {
			if !index.Equal(self) {
				panic("wrong index")
			}
			peers := make([]tla.TLAValue, 0)
			for nid := range nodeAddrMap {
				if !nid.Equal(self) {
					peers = append(peers, nid)
				}
			}
			return resources.CRDTMaker(index, peers, func(index tla.TLAValue) string {
				return nodeAddrMap[index]
			}, 100*time.Millisecond, len(peers), resources.MakeGCounter)
		})))...)
	return ctx
}

func makeNodeBenchCtx(self tla.TLAValue, nodeAddrMap map[tla.TLAValue]string,
	constants []distsys.MPCalContextConfigFn, outCh chan tla.TLAValue) *distsys.MPCalContext {
	ctx := distsys.NewMPCalContext(self, ANodeBench, append(constants,
		distsys.EnsureArchetypeRefParam("cntr", resources.IncrementalMapMaker(func(index tla.TLAValue) distsys.ArchetypeResourceMaker {
			if !index.Equal(self) {
				panic("wrong index")
			}
			var peers []tla.TLAValue
			for nid := range nodeAddrMap {
				if !nid.Equal(self) {
					peers = append(peers, nid)
				}
			}
			return resources.CRDTMaker(index, peers, func(index tla.TLAValue) string {
				return nodeAddrMap[index]
			}, 100*time.Millisecond, len(peers), resources.MakeGCounter)
		})),
		distsys.EnsureArchetypeRefParam("out", resources.OutputChannelMaker(outCh)),
	)...)
	return ctx
}

func TestGCounter_Node(t *testing.T) {
	numNodes := 10
	constants := []distsys.MPCalContextConfigFn{
		distsys.DefineConstantValue("NUM_NODES", tla.MakeTLANumber(int32(numNodes))),
		distsys.DefineConstantValue("BENCH_NUM_ROUNDS", tla.MakeTLANumber(0)),
	}

	nodeAddrMap := make(map[tla.TLAValue]string, numNodes+1)
	for i := 1; i <= numNodes; i++ {
		portNum := 9000 + i
		addr := fmt.Sprintf("localhost:%d", portNum)
		nodeAddrMap[tla.MakeTLANumber(int32(i))] = addr
	}

	var replicaCtxs []*distsys.MPCalContext
	errs := make(chan error, numNodes)
	for i := 1; i <= numNodes; i++ {
		ctx := getNodeMapCtx(tla.MakeTLANumber(int32(i)), nodeAddrMap, constants)
		replicaCtxs = append(replicaCtxs, ctx)
		go func() {
			errs <- ctx.Run()
		}()
	}

	defer func() {
		for _, ctx := range replicaCtxs {
			ctx.Stop()
		}
	}()

	getVal := func(ctx *distsys.MPCalContext) (tla.TLAValue, error) {
		fs, err := ctx.IFace().RequireArchetypeResourceRef("ANode.cntr")
		if err != nil {
			return tla.TLAValue{}, err
		}
		return ctx.IFace().Read(fs, []tla.TLAValue{ctx.IFace().Self()})
	}

	for i := 1; i <= numNodes; i++ {
		err := <-errs
		if err != nil {
			t.Fatalf("non-nil error from ANode archetype: %s", err)
		}
	}

	for _, ctx := range replicaCtxs {
		replicaVal, err := getVal(ctx)
		log.Printf("node %s's count: %s", ctx.IFace().Self(), replicaVal)
		if err != nil {
			t.Fatalf("could not read value from cntr")
		}
		if !replicaVal.Equal(tla.MakeTLANumber(int32(numNodes))) {
			t.Fatalf("expected values %v and %v to be equal", replicaVal, numNodes)
		}
	}
}

func TestGCounter_NodeBench(t *testing.T) {
	numNodes := 3
	numRounds := 2
	numEvents := numNodes * numRounds * 2

	constants := []distsys.MPCalContextConfigFn{
		distsys.DefineConstantValue("NUM_NODES", tla.MakeTLANumber(int32(numNodes))),
		distsys.DefineConstantValue("BENCH_NUM_ROUNDS", tla.MakeTLANumber(int32(numRounds))),
	}
	iface := distsys.NewMPCalContextWithoutArchetype(constants...).IFace()

	nodeAddrMap := make(map[tla.TLAValue]string, numNodes+1)
	for i := 1; i <= numNodes; i++ {
		portNum := 9000 + i
		addr := fmt.Sprintf("localhost:%d", portNum)
		nodeAddrMap[tla.MakeTLANumber(int32(i))] = addr
	}

	var replicaCtxs []*distsys.MPCalContext
	outCh := make(chan tla.TLAValue, numEvents)
	errs := make(chan error, numNodes)
	for i := 1; i <= numNodes; i++ {
		ctx := makeNodeBenchCtx(tla.MakeTLANumber(int32(i)), nodeAddrMap, constants, outCh)
		replicaCtxs = append(replicaCtxs, ctx)
		go func() {
			errs <- ctx.Run()
		}()
	}

	starts := make(map[int32]time.Time)
	for i := 0; i < numEvents; i++ {
		resp := <-outCh
		node := resp.ApplyFunction(tla.MakeTLAString("node")).AsNumber()
		event := resp.ApplyFunction(tla.MakeTLAString("event"))
		if event.Equal(IncStart(iface)) {
			starts[node] = time.Now()
		} else if event.Equal(IncFinish(iface)) {
			elapsed := time.Since(starts[node])
			log.Println(node, elapsed)
		}
	}

	for i := 0; i < numNodes; i++ {
		err := <-errs
		if err != nil {
			t.Fatal(err)
		}
	}
}
