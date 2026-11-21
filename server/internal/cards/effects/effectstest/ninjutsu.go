package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ninjutsu_test.go — the catalog half of #1227: ninjutsu (CR 702.49)
// on Ninja of the Deep Hours, Ingenious Infiltrator, Moonblade Shinobi
// and Prosperous Thief. The engine half — the unblocked-attacker read,
// the attacking entry and the paid-cost record — is in
// game/attacking_entry_test.go.

const (
	NinjaOfTheDeepHoursOracle = "1f3c2b00-0000-4ae1-9650-9553accac52e"
	IngeniousInfiltratorOracl = "354defd2-f63f-4a48-9fd4-2526b3878a84"
	MoonbladeShinobiOracle    = "046e25ee-c96d-4cff-93be-f7e3379d713c"
	ProsperousThiefOracle     = "6d03971a-8365-449a-8fbe-07b5b9bb42dc"
)

// pushNinjaToHand seeds a catalog card into a player's hand WITH a
// body — the zone every ninjutsu ability functions from (CR 702.49a),
// and the power that has to connect once the ninja is attacking.
// pushCatalogHandCard (cycling_test.go) is the P/T-free sibling.
func PushNinjaToHand(p *game.Player, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness,
		Owner: p.ID, Controller: p.ID,
		KnownBy: map[uuid.UUID]bool{p.ID: true},
	})
	return id
}

// ninjutsuBoard sets a combat up the way the keyword wants it: one
// attacker of mine swinging at seat 1, the named ninja in my hand, the
// cursor in the declare-blockers step and `mana` floating. Returns the
// ninja, the attacker and the two seats.
func NinjutsuBoard(t *testing.T, g *game.Game, name, oracle, mana string) (ninja, attacker uuid.UUID, me, opp *game.Player) {
	t.Helper()
	me, opp = g.Seats[0], g.Seats[1]
	attacker = B12Creature(g, me.ID, "Sneaky Rat", "Creature — Rat", 1, 1)
	ninja = PushNinjaToHand(me, name, "Creature — Human Ninja", oracle, 2, 2)
	DeclareAttack(t, g, opp.ID, attacker)
	AdvanceTo(t, g, game.StepDeclareBlockers)
	if mana != "" {
		if err := g.AddManaForEffect(me.ID, uuid.Nil, mana); err != nil {
			t.Fatalf("AddManaForEffect: %v", err)
		}
	}
	return ninja, attacker, me, opp
}

// attackEventsFor counts the EventAttack events naming `id` — the
// CR 506.3c assertion, since a permanent PUT onto the battlefield
// attacking was never declared and must announce nothing.
func AttackEventsFor(g *game.Game, id uuid.UUID) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventAttack && ev.CardID == id {
			n++
		}
	}
	return n
}
