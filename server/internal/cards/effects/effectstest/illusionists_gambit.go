package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// illusionists_gambit_test.go — Illusionist's Gambit (#5685) and the seam
// it waited on (ADR 0188 Decision 14): the requirement and the
// restriction follow the creatures the spell removed, in the combat it
// adds and in no other.

const IllusionistsGambitOracle = "333745d9-e930-439b-94d6-3aeea2877f69"

func AttackingNow(g *game.Game, id uuid.UUID) bool {
	c, ok := g.LookupCardForEffect(id)
	return ok && c.AttackingTarget != uuid.Nil
}
