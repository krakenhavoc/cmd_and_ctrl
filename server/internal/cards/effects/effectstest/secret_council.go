package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// secret_council_test.go — ADR 0180, #2926: secret council. No ballot
// reaches another seat, a spectator, the admin's view or the log before
// the last one is in; the reveal is one log line; a restore keeps the
// ballots hidden; and the six cards.

const (
	TruthOrConsequencesOracle = "58d944fb-d1ba-4141-baaf-37dd10978656"
	MobVerdictOracle          = "cb91e09e-f7fb-44bc-b78f-de29705bd2d6"
	CirdanOracle              = "063e05b0-554f-4b55-9feb-e11f3927f3a6"
	TrapTheTrespassersOracle  = "ef72af33-37fc-48ed-90d8-547c90e1609d"
	Vault11Oracle             = "e2c00aff-696c-4b20-acd4-95b71b5f8b10"
	ElrondOracle              = "5ae4a172-b846-4ec2-8178-a8b762134bfe"
)

func CastTruthOrConsequences(t *testing.T, g *game.Game) {
	t.Helper()
	CastCatalogSpell(t, g, "Truth or Consequences", "Sorcery", TruthOrConsequencesOracle, nil)
	PassPriorityAroundTable(t, g)
}
