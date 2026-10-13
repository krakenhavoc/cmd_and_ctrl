package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// obstinate_baloth_test.go — ADR 0153 PR 2 (#3168). The engine half is
// pinned in game/discard_entry_test.go; what is here is the card, and the
// shape every Baloth-family test file shares: an opponent's spell puts it
// onto the battlefield, your own never does.

const obstinateBalothOracle = "ee2ab1ab-be1e-4e56-99f2-7784f7350b41"

// An opponent's Mind Rot: the Baloth enters under its owner's control
// instead of going to the graveyard, its enters trigger gains 4 life, and
// the discard still counted (the filler is in the graveyard, and a
// "whenever you discard" watcher sees the Baloth's discard land on the
// battlefield).
func TestObstinateBalothEntersWhenAnOpponentMakesYouDiscardIt(t *testing.T) {
	d := newDBOTable(t)
	w := d.watch()
	baloth := d.handCard(d.victim, "Obstinate Baloth", "Creature — Beast", "{2}{G}{G}", obstinateBalothOracle, 4, 4)
	filler := d.filler(d.victim)
	life := d.victim.Life

	d.theyDiscard(t, baloth, filler)

	c, ok := d.onBattlefield(baloth)
	if !ok {
		t.Fatalf("the Baloth is in %q, want the battlefield", d.zoneOf(baloth))
	}
	if c.Controller != d.victim.ID {
		t.Errorf("the Baloth is controlled by %v, want its owner %v", c.Controller, d.victim.ID)
	}
	if d.victim.Graveyard.Contains(baloth) {
		t.Error("the Baloth is also in the graveyard")
	}
	if !d.victim.Graveyard.Contains(filler) {
		t.Error("the other discarded card did not reach the graveyard")
	}
	if d.victim.Life != life+4 {
		t.Errorf("life %d -> %d, want +4 from the enters trigger", life, d.victim.Life)
	}
	var landed game.ZoneKind
	for _, ev := range w.events {
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
	d := newDBOTable(t)
	baloth := d.handCard(d.caster, "Obstinate Baloth", "Creature — Beast", "{2}{G}{G}", obstinateBalothOracle, 4, 4)
	filler := d.filler(d.caster)
	life := d.caster.Life

	d.iDiscard(t, baloth, filler)

	if !d.caster.Graveyard.Contains(baloth) {
		t.Fatalf("the Baloth is in %q, want the graveyard", d.zoneOf(baloth))
	}
	if d.caster.Life != life {
		t.Errorf("life %d -> %d, want unchanged", life, d.caster.Life)
	}
}

// Painful Quandary asks the caster to discard while its trigger resolves,
// which is an opponent's ability making them discard: the Baloth lands
// (owner decision 4, ADR 0153). Pinned so a change to the pay-unless
// path shows up here.
func TestObstinateBalothEntersForAnOpponentsPainfulQuandary(t *testing.T) {
	d := newDBOTable(t)
	b12Push(d.g, d.victim.ID, "Painful Quandary", "Enchantment", painfulQuandaryOracle, 0, 0)
	baloth := d.handCard(d.caster, "Obstinate Baloth", "Creature — Beast", "{2}{G}{G}", obstinateBalothOracle, 4, 4)

	castCatalogSpell(t, d.g, "Filler Spell", "Instant", "not-in-catalog", nil)
	passPriorityAroundTable(t, d.g)
	answerMayChoice(t, d.g, d.caster.ID, true)
	pick := latestChooseCardsFor(d.g, d.caster.ID)
	if pick == nil {
		t.Fatalf("the discard branch asks which card: %+v", d.g.PendingChoices)
	}
	if err := d.g.ResolveChooseCards(pick.ID, d.caster.ID, []uuid.UUID{baloth}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, d.g)

	c, ok := d.onBattlefield(baloth)
	if !ok {
		t.Fatalf("the Baloth is in %q, want the battlefield", d.zoneOf(baloth))
	}
	if c.Controller != d.caster.ID {
		t.Errorf("the Baloth is controlled by %v, want its owner", c.Controller)
	}
}
