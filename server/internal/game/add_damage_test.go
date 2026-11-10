package game

import (
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"
)

// add_damage_test.go — #4468: ModAddDamage, "if a source you control
// would deal noncombat damage to a permanent or player this turn, it
// deals that much damage plus X instead" (Taii Wakeen, Perfect Shot):
// the additive sibling of the ADR 0108 §3 multiplier.

func adder(t *testing.T, g *Game, d DamageAddition) {
	t.Helper()
	g.WithWriteLock(func() {
		if !g.AddDamageForEffect(d) {
			t.Fatalf("no addition registered for %+v", d)
		}
	})
}

// Your sources' noncombat damage, to players and permanents alike, is
// increased; combat damage and an opponent's source are not.
func TestAddDamageNoncombatFromYourSources(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushToughCreature(g, me, "Mine")
	theirs := pushToughCreature(g, opp, "Theirs")
	adder(t, g, DamageAddition{Controller: me.ID, Amount: 2, Sources: DamageSourcesYours, NoncombatOnly: true, Label: "Taii Wakeen"})
	oppLife, myLife := lifeOf(g, opp.ID), lifeOf(g, me.ID)

	dealToPlayer(t, g, mine, opp.ID, 3)
	dealToCreature(t, g, mine, theirs, 1)
	dealToCreature(t, g, mine, mine, 1)
	dealToPlayer(t, g, theirs, me.ID, 3)
	if got := lifeOf(g, opp.ID); got != oppLife-5 {
		t.Errorf("opponent life %d, want %d: 3 plus 2", got, oppLife-5)
	}
	if got := damageOn(g, theirs); got != 3 {
		t.Errorf("their creature has %d damage, want 3: 1 plus 2", got)
	}
	if got := damageOn(g, mine); got != 3 {
		t.Errorf("your own creature has %d damage, want 3: \"a permanent\" is any permanent", got)
	}
	if got := lifeOf(g, me.ID); got != myLife-3 {
		t.Errorf("your life %d, want %d: an opponent's source is not yours", got, myLife-3)
	}

	g.WithWriteLock(func() {
		g.markCombatDamageOnCardLocked(theirs, 2, mine, CombatStepRegular)
		g.markCombatDamageToPlayerLocked(opp.ID, mine, 2, CombatStepRegular)
	})
	if got := damageOn(g, theirs); got != 5 {
		t.Errorf("their creature has %d damage, want 5: combat damage is not increased", got)
	}
	if got := lifeOf(g, opp.ID); got != oppLife-7 {
		t.Errorf("opponent life %d, want %d: combat damage is not increased", got, oppLife-7)
	}
}

// Two additions commute (no CR 616 prompt); beside a multiplier the
// affected player orders them.
func TestAddDamageOrderingAgainstAMultiplier(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushToughCreature(g, me, "Mine")
	adder(t, g, DamageAddition{Controller: me.ID, Amount: 1, Sources: DamageSourcesYours, NoncombatOnly: true})
	adder(t, g, DamageAddition{Controller: me.ID, Amount: 2, Sources: DamageSourcesYours, NoncombatOnly: true})
	start := lifeOf(g, opp.ID)
	dealToPlayer(t, g, mine, opp.ID, 1)
	if len(g.PendingChoices) != 0 {
		t.Fatalf("an order prompt was asked between two additions: %+v", g.PendingChoices[0])
	}
	if got := lifeOf(g, opp.ID); got != start-4 {
		t.Errorf("opponent life %d, want %d: 1 plus 1 plus 2", got, start-4)
	}

	multiplier(t, g, DamageMultiplier{Controller: me.ID, Factor: 2, Sources: DamageSourcesYours, Label: "Insult"})
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(mine, opp.ID, 1) })
	if len(g.PendingChoices) == 0 {
		t.Errorf("no CR 616.1 order prompt beside a multiplier")
	}
}

// All of the damage prevented: nothing to add to (the Taii ruling,
// CR 120.8).
func TestAddDamageDoesNotApplyToZero(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushToughCreature(g, me, "Mine")
	adder(t, g, DamageAddition{Controller: me.ID, Amount: 3, Sources: DamageSourcesYours, NoncombatOnly: true})
	start := lifeOf(g, opp.ID)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(mine, opp.ID, 0) })
	if got := lifeOf(g, opp.ID); got != start {
		t.Errorf("opponent life %d, want %d: 0 damage is no damage to add to", got, start)
	}
}

func TestAddDamageRegistrationRefusals(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	for name, d := range map[string]DamageAddition{
		"plus 0":                        {Controller: me.ID, Amount: 0},
		"negative":                      {Controller: me.ID, Amount: -1},
		"combat and noncombat":          {Controller: me.ID, Amount: 1, CombatOnly: true, NoncombatOnly: true},
		"player recipients, no player":  {Controller: me.ID, Amount: 1, Recipients: DamageRecipientsPlayerAndTheirPermanents},
		"a player for other recipients": {Controller: me.ID, Amount: 1, Recipients: DamageRecipientsOpponents, Player: me.ID},
		"unknown sources":               {Controller: me.ID, Amount: 1, Sources: "theirs"},
	} {
		g.WithWriteLock(func() {
			if g.AddDamageForEffect(d) {
				t.Errorf("%s: registered", name)
			}
		})
	}
	for _, m := range []Mod{
		{Kind: ModPreventDamage, Amount: 1, NoncombatOnly: true},
		{Kind: ModMultiplyDamage, Amount: 2, NoncombatOnly: true},
		{Kind: ModAddDamage, Amount: 1, Next: true},
		{Kind: ModAddDamage, Amount: 1, Objects: []ObjectRef{{ID: g.Seats[1].ID}}},
	} {
		if multiplyDamageModProblem(m) == "" {
			t.Errorf("%q %+v passes", m.Kind, m)
		}
	}
}

// Restore round-trips a live addition and refuses one with a value this
// binary does not know (ADR 0041 P4).
func TestAddDamageSnapshotRoundTripAndRefusal(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushToughCreature(g, me, "Source")
	adder(t, g, DamageAddition{Controller: me.ID, Amount: 2, Sources: DamageSourcesYours, NoncombatOnly: true, Label: "Taii Wakeen"})
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var back GameSnapshot
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatal(err)
	}
	restored, err := back.RestoreStrict()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	start := lifeOf(restored, opp.ID)
	dealToPlayer(t, restored, src, opp.ID, 1)
	if got := lifeOf(restored, opp.ID); got != start-3 {
		t.Errorf("restored: opponent life %d, want %d", got, start-3)
	}
	bad := strings.Replace(string(raw), `"sources":"yours"`, `"sources":"theirs"`, 1)
	var refused GameSnapshot
	if err := json.Unmarshal([]byte(bad), &refused); err != nil {
		t.Fatal(err)
	}
	if _, err := refused.RestoreStrict(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Errorf("restore of an unknown sources value: %v, want ErrUnknownEffectKey", err)
	}
}

func TestAddDamageBannerLine(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	adder(t, g, DamageAddition{Controller: me.ID, Amount: 2, Sources: DamageSourcesYours, NoncombatOnly: true, Label: "Taii Wakeen, Perfect Shot"})
	var lines []string
	g.WithWriteLock(func() { lines = g.DamageMultiplierLines() })
	want := []string{me.Name + "'s sources deal 2 more noncombat damage this turn — Taii Wakeen, Perfect Shot"}
	if !slices.Equal(lines, want) {
		t.Errorf("lines\n%q\nwant\n%q", lines, want)
	}
}
