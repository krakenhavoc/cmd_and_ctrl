package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// computed_life_cost_test.go — #1594, ADR 0020 Decision 47: a life
// component whose amount is a registered count (AbilityCost.LifeFrom),
// read once at announce (CR 601.2f–g via CR 602.2b) by the one function
// the engine, the legal enumerator and the view share.

const (
	warRoomOracle           = "71c52bf5-2a5d-488e-8b15-7ef290e4b77d"
	MurderousBetrayalOracle = "a2fe1e2b-b621-4406-b5e0-f2eebd12c2ba"
	LurkingEvilOracle       = "81d6b42e-9c41-4d1e-bd28-4662e1fde572"
)

// warRoomDraw is the index of War Room's draw ability in its
// Activated list (the {C} ability is a mana ability, listed apart).
const WarRoomDraw = 0

// clrCommander puts a commander `owner` owns into their command zone
// with the given colour identity. A real imported card (an OracleID),
// so an empty identity is an authoritative "colourless" rather than
// missing data (CR 903.4f, #844).
func ClrCommander(p *game.Player, name string, identity ...string) uuid.UUID {
	id := uuid.New()
	p.Command.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Legendary Creature — Human",
		OracleID: "test-commander-" + name, IsCommander: true,
		ColorIdentity: identity, Owner: p.ID, Controller: p.ID,
	})
	return id
}

// clrWarRoom seeds War Room under `p`, walks to main and floats {3}.
func ClrWarRoom(t *testing.T, g *game.Game, p *game.Player) uuid.UUID {
	t.Helper()
	room := B12Push(g, p.ID, "War Room", "Land", warRoomOracle, 0, 0)
	AdvanceToMain(t, g)
	FillPool(p, 3)
	return room
}

// clrWarRoomCosts activates War Room's draw at `life` and asserts the
// payment and the draw.
func ClrWarRoomCosts(t *testing.T, g *game.Game, me *game.Player, room uuid.UUID, want int) {
	t.Helper()
	life, hand := me.Life, me.Hand.Size()
	B16Activate(t, g, me.ID, room, WarRoomDraw, game.ActivateAbilityParams{})
	if got := life - me.Life; got != want {
		t.Errorf("War Room cost %d life, want %d", got, want)
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}
