package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// attached_aura_cost_test.go — #1945: "Sacrifice an Aura attached to
// this creature" (Faunsbane Troll). The enumerator's pool is the Auras
// on the paying creature and nothing else, and with none attached the
// ability is not offered at all (#544).

const oracleFaunsbaneTroll = "b707c131-de13-4d4d-839d-b9f47d62f090"

func roleOn(g *game.Game, p *game.Player, k effects.RoleKind, host uuid.UUID) uuid.UUID {
	r := effects.RoleToken(k)
	r.AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: host}
	r.AttachedAt = 1
	return battlefieldCard(g, p, r)
}

func TestSacrificeAnAuraAttachedToThisOffersOnlyItsOwnAuras(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	troll := battlefieldCard(g, active, game.Card{Name: "Faunsbane Troll", TypeLine: "Creature — Troll", OracleID: oracleFaunsbaneTroll, Power: 4, Toughness: 4})
	bear := battlefieldCard(g, active, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	victim := battlefieldCard(g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)], game.Card{Name: "Victim", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	battlefieldCard(g, active, basic("Forest", "Forest"))
	advanceTo(t, g, game.StepPrecombatMain)

	// A Role on the bear only: the Troll has no Aura of its own.
	strayRole := roleOn(g, active, effects.RoleCursed, bear)
	if got := movesFrom(legal.EnumerateFor(g, active.ID), troll, legal.KindActivate); len(got) != 0 {
		t.Fatalf("no Aura on the Troll, yet the sacrifice ability is offered: %v", labels(got))
	}

	own := roleOn(g, active, effects.RoleMonster, troll)
	moves := legal.EnumerateFor(g, active.ID)
	dispatchAll(t, g, active.ID, moves)
	got := movesFrom(moves, troll, legal.KindActivate)
	if len(got) != 1 {
		t.Fatalf("want one payment (the Troll's own Role), got %v", labels(got))
	}
	if ids := sacrificeIDsOf(t, got[0]); len(ids) != 1 || ids[0] != own.String() {
		t.Errorf("sacrifice_ids = %v, want only the Troll's Role %s (not the bear's %s) for target %s", ids, own, strayRole, victim)
	}
}
