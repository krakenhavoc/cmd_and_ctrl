package testsa

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// An opponent's discard with a Plains out creates a 4/4 white Avatar with
// flying for the card's owner.
func TestAjanisLastStandMakesAnAvatarWhenAnOpponentMakesYouDiscardItWithAPlains(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	d.Permanent(d.Victim, "Plains", "Basic Land — Plains", "")
	stand := d.HandCard(d.Victim, "Ajani's Last Stand", "Enchantment", "{2}{W}{W}", effectstest.AjanisLastStandOracle, 0, 0)
	filler := d.Filler(d.Victim)

	d.TheyDiscard(t, stand, filler)

	if !d.Victim.Graveyard.Contains(stand) {
		t.Errorf("the enchantment is in %q, want the graveyard", d.ZoneOf(stand))
	}
	if n := d.Avatars(d.Victim); n != 1 {
		t.Fatalf("Avatars = %d, want 1", n)
	}
	for _, c := range d.G.BattlefieldCardsForEffect() {
		if c.Name == "Avatar" {
			if c.Power != 4 || c.Toughness != 4 || !game.HasKeyword(effectstest.ApaLive(d.G, c.InstanceID), "flying") {
				t.Errorf("the Avatar is %d/%d, flying %v; want a 4/4 flier", c.Power, c.Toughness,
					game.HasKeyword(effectstest.ApaLive(d.G, c.InstanceID), "flying"))
			}
		}
	}
}

// The intervening "if you control a Plains": no Plains, no trigger.
func TestAjanisLastStandNeedsAPlains(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	stand := d.HandCard(d.Victim, "Ajani's Last Stand", "Enchantment", "{2}{W}{W}", effectstest.AjanisLastStandOracle, 0, 0)
	filler := d.Filler(d.Victim)

	d.TheyDiscard(t, stand, filler)

	if n := d.Avatars(d.Victim); n != 0 {
		t.Errorf("Avatars = %d, want 0 without a Plains", n)
	}
}

func TestAjanisLastStandDoesNothingWhenYouDiscardItYourself(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	d.Permanent(d.Caster, "Plains", "Basic Land — Plains", "")
	stand := d.HandCard(d.Caster, "Ajani's Last Stand", "Enchantment", "{2}{W}{W}", effectstest.AjanisLastStandOracle, 0, 0)
	filler := d.Filler(d.Caster)

	d.IDiscard(t, stand, filler)

	if n := d.Avatars(d.Caster); n != 0 {
		t.Errorf("Avatars = %d, want 0 for your own discard", n)
	}
}

// On the battlefield: whenever a creature or planeswalker you control
// dies, you may sacrifice it; if you do, a 4/4 Avatar. A "no" keeps the
// enchantment and makes nothing.
func TestAjanisLastStandSacrificesForAnAvatarWhenYourCreatureDies(t *testing.T) {
	for _, yes := range []bool{true, false} {
		d := effectstest.NewDBOTable(t)
		stand := d.Permanent(d.Caster, "Ajani's Last Stand", "Enchantment", effectstest.AjanisLastStandOracle)
		bear := effectstest.B12Creature(d.G, d.Caster.ID, "Bear", "Creature — Bear", 2, 2)

		d.G.WithWriteLock(func() {
			if err := d.G.DestroyPermanentForEffect(bear); err != nil {
				t.Fatalf("destroy: %v", err)
			}
		})
		effectstest.PassPriorityAroundTable(t, d.G)
		effectstest.AnswerMayChoice(t, d.G, d.Caster.ID, yes)
		effectstest.PassPriorityAroundTable(t, d.G)

		if got := d.Avatars(d.Caster); (got == 1) != yes {
			t.Errorf("answer %v: Avatars = %d", yes, got)
		}
		if d.G.Battlefield.Contains(stand) == yes {
			t.Errorf("answer %v: the enchantment is on the battlefield = %v", yes, d.G.Battlefield.Contains(stand))
		}
	}
}

// An opponent's creature dying does not trigger it.
func TestAjanisLastStandIgnoresAnOpponentsCreature(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	d.Permanent(d.Caster, "Ajani's Last Stand", "Enchantment", effectstest.AjanisLastStandOracle)
	bear := effectstest.B12Creature(d.G, d.Victim.ID, "Bear", "Creature — Bear", 2, 2)

	d.G.WithWriteLock(func() {
		if err := d.G.DestroyPermanentForEffect(bear); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, d.G)

	if effectstest.LatestConfirmFor(d.G, d.Caster.ID) != nil {
		t.Error("an opponent's creature dying asked a question")
	}
}
