package heuristic_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/tiers"
)

// seat_test.go gives this package's pin tests the object a real seat
// is given, instead of the bare policy they used to assert against.
//
// #1060 is why. `heuristic.New()` is a complete aiseat.Policy and
// implements every hook this package ships, and no game ever holds
// one: the lobby seats what tiers.Factory builds, which wraps it in a
// rules.Filter (`heuristic`) or a model.Policy (`assisted`,
// `strong`). Both wrappers swallowed #687's and #1013's hooks, and
// the pins here went on passing for a month while neither ordering
// happened at a table.
//
// Importing aiseat/tiers from an EXTERNAL test package is what makes
// this expressible. tiers imports heuristic, so the non-test package
// may never import tiers back; heuristic_test is a separate package
// compiled after both, so it can, and the import ban in
// imports_test.go still covers the real package.

// seatedTiers are the tiers whose policy has this package somewhere
// inside it — every tier but `random`, which deliberately has no
// opinion about a board.
func seatedTiers() []string {
	return []string{
		string(aiseat.TierHeuristic),
		string(aiseat.TierAssisted),
		string(aiseat.TierStrong),
	}
}

// factoryPolicy builds one seat's policy the way the lobby does: the
// production factory, with a transport configured so that the model
// tiers are available.
func factoryPolicy(t *testing.T, tier string) aiseat.Policy {
	t.Helper()
	f := tiers.NewFactory(tiers.FactoryOptions{Client: model.AlwaysIndex(0)})
	p, err := f.NewPolicy(aiseat.SeatSpec{PlayerID: uuid.New(), Tier: tier})
	if err != nil {
		t.Fatalf("factory could not build the %s seat: %v", tier, err)
	}
	return p
}
