package effectstest

import (
	"sort"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// productionDefKeys is every catalog key the production registry filed,
// captured by TestMain after every init() and before any test registers
// a fixture of its own.
var ProductionDefKeys []string

func snapshotProductionDefKeys() {
	ProductionDefKeys = ProductionDefKeys[:0]
	for k := range effects.Defs {
		ProductionDefKeys = append(ProductionDefKeys, k)
	}
	sort.Strings(ProductionDefKeys)
}
