package testso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// An opponent's Mind Rot: the Baloth enters under its owner's control
// instead of going to the graveyard, its enters trigger gains 4 life, and
// the discard still counted (the filler is in the graveyard, and a
// "whenever you discard" watcher sees the Baloth's discard land on the
// battlefield).
func TestObstinateBalothEntersWhenAnOpponentMakesYouDiscardIt(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	w := d.Watch()
	baloth := d.HandCard(d.Victim, "Obstinate Baloth", "Creature — Beast", "{2}{G}{G}", effectstest.ObstinateBalothOracle, 4, 4)
	filler := d.Filler(d.Victim)
	life := d.Victim.Life

	d.TheyDiscard(t, baloth, filler)

	c, ok := d.OnBattlefield(baloth)
	if !ok {
		t.Fatalf("the Baloth is in %q, want the battlefield", d.ZoneOf(baloth))
	}
	if c.Controller != d.Victim.ID {
		t.Errorf("the Baloth is controlled by %v, want its owner %v", c.Controller, d.Victim.ID)
	}
	if d.Victim.Graveyard.Contains(baloth) {
		t.Error("the Baloth is also in the graveyard")
	}
	if !d.Victim.Graveyard.Contains(filler) {
		t.Error("the other discarded card did not reach the graveyard")
	}
	if d.Victim.Life != life+4 {
		t.Errorf("life %d -> %d, want +4 from the enters trigger", life, d.Victim.Life)
	}
	var landed game.ZoneKind
	for _, ev := range w.Events {
		if ev.CardID == baloth {
			landed = ev.NewZone
		}
	}
	if landed != game.ZoneBattlefield {
		t.Errorf("the Baloth's discard event names %q, want the battlefield", landed)
	}
}

// Your own spell never triggers it: the cause controller is the discarding
// player, so the Baloth goes to the graveyard and gains nothing.
func TestObstinateBalothStaysInTheGraveyardForYourOwnDiscard(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	baloth := d.HandCard(d.Caster, "Obstinate Baloth", "Creature — Beast", "{2}{G}{G}", effectstest.ObstinateBalothOracle, 4, 4)
	filler := d.Filler(d.Caster)
	life := d.Caster.Life

	d.IDiscard(t, baloth, filler)

	if !d.Caster.Graveyard.Contains(baloth) {
		t.Fatalf("the Baloth is in %q, want the graveyard", d.ZoneOf(baloth))
	}
	if d.Caster.Life != life {
		t.Errorf("life %d -> %d, want unchanged", life, d.Caster.Life)
	}
}

// Painful Quandary asks the caster to discard while its trigger resolves,
// which is an opponent's ability making them discard: the Baloth lands
// (owner decision 4, ADR 0153). Pinned so a change to the pay-unless
// path shows up here.
func TestObstinateBalothEntersForAnOpponentsPainfulQuandary(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	effectstest.B12Push(d.G, d.Victim.ID, "Painful Quandary", "Enchantment", effectstest.PainfulQuandaryOracle, 0, 0)
	baloth := d.HandCard(d.Caster, "Obstinate Baloth", "Creature — Beast", "{2}{G}{G}", effectstest.ObstinateBalothOracle, 4, 4)

	effectstest.CastCatalogSpell(t, d.G, "Filler Spell", "Instant", "not-in-catalog", nil)
	effectstest.PassPriorityAroundTable(t, d.G)
	effectstest.AnswerMayChoice(t, d.G, d.Caster.ID, true)
	pick := effectstest.LatestChooseCardsFor(d.G, d.Caster.ID)
	if pick == nil {
		t.Fatalf("the discard branch asks which card: %+v", d.G.PendingChoices)
	}
	if err := d.G.ResolveChooseCards(pick.ID, d.Caster.ID, []uuid.UUID{baloth}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, d.G)

	c, ok := d.OnBattlefield(baloth)
	if !ok {
		t.Fatalf("the Baloth is in %q, want the battlefield", d.ZoneOf(baloth))
	}
	if c.Controller != d.Caster.ID {
		t.Errorf("the Baloth is controlled by %v, want its owner", c.Controller)
	}
}
