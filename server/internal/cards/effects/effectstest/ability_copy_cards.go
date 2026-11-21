package effectstest

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ability_copy_cards_test.go — the five cards #1223 ships, and the
// two seams they prove: copying an ability on the stack (CR 707.10)
// and a trigger that reads the event that fired it (CR 603.2).

const (
	StrionicResonatorOracle   = "cf751552-156f-4f81-ac94-9814dce099f9"
	LithoformEngineOracle     = "0bf299da-1854-4153-baea-3cee2eb01ee8"
	RingsOfBrighthearthOracle = "bbf9494c-c4bb-4d36-98fe-8387846b342e"
	ScrapTrawlerOracle        = "164f3f85-21fc-40b7-9871-4f303ba98428"
	CloudstoneCurioOracle     = "5cd2fd32-4da2-40eb-b003-c0b9a9ec91c1"
	// Basalt Monolith prints both kinds of activation on one card,
	// which is what makes it the fixture for the Rings: {3}: Untap is
	// a CR 602 ability that uses the stack, and {T}: Add {C}{C}{C} is
	// a CR 605 mana ability that never does.
	AcBasaltMonolithOracle = "6b8cf2a0-b045-4d91-9d91-c602d40c6237"
)

// --- helpers -------------------------------------------------------

// acItemLabelled is the id of the one stack item carrying `label`.
func AcItemLabelled(g *game.Game, label string) uuid.UUID {
	var out uuid.UUID
	g.WithWriteLock(func() {
		for id, it := range g.StackMeta {
			if it != nil && it.Label == label {
				out = id
			}
		}
	})
	return out
}

// acLegalAbilityTargets is the set a card's ability clause offers
// right now, as a membership map.
func AcLegalAbilityTargets(g *game.Game, chooser uuid.UUID, oracle string, index int) map[uuid.UUID]bool {
	ab := game.ActivatedAbilitiesForCard(game.Card{OracleID: oracle})[index]
	out := map[uuid.UUID]bool{}
	g.WithWriteLock(func() {
		lt := g.LegalTargetsForEffect(game.SourceChooser(chooser), ab.Targets)
		for _, id := range lt.Cards {
			out[id] = true
		}
	})
	return out
}

// acCopiesOnStack is how many CR 707.10 copies are on the stack.
func AcCopiesOnStack(g *game.Game) int {
	n := 0
	g.WithWriteLock(func() {
		for _, it := range g.StackMeta {
			if it != nil && it.IsCopy {
				n++
			}
		}
	})
	return n
}

// acItemOnStack is the stack item whose label starts with `prefix`.
func AcItemOnStack(g *game.Game, prefix string) *game.StackItem {
	var out *game.StackItem
	g.WithWriteLock(func() {
		for _, it := range g.StackMeta {
			if it != nil && strings.HasPrefix(it.Label, prefix) {
				out = it
			}
		}
	})
	return out
}

// acCountEvents is how many events of a kind the log holds.
func AcCountEvents(g *game.Game, kind game.EventKind) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == kind {
			n++
		}
	}
	return n
}

// acCountTriggerAnnouncements is how many EventTrigger breadcrumbs
// carry a label starting with `prefix`.
func AcCountTriggerAnnouncements(g *game.Game, prefix string) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventTrigger && strings.HasPrefix(ev.Label, prefix) {
			n++
		}
	}
	return n
}

// acCopyOnStack is the one CR 707.10 copy on the stack, or nil.
func AcCopyOnStack(g *game.Game) *game.StackItem {
	var out *game.StackItem
	g.WithWriteLock(func() {
		for _, it := range g.StackMeta {
			if it != nil && it.IsCopy {
				out = it
			}
		}
	})
	return out
}

// acPassUntilCopy passes priority until a copy appears on the stack,
// and stops there. Passing on would resolve the copy, and an item
// that has resolved is not on the stack to be asserted about.
func AcPassUntilCopy(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 16 && AcCopiesOnStack(g) == 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority iter %d: %v", i, err)
		}
	}
}
