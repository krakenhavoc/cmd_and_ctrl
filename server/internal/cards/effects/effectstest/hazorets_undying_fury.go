package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hazorets_undying_fury_test.go — Hazoret's Undying Fury (#4021): the top
// four exiled and cast free, any number of them, each capped at mana value 5
// on its own, then the caster's lands held through their next untap step.

const hazoretUndyingFuryOracle = "af4d3489-6bba-4538-928c-50c9a7073813"

// furyCast casts the sorcery and settles it, leaving the exile offered.
func FuryCast(t *testing.T, g *game.Game) {
	t.Helper()
	CastCatalogSpell(t, g, "Hazoret's Undying Fury", "Sorcery", hazoretUndyingFuryOracle, nil)
	PassPriorityAroundTable(t, g)
}

// furyLibrary seeds a library of exactly these four cards, so the shuffle
// cannot change which ones the exile takes.
func FuryLibrary(me *game.Player) (five, six, two, forest uuid.UUID) {
	me.Library.Cards = nil
	five = PushLibraryCardForTest(me, game.Card{Name: "Five", TypeLine: "Artifact", ManaCost: "{5}"})
	six = PushLibraryCardForTest(me, game.Card{Name: "Six", TypeLine: "Artifact", ManaCost: "{6}"})
	two = PushLibraryCardForTest(me, game.Card{Name: "Two", TypeLine: "Artifact", ManaCost: "{2}"})
	forest = PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	return
}
