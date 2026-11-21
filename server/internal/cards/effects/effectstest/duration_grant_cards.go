package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// duration_grant_cards_test.go — ADR 0093 PR 4 (#1584): abilities a
// RESOLVING spell or ability grants for a duration, on real cards. The
// grant is a ScopedEffect record with a grantAbilities mod; every test
// asserts what a player sees — the creature coming back, the Treasure,
// the bounce, the Saga's mana — not merely that a record was written.

const (
	FeignDeathOracle       = "f718e507-296b-4f22-842b-5fb91322069b"
	FakeYourOwnDeathOracle = "ad01df89-29fe-44c7-a133-91425f8ff09c"
	RetractionHelixOracle  = "2707f9f4-2b80-47ac-af0a-99fc53af94bf"
	MalakirRebirthOracle   = "a731e87b-8d99-4b64-8ee3-8e540d652366"
	UrzasSagaOracle        = "4c6a0c30-b547-4eff-8ff4-0ca25803c076"
)

func DgCardRef(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

// dgCastOn casts an instant from the active seat's hand at `target`
// and resolves it.
func DgCastOn(t *testing.T, g *game.Game, name, oracle string, target uuid.UUID) {
	t.Helper()
	CastCatalogSpell(t, g, name, "Instant", oracle, DgCardRef(target))
	PassPriorityAroundTable(t, g)
}

func DgTreasuresOf(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Treasure" && c.Controller == controller {
			n++
		}
	}
	return n
}
