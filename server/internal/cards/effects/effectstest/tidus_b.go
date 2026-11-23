package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tidus_b_test.go — the Tidus, Yuna's Guardian deck request (#3246),
// slice B: Tidus and the counter-moving guardians.

const (
	TdAuronOracle            = "5e166b42-252e-4f34-9610-dcddc04a62ff"
	TdForgottenAncientOracle = "2ef75cf3-30dc-4852-a6d1-1d6391fad022"
	TdGattaOracle            = "c01d6fde-1bd2-4412-932a-9bc97d4e5aef"
	TdJyscalOracle           = "52791629-0f22-4023-bb04-c829d1600568"
	TdRikkuOracle            = "81ba75af-eddc-4344-befb-419ebfb855a3"
	TdSphereGridOracle       = "f4770345-f15e-4e31-b6ef-4c2608df9644"
	TdTidusOracle            = "e8fbcafd-3617-4cd7-b55d-457fb7b60162"
	TdTromellOracle          = "f088817a-570d-4480-bc59-f2ad6675a71b"
	TdWakkaOracle            = "fbd37284-35ff-4409-b6d5-f813c06c4931"
)

func TdAddCounter(g *game.Game, by, id uuid.UUID, kind string, n int) {
	g.WithWriteLock(func() { _ = g.AddCounterByForEffect(by, id, kind, n) })
}
