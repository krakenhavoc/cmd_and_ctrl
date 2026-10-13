package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ajanis_last_stand_test.go — ADR 0153 PR 2 (#3168).

const ajanisLastStandOracle = "17c4807c-158e-4a86-90f1-d8fb2ebe892a"

func (d dboTable) avatars(p *game.Player) int { return d.countNamedUnder(p, "Avatar") }

// An opponent's discard with a Plains out creates a 4/4 white Avatar with
// flying for the card's owner.
func TestAjanisLastStandMakesAnAvatarWhenAnOpponentMakesYouDiscardItWithAPlains(t *testing.T) {
	d := newDBOTable(t)
	d.permanent(d.victim, "Plains", "Basic Land — Plains", "")
	stand := d.handCard(d.victim, "Ajani's Last Stand", "Enchantment", "{2}{W}{W}", ajanisLastStandOracle, 0, 0)
	filler := d.filler(d.victim)

	d.theyDiscard(t, stand, filler)

	if !d.victim.Graveyard.Contains(stand) {
		t.Errorf("the enchantment is in %q, want the graveyard", d.zoneOf(stand))
	}
	if n := d.avatars(d.victim); n != 1 {
		t.Fatalf("Avatars = %d, want 1", n)
	}
	for _, c := range d.g.BattlefieldCardsForEffect() {
		if c.Name == "Avatar" {
			if c.Power != 4 || c.Toughness != 4 || !game.HasKeyword(apaLive(d.g, c.InstanceID), "flying") {
				t.Errorf("the Avatar is %d/%d, flying %v; want a 4/4 flier", c.Power, c.Toughness,
					game.HasKeyword(apaLive(d.g, c.InstanceID), "flying"))
			}
		}
	}
}

// The intervening "if you control a Plains": no Plains, no trigger.
func TestAjanisLastStandNeedsAPlains(t *testing.T) {
	d := newDBOTable(t)
	stand := d.handCard(d.victim, "Ajani's Last Stand", "Enchantment", "{2}{W}{W}", ajanisLastStandOracle, 0, 0)
	filler := d.filler(d.victim)

	d.theyDiscard(t, stand, filler)

	if n := d.avatars(d.victim); n != 0 {
		t.Errorf("Avatars = %d, want 0 without a Plains", n)
	}
}

func TestAjanisLastStandDoesNothingWhenYouDiscardItYourself(t *testing.T) {
	d := newDBOTable(t)
	d.permanent(d.caster, "Plains", "Basic Land — Plains", "")
	stand := d.handCard(d.caster, "Ajani's Last Stand", "Enchantment", "{2}{W}{W}", ajanisLastStandOracle, 0, 0)
	filler := d.filler(d.caster)

	d.iDiscard(t, stand, filler)

	if n := d.avatars(d.caster); n != 0 {
		t.Errorf("Avatars = %d, want 0 for your own discard", n)
	}
}

// On the battlefield: whenever a creature or planeswalker you control
// dies, you may sacrifice it; if you do, a 4/4 Avatar. A "no" keeps the
// enchantment and makes nothing.
func TestAjanisLastStandSacrificesForAnAvatarWhenYourCreatureDies(t *testing.T) {
	for _, yes := range []bool{true, false} {
		d := newDBOTable(t)
		stand := d.permanent(d.caster, "Ajani's Last Stand", "Enchantment", ajanisLastStandOracle)
		bear := b12Creature(d.g, d.caster.ID, "Bear", "Creature — Bear", 2, 2)

		d.g.WithWriteLock(func() {
			if err := d.g.DestroyPermanentForEffect(bear); err != nil {
				t.Fatalf("destroy: %v", err)
			}
		})
		passPriorityAroundTable(t, d.g)
		answerMayChoice(t, d.g, d.caster.ID, yes)
		passPriorityAroundTable(t, d.g)

		if got := d.avatars(d.caster); (got == 1) != yes {
			t.Errorf("answer %v: Avatars = %d", yes, got)
		}
		if d.g.Battlefield.Contains(stand) == yes {
			t.Errorf("answer %v: the enchantment is on the battlefield = %v", yes, d.g.Battlefield.Contains(stand))
		}
	}
}

// An opponent's creature dying does not trigger it.
func TestAjanisLastStandIgnoresAnOpponentsCreature(t *testing.T) {
	d := newDBOTable(t)
	d.permanent(d.caster, "Ajani's Last Stand", "Enchantment", ajanisLastStandOracle)
	bear := b12Creature(d.g, d.victim.ID, "Bear", "Creature — Bear", 2, 2)

	d.g.WithWriteLock(func() {
		if err := d.g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	passPriorityAroundTable(t, d.g)

	if latestConfirmFor(d.g, d.caster.ID) != nil {
		t.Error("an opponent's creature dying asked a question")
	}
}
