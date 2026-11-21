package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// all_will_be_black_cards_test.go — #2953, the seven cards the
// All Will Be Black deck request (#2910) was missing: Liliana of the
// Dark Realms, Oathsworn Vampire, Phyresis, Pontiff of Blight,
// Timeline Culler, Tragic Slip and Yawgmoth's Vile Offering.

const (
	lilianaDarkRealmsOracle = "92301066-8904-41bc-84bd-4aa286cbab9c"
	OathswornVampireOracle  = "e1fb696f-1df8-4571-9710-0d79001326d3"
	PhyresisOracle          = "958628af-243c-4eb0-8620-5d483b6da768"
	PontiffOfBlightOracle   = "f5a95092-fb71-4b55-bc54-782bd1c603e1"
	TimelineCullerOracle    = "3c307122-bc18-41f1-87d5-ae6b3ca9b8b9"
	TragicSlipOracle        = "fa6e2a47-a8cb-4224-ae3e-f61c90477285"
	VileOfferingOracle      = "740f9ab2-fc94-4abc-82b4-6b53116bf1a4"
)

// awbPadLibrary seeds filler so a shuffle is observable.
func AwbPadLibrary(p *game.Player) {
	for i := 0; i < 12; i++ {
		PushLibraryCardForTest(p, game.Card{Name: "Filler", TypeLine: "Creature — Bear"})
	}
}

// awbSwamps puts n Swamps onto the battlefield under `owner`.
func AwbSwamps(g *game.Game, owner uuid.UUID, n int) []uuid.UUID {
	var ids []uuid.UUID
	for i := 0; i < n; i++ {
		ids = append(ids, GaPut(g, owner, "Swamp", "Basic Land — Swamp", ""))
	}
	return ids
}

// awbLiliana puts Liliana on the battlefield and walks to her
// controller's main phase.
func AwbLiliana(t *testing.T, g *game.Game, loyalty int) (*game.Player, uuid.UUID) {
	t.Helper()
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	lili := PushWalkerForTest(g, me.ID, "Liliana of the Dark Realms", lilianaDarkRealmsOracle, loyalty)
	AdvanceToMainOf(t, g, seat)
	return me, lili
}

func AwbPontiffBoard(t *testing.T, g *game.Game, others int) (*game.Player, uuid.UUID, []uuid.UUID) {
	t.Helper()
	me := g.Seats[0]
	pontiff := PushPermanentForTest(g, me.ID, "Pontiff of Blight", PontiffOfBlightOracle, "Creature — Zombie Cleric")
	var out []uuid.UUID
	for i := 0; i < others; i++ {
		out = append(out, GaPut(g, me.ID, "Bear", "Creature — Bear", ""))
	}
	return me, pontiff, out
}
