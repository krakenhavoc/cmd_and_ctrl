package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sfPush puts a permanent with the given type line, power and colours on
// the battlefield for `owner`, toughness 5 so combat damage marks rather
// than kills.
func SfPush(g *game.Game, owner uuid.UUID, name, typeLine string, power int, colors []string, keywords ...string) uuid.UUID {
	return ApaPush(g, owner, owner, game.Card{Name: name, TypeLine: typeLine, Power: power, Toughness: 5,
		Colors: colors, Keywords: keywords})
}

// sfCast has `p` cast a card where the game stands — in combat, with
// whoever's priority — and resolves the stack.
func SfCast(t *testing.T, g *game.Game, p *game.Player, card game.Card, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	card.InstanceID = uuid.New()
	card.Owner, card.Controller = p.ID, p.ID
	p.Hand.PushTop(card)
	if err := g.CastSpell(p.ID, card.InstanceID, params); err != nil {
		t.Fatalf("CastSpell %s: %v", card.Name, err)
	}
	PassPriorityAroundTable(t, g)
	return card.InstanceID
}

// sfAttackHold declares the attacks and locks them in, leaving the
// cursor in the declare attackers step: the window in which a defender
// casts Heavy Fog.
func SfAttackHold(t *testing.T, g *game.Game, defender uuid.UUID, attackers ...uuid.UUID) {
	t.Helper()
	AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, a := range attackers {
		if err := g.DeclareAttacker(a, defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	// CR 508.2: the declaration is locked in as a player is next about
	// to receive priority (the state checks).
	g.RunStateChecksForTest()
	if g.Turn.Step != game.StepDeclareAttackers {
		t.Fatalf("left the declare attackers step (at %v)", g.Turn.Step)
	}
}

func SfInstant(name, oracle string) game.Card {
	return game.Card{Name: name, TypeLine: "Instant", OracleID: oracle}
}

// sfCombat walks to the combat damage step and back out of its priority
// boundary, returning how much life each player lost.
func SfCombat(t *testing.T, g *game.Game) map[uuid.UUID]int {
	t.Helper()
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Life
	}
	Pr7bDamageStep(t, g)
	PassPriorityAroundTable(t, g)
	lost := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lost[p.ID] = before[p.ID] - p.Life
	}
	return lost
}
