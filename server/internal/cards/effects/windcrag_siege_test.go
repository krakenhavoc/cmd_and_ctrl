package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	windcragSiegeOracle            = "64df560a-905f-45d1-bc70-14d99e3112d3"
	adelineResplendentCatharOracle = "38515f89-348b-4cf3-b7bd-1f6fe4ce2fba"
)

// pushWindcragSiege pushes Windcrag Siege directly onto the
// battlefield with its as-enters answer already recorded (or "" for
// unanswered), bypassing the prompt — the real prompt-to-doubling
// path is covered by TestWindcragSiegeMarduDoubledThroughTheRealPrompt.
func pushWindcragSiege(g *game.Game, controller uuid.UUID, option string) uuid.UUID {
	id := pushCatalogPermanent(g, controller, "Windcrag Siege", "Enchantment", windcragSiegeOracle, false)
	if option != "" {
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == id {
					g.Battlefield.Cards[i].ChosenOption = option
					return
				}
			}
		})
	}
	return id
}

// pushAdelineAndAttacker seeds Adeline and one attacking creature, and
// emits the attack. countTokensControlled(g, controller, "Human")
// afterward reads the result.
func pushAdelineAndAttacker(t *testing.T, g *game.Game, controller, defender uuid.UUID) {
	t.Helper()
	pushCatalogPermanent(g, controller, "Adeline, Resplendent Cathar", "Legendary Creature — Human Knight", adelineResplendentCatharOracle, false)
	attacker := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Attacker", TypeLine: "Creature — Soldier",
		Power: 2, Toughness: 2, Owner: controller, Controller: controller,
	})
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventAttack, CardID: attacker, Actor: controller, Target: defender})
	})
	passPriorityAroundTable(t, g)
}

// TestWindcragSiegeMarduDoublesAnAttackCausedTriggerOnly is the
// three-way split #1647 asks for: Mardu doubles an attack-caused
// trigger, Jeskai does not double the same trigger, and Mardu does
// not double a trigger that isn't attack-caused.
func TestWindcragSiegeMarduDoublesAnAttackCausedTriggerOnly(t *testing.T) {
	t.Run("Mardu doubles Adeline's attack trigger", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		pushWindcragSiege(g, me.ID, "Mardu")
		pushAdelineAndAttacker(t, g, me.ID, opp.ID)
		if got := countTokensControlled(g, me.ID, "Human"); got != 6 {
			t.Errorf("Adeline tokens with a Mardu Windcrag Siege = %d, want 6 (doubled)", got)
		}
	})

	t.Run("Jeskai does not double it", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		pushWindcragSiege(g, me.ID, "Jeskai")
		pushAdelineAndAttacker(t, g, me.ID, opp.ID)
		if got := countTokensControlled(g, me.ID, "Human"); got != 3 {
			t.Errorf("Adeline tokens with a Jeskai Windcrag Siege = %d, want 3 (undoubled)", got)
		}
	})

	t.Run("Mardu does not double a non-attack trigger", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		pushWindcragSiege(g, me.ID, "Mardu")
		pushCatalogPermanent(g, me.ID, "Impact Tremors", "Enchantment", impactTremorsOracle, false)
		before := lifeOfOpponents(g)
		g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 1) })
		passPriorityAroundTable(t, g)
		for i, life := range before {
			if got := g.Seats[i+1].Life; got != life-1 {
				t.Errorf("Impact Tremors with a Mardu Windcrag Siege: %d -> %d, want -1 (an ETB is not an attacking cause)", life, got)
			}
		}
	})
}

// TestWindcragSiegeUnansweredChosenOptionMeansNoDoubling pins
// Designation.Active's own rule directly: an empty Card.ChosenOption
// — what a Siege has before its as-enters prompt is answered — matches
// no anchor word, so the doubler is inactive.
func TestWindcragSiegeUnansweredChosenOptionMeansNoDoubling(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushWindcragSiege(g, me.ID, "")
	pushAdelineAndAttacker(t, g, me.ID, opp.ID)
	if got := countTokensControlled(g, me.ID, "Human"); got != 3 {
		t.Errorf("Adeline tokens with an unanswered Windcrag Siege = %d, want 3 (undoubled)", got)
	}
}

// TestWindcragSiegeMarduDoubledThroughTheRealPrompt is the same Mardu
// case end to end: cast from hand, answer the real as-enters prompt,
// and the doubler is active from that answer.
func TestWindcragSiegeMarduDoubledThroughTheRealPrompt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	castSiege(t, g, "Windcrag Siege", windcragSiegeOracle, "Mardu")
	pushAdelineAndAttacker(t, g, me.ID, opp.ID)
	if got := countTokensControlled(g, me.ID, "Human"); got != 6 {
		t.Errorf("Adeline tokens with a cast-and-answered Mardu Windcrag Siege = %d, want 6 (doubled)", got)
	}
}

// TestWindcragSiegeJeskaiCreatesAGoblinWithLifelinkAndHasteUntilEndOfTurn
// — the token really is a 1/1 red Goblin, and it has both granted
// keywords; no Mardu doubler line either.
func TestWindcragSiegeJeskaiCreatesAGoblinWithLifelinkAndHasteUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	castSiege(t, g, "Windcrag Siege", windcragSiegeOracle, "Jeskai")

	advanceToUpkeepOf(t, g, (seat+1)%len(g.Seats))
	before := b30TokensNamed(g, me.ID, "Goblin")
	advanceToUpkeepOf(t, g, seat)
	passPriorityAroundTable(t, g)

	if got := b30TokensNamed(g, me.ID, "Goblin") - before; got != 1 {
		t.Fatalf("Goblin tokens created = %d, want 1", got)
	}
	var token uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Goblin" && c.Controller == me.ID {
			token = c.InstanceID
		}
	}
	if token == uuid.Nil {
		t.Fatal("no Goblin token found")
	}
	if p, tough := effectivePower(t, g, token), effectiveToughness(t, g, token); p != 1 || tough != 1 {
		t.Errorf("the Goblin is %d/%d, want 1/1", p, tough)
	}
	for _, kw := range []string{"lifelink", "haste"} {
		if !effectiveAbilitiesContain(t, g, token, kw) {
			t.Errorf("the Goblin lacks %s", kw)
		}
	}
}
