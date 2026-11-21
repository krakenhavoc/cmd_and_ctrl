package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cycling_test.go — the catalog half of #660. The engine half (the
// zone dimension, the cost components, the commander case) is in
// game/hand_ability_test.go; what is pinned here is that the six
// cards that used to declare "cycling isn't implemented" now do it,
// that typecycling searches rather than draws, and that Fauna Shaman
// pays a real discard for its tutor.

const (
	RaffinesTowerOracle      = "6e9ef5ef-6aed-4d3e-a59b-9e3dc8740b1b"
	KetriaTriomeOracle       = "6bae00e8-06cf-4ac4-a1cc-757e454109fe"
	ScroungingSkyrayOracle   = "3a46d85b-ce1a-4842-a342-92a5bddb1053"
	MagmakinArtilleristOracl = "900b9409-9c16-414d-8674-2ea42c2415a1"
	FaunaShamanOracle        = "35b8fa77-4e85-418b-b335-cd1af127075c"
)

// pushCatalogHandCard seeds a catalog card into a player's hand.
func PushCatalogHandCard(p *game.Player, name, typeLine, oracle string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: p.ID, Controller: p.ID,
		KnownBy: map[uuid.UUID]bool{p.ID: true},
	})
	return id
}

// cycleFromHand seeds `oracle` into the active seat's hand, floats
// `mana` (a produced-mana string, so colourless pips are spelled
// {C}{C}), and activates the card's cycling ability.
func CycleFromHand(t *testing.T, g *game.Game, name, typeLine, oracle, mana string) (uuid.UUID, *game.Player) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	id := PushCatalogHandCard(me, name, typeLine, oracle)
	if mana != "" {
		if err := g.AddManaForEffect(me.ID, uuid.Nil, mana); err != nil {
			t.Fatalf("AddManaForEffect: %v", err)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("%s: cycle: %v", name, err)
	}
	return id, me
}
