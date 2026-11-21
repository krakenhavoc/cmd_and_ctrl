package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// designations_test.go is the Class and Case half of ADR 0071's
// catalog side — the three cards that prove the one gate on real
// printed text: a gated static's worth of level lines (Wizard Class),
// a gated cost modifier (Fortune Teller's Talent) and a gated trigger
// (Case of the Shattered Pact). The station half is station_test.go;
// the engine contract behind both — the predicate, the filter, the
// layer invalidation, the snapshot — is game/designations_test.go.

const (
	WizardClassOracle           = "36f68aa3-9955-46f1-bc87-497f16ef5222"
	FortuneTellersTalentOracle  = "1b430f67-3686-4452-8594-b060f1a5a04e"
	CaseOfTheShatteredPactOracl = "d00a089b-7d0d-4b89-902b-7b914e892178"
)

// pushClass seeds a Class permanent that is not summoning sick and
// carries a CR 613 timestamp, so the layer pass sees it.
func PushClass(g *game.Game, owner uuid.UUID, name, typeLine, oracle string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   oracle,
		Owner:      owner,
		Controller: owner,
	})
}

// levelOf reads a battlefield permanent's CR 716.2 level through the
// engine's accessor, which is the same one the wire and the level-up
// condition read.
func LevelOf(g *game.Game, id uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() { n = g.ClassLevelFor(id) })
	return n
}

// levelUpTo activates the Class's "Level N" ability and resolves it,
// paying with mana conjured into the pool. Returns the activation
// error so the timing and CR 716.2a tests can assert on it.
func LevelUpTo(t *testing.T, g *game.Game, controller, classID uuid.UUID, index int, mana string) error {
	t.Helper()
	g.WithWriteLock(func() { _ = g.AddManaForEffect(controller, uuid.Nil, mana) })
	if err := g.ActivateCatalogAbility(controller, classID, index, game.ActivateAbilityParams{}); err != nil {
		return err
	}
	PassPriorityAroundTable(t, g)
	return nil
}

// --- shared helpers --------------------------------------------------
//
// layeredCard (ability_removal_test.go) is the read that forces a
// recompute first, which is what makes Effective()-backed answers
// (IsCreature, HasKeyword, CurrentPower) reflect the current board.

func IsSolved(g *game.Game, id uuid.UUID) bool {
	solved := false
	g.ReadSnapshot(func() { solved = g.IsSolved(id) })
	return solved
}
