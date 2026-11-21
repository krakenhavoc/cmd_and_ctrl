package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// planeswalker_loyalty_test.go — issues #329 / #334. Before this,
// no catalog card declared a loyalty ability and none could: the
// cost had nowhere to live. These tests are the end-to-end proof
// that a loyalty ability is now an ordinary CR 602 activation.

const (
	TeferiTimeRavelerOracle = "ae7604bb-4818-45a3-960c-cf3d83f15964"
	WanderingEmperorOracle  = "0c7f18d5-36cb-4bc6-a358-443b97666215"
)

// walkerSubtypes is the REAL planeswalker subtype of every card
// pushCatalogWalker seats, from the printed type line. It is a table and
// not a derivation because the subtype is not in the name: "Jace, the
// Mind Sculptor" is a Jace, "Nicol Bolas, God-Pharaoh" a Bolas and "The
// Wandering Emperor" an Emperor. The helper used to build the type line
// as "Legendary Planeswalker — " + name, which gave the first of those
// the subtypes "Jace," and "the Mind Sculptor": a Jace-typed rule
// (Jace's Machinations, "a Jace you control", behold a Jace) silently
// missed the walker (#2803). An unlisted name panics, so a new walker
// has to state its subtype rather than inherit a wrong one.
var walkerSubtypes = map[string]string{
	"Dovin Baan":                   "Dovin",
	"Elspeth, Sun's Champion":      "Elspeth",
	"Gideon of the Trials":         "Gideon",
	"Jace, the Mind Sculptor":      "Jace",
	"Kaya, Orzhov Usurper":         "Kaya",
	"Narset Transcendent":          "Narset",
	"Nicol Bolas, God-Pharaoh":     "Bolas",
	"Tamiyo, Collector of Tales":   "Tamiyo",
	"Teferi, Hero of Dominaria":    "Teferi",
	"Teferi, Master of Time":       "Teferi",
	"Teferi, Temporal Archmage":    "Teferi",
	"Teferi, Time Raveler":         "Teferi",
	"Teferi, Who Slows the Sunset": "Teferi",
	"Tezzeret, Cruel Captain":      "Tezzeret",
	"The Wandering Emperor":        "Emperor",
	"Ugin, Eye of the Storms":      "Ugin",
}

// pushCatalogWalker seats a catalog planeswalker with `loyalty`
// loyalty counters under the active seat's control. Its type line is the
// printed one: see walkerSubtypes.
func PushCatalogWalker(g *game.Game, owner uuid.UUID, name, oracle string, loyalty int) uuid.UUID {
	sub, ok := walkerSubtypes[name]
	if !ok {
		panic("pushCatalogWalker: no real subtype recorded for " + name + "; add it to walkerSubtypes")
	}
	return PushCatalogWalkerWithLine(g, owner, name, "Legendary Planeswalker — "+sub, oracle, loyalty)
}

// pushCatalogWalkerWithLine is pushCatalogWalker with the type line
// spelled out. Use it only where a frozen fixture records a type line
// that is not the printed one (corpusNarset).
func PushCatalogWalkerWithLine(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, loyalty int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   oracle,
		Owner:      owner,
		Controller: owner,
		Counters:   map[string]int{game.CounterLoyalty: loyalty},
	})
	return id
}

func ToMain(t *testing.T, g *game.Game) {
	t.Helper()
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}

func LoyaltyCount(g *game.Game, id uuid.UUID) int {
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return -1
	}
	return c.Counters[game.CounterLoyalty]
}
