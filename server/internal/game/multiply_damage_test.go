package game

import (
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// multiply_damage_test.go — ADR 0108 §3 (#1890): damage doubled or
// tripled by a resolved spell or ability (CR 614.1a, 616.1, 120.8,
// 611.2c), its "next time" form (CR 615.8's instance), and the
// controller filter on choose_source.

// pushToughCreature is pushColouredCreature with toughness high enough
// that a doubled hit never kills it, so the damage marked on it is the
// whole answer.
func pushToughCreature(g *Game, owner *Player, name string) uuid.UUID {
	id := pushColouredCreature(g, owner, name, []string{"R"})
	findBattlefieldCard(g, id).Toughness = 20
	return id
}

// multiplier registers d (Controller defaulting to `you`) and fails the
// test if nothing was written.
func multiplier(t *testing.T, g *Game, d DamageMultiplier) {
	t.Helper()
	g.WithWriteLock(func() {
		if !g.MultiplyDamageForEffect(d) {
			t.Fatalf("no multiplier registered for %+v", d)
		}
	})
}

func dealToPlayer(t *testing.T, g *Game, source, player uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(source, player, n); err != nil {
			t.Fatal(err)
		}
	})
}

func dealToCreature(t *testing.T, g *Game, source, card uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(source, card, n); err != nil {
			t.Fatal(err)
		}
	})
}

// Insult: "If a source you control would deal damage this turn, it
// deals double that damage instead." Your sources only, to anything.
func TestMultiplyDamageDoublesYourSourcesOnly(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushToughCreature(g, me, "Mine")
	theirs := pushToughCreature(g, opp, "Theirs")
	multiplier(t, g, DamageMultiplier{Controller: me.ID, Factor: 2, Sources: DamageSourcesYours, Label: "Insult"})
	oppLife, myLife := lifeOf(g, opp.ID), lifeOf(g, me.ID)

	dealToPlayer(t, g, mine, opp.ID, 3)
	dealToCreature(t, g, mine, theirs, 2)
	dealToPlayer(t, g, theirs, me.ID, 3)
	if got := lifeOf(g, opp.ID); got != oppLife-6 {
		t.Errorf("opponent life %d, want %d: your 3 is doubled", got, oppLife-6)
	}
	if got := damageOn(g, theirs); got != 4 {
		t.Errorf("their creature has %d damage, want 4", got)
	}
	if got := lifeOf(g, me.ID); got != myLife-3 {
		t.Errorf("your life %d, want %d: an opponent's source is not yours", got, myLife-3)
	}
}

// The Insult // Injury ruling: two Insults are ×4. The two commute, so
// the affected player is not asked to order them (CR 616.1 has one
// answer).
func TestTwoMultipliersStackWithoutAnOrderPrompt(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushToughCreature(g, me, "Mine")
	multiplier(t, g, DamageMultiplier{Controller: me.ID, Factor: 2, Sources: DamageSourcesYours, Label: "Insult"})
	multiplier(t, g, DamageMultiplier{Controller: me.ID, Factor: 3, Sources: DamageSourcesYours,
		Recipients: DamageRecipientsOpponentsAndTheirPermanents, Label: "Isengard Unleashed"})
	start := lifeOf(g, opp.ID)
	dealToPlayer(t, g, mine, opp.ID, 2)
	if len(g.PendingChoices) != 0 {
		t.Fatalf("an order prompt was asked between two multipliers: %+v", g.PendingChoices[0])
	}
	if got := lifeOf(g, opp.ID); got != start-12 {
		t.Errorf("opponent life %d, want %d: 2 doubled and tripled", got, start-12)
	}
}

// Isengard Unleashed: "to an opponent or a permanent an opponent
// controls … triple". Your own creature and you are dealt damage as
// printed.
func TestMultiplyDamageToOpponentsAndTheirPermanents(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushToughCreature(g, me, "Mine")
	other := pushToughCreature(g, me, "Other of mine")
	theirs := pushToughCreature(g, opp, "Theirs")
	multiplier(t, g, DamageMultiplier{Controller: me.ID, Factor: 3, Sources: DamageSourcesYours,
		Recipients: DamageRecipientsOpponentsAndTheirPermanents, Label: "Isengard Unleashed"})
	oppLife, myLife := lifeOf(g, opp.ID), lifeOf(g, me.ID)
	dealToPlayer(t, g, mine, opp.ID, 2)
	dealToCreature(t, g, mine, theirs, 2)
	dealToCreature(t, g, mine, other, 2)
	dealToPlayer(t, g, mine, me.ID, 2)
	if got := lifeOf(g, opp.ID); got != oppLife-6 {
		t.Errorf("opponent life %d, want %d", got, oppLife-6)
	}
	if got := damageOn(g, theirs); got != 6 {
		t.Errorf("opponent's creature has %d damage, want 6", got)
	}
	if got := damageOn(g, other); got != 2 {
		t.Errorf("your creature has %d damage, want 2", got)
	}
	if got := lifeOf(g, me.ID); got != myLife-2 {
		t.Errorf("your life %d, want %d", got, myLife-2)
	}
}

// Goblin Goliath: "to an opponent … to that player" — players only.
func TestMultiplyDamageToOpponentsIsPlayersOnly(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushToughCreature(g, me, "Mine")
	theirs := pushToughCreature(g, opp, "Theirs")
	multiplier(t, g, DamageMultiplier{Controller: me.ID, Factor: 2, Sources: DamageSourcesYours,
		Recipients: DamageRecipientsOpponents, Label: "Goblin Goliath"})
	start := lifeOf(g, opp.ID)
	dealToPlayer(t, g, mine, opp.ID, 2)
	dealToCreature(t, g, mine, theirs, 2)
	if got := lifeOf(g, opp.ID); got != start-4 {
		t.Errorf("opponent life %d, want %d", got, start-4)
	}
	if got := damageOn(g, theirs); got != 2 {
		t.Errorf("opponent's creature has %d damage, want 2: it is not an opponent", got)
	}
}

// Lightning, Army of One: ANY source dealing damage to that player or a
// permanent that player controls, until your next turn — the damaged
// player's own sources included, and nobody else.
func TestMultiplyDamageToOnePlayerAndTheirPermanentsFromAnySource(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	theirs := pushToughCreature(g, opp, "Theirs")
	thirds := pushToughCreature(g, third, "Third's")
	multiplier(t, g, DamageMultiplier{Controller: me.ID, Factor: 2, Recipients: DamageRecipientsPlayerAndTheirPermanents,
		Player: opp.ID, UntilNextTurnOf: me.ID, Label: "Lightning, Army of One"})
	oppLife, thirdLife := lifeOf(g, opp.ID), lifeOf(g, third.ID)
	dealToPlayer(t, g, thirds, opp.ID, 2)
	dealToCreature(t, g, theirs, theirs, 1)
	dealToPlayer(t, g, theirs, third.ID, 2)
	if got := lifeOf(g, opp.ID); got != oppLife-4 {
		t.Errorf("that player's life %d, want %d", got, oppLife-4)
	}
	if got := damageOn(g, theirs); got != 2 {
		t.Errorf("that player's creature has %d damage, want 2: its own damage is doubled too", got)
	}
	if got := lifeOf(g, third.ID); got != thirdLife-2 {
		t.Errorf("another player's life %d, want %d", got, thirdLife-2)
	}
	if d := g.ScopedEffects[0].Duration; d.Kind != UntilYourNextTurn || d.Player != me.ID {
		t.Errorf("duration %+v, want until the controller's next turn", d)
	}
}

// Blind Fury: "If a creature would deal combat damage to a creature this
// turn" — combat only, creature to creature.
func TestMultiplyCombatDamageCreatureToCreature(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := pushToughCreature(g, me, "Attacker")
	blocker := pushToughCreature(g, opp, "Blocker")
	multiplier(t, g, DamageMultiplier{Controller: me.ID, Factor: 2, Sources: DamageSourcesCreatures,
		Recipients: DamageRecipientsCreatures, CombatOnly: true, Label: "Blind Fury"})
	start := lifeOf(g, opp.ID)
	g.WithWriteLock(func() {
		g.markCombatDamageOnCardLocked(blocker, 3, attacker, CombatStepRegular)
		g.markCombatDamageToPlayerLocked(opp.ID, attacker, 3, CombatStepRegular)
	})
	dealToCreature(t, g, attacker, blocker, 1)
	if got := damageOn(g, blocker); got != 7 {
		t.Errorf("blocker has %d damage, want 7: 3 combat doubled, 1 non-combat not", got)
	}
	if got := lifeOf(g, opp.ID); got != start-3 {
		t.Errorf("player life %d, want %d: a player is not a creature", got, start-3)
	}
}

// "The next time that source would deal damage this turn" (Desperate
// Gambit, CR 615.8): the next INSTANCE, every event of it, then nothing.
// A source that would deal 0 deals none (CR 120.8) and does not spend it.
func TestNextTimeMultiplierCoversOneInstance(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	trampler := pushToughCreature(g, me, "Trampler")
	blocker := pushToughCreature(g, opp, "Blocker")
	g.WithWriteLock(func() {
		ref, zone, _ := g.DamageSourceRefLocked(trampler)
		if !g.MultiplyDamageForEffect(DamageMultiplier{Controller: me.ID, Factor: 2, Source: ref, SourceZone: zone,
			Next: true, Label: "Desperate Gambit"}) {
			t.Fatal("no multiplier")
		}
	})
	start := lifeOf(g, opp.ID)
	dealToPlayer(t, g, trampler, opp.ID, 0)
	if g.ScopedEffects[0].Mods[0].SpentInstance != 0 {
		t.Fatal("0 damage spent the next-time multiplier")
	}
	g.WithWriteLock(func() {
		g.markCombatDamageOnCardLocked(blocker, 2, trampler, CombatStepRegular)
		g.markCombatDamageToPlayerLocked(opp.ID, trampler, 3, CombatStepRegular)
	})
	if got := damageOn(g, blocker); got != 4 {
		t.Errorf("blocker has %d damage, want 4", got)
	}
	if got := lifeOf(g, opp.ID); got != start-6 {
		t.Errorf("life %d, want %d: the trampled-over damage is the same instance", got, start-6)
	}
	g.WithWriteLock(func() {
		g.beginEventBatchLocked()
		if len(g.ScopedEffects) != 0 {
			t.Fatalf("spent multiplier outlived its instance: %+v", g.ScopedEffects)
		}
	})
	dealToPlayer(t, g, trampler, opp.ID, 2)
	if got := lifeOf(g, opp.ID); got != start-8 {
		t.Errorf("life %d, want %d: the next instance is dealt normally", got, start-8)
	}
}

// Two instructions in one resolution are two instances (CR 608.2c): the
// "next time" multiplier doubles only the first.
func TestNextTimeMultiplierSkipsALaterInstanceInTheSameBatch(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushToughCreature(g, me, "Source")
	g.WithWriteLock(func() {
		ref, zone, _ := g.DamageSourceRefLocked(src)
		g.MultiplyDamageForEffect(DamageMultiplier{Controller: me.ID, Factor: 2, Source: ref, SourceZone: zone, Next: true})
	})
	start := lifeOf(g, opp.ID)
	g.WithWriteLock(func() {
		_ = g.DealDamageToPlayerForEffect(src, opp.ID, 2)
		_ = g.DealDamageToPlayerForEffect(src, opp.ID, 2)
	})
	if got := lifeOf(g, opp.ID); got != start-6 {
		t.Errorf("life %d, want %d: 4 then 2", got, start-6)
	}
}

// CR 614.1a: a multiplier is a replacement, not a prevention effect, so
// "damage can't be prevented" leaves it alone. CR 611.2c: "a source you
// control" is read as the damage would be dealt, so a creature stolen
// after the effect began is yours now.
func TestMultiplierIsNotPreventionAndReadsControlLive(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stolen := pushToughCreature(g, opp, "Stolen")
	multiplier(t, g, DamageMultiplier{Controller: me.ID, Factor: 2, Sources: DamageSourcesYours, Label: "Insult"})
	g.WithWriteLock(func() {
		if !g.DamageCantBePreventedThisTurnForEffect(uuid.Nil, "Insult") {
			t.Fatal("no gate")
		}
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(stolen),
			[]Mod{SetControllerMod(me.ID)}, IndefiniteDuration(), "steal")
		g.RecomputeLayersIfStaleLocked()
	})
	start := lifeOf(g, opp.ID)
	dealToPlayer(t, g, stolen, opp.ID, 3)
	if got := lifeOf(g, opp.ID); got != start-6 {
		t.Errorf("life %d, want %d", got, start-6)
	}
}

// Registration refuses what the vocabulary cannot say, and a field of
// the kind on another kind.
func TestMultiplyDamageRegistrationRefusals(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	for name, d := range map[string]DamageMultiplier{
		"factor 1":                      {Controller: me.ID, Factor: 1},
		"next with no source":           {Controller: me.ID, Factor: 2, Next: true},
		"player recipients, no player":  {Controller: me.ID, Factor: 2, Recipients: DamageRecipientsPlayerAndTheirPermanents},
		"a player for other recipients": {Controller: me.ID, Factor: 2, Recipients: DamageRecipientsOpponents, Player: me.ID},
		"unknown sources":               {Controller: me.ID, Factor: 2, Sources: "theirs"},
	} {
		g.WithWriteLock(func() {
			if g.MultiplyDamageForEffect(d) {
				t.Errorf("%s: registered", name)
			}
		})
	}
	for _, m := range []Mod{
		{Kind: ModPreventDamage, Amount: 1, Sources: DamageSourcesYours},
		{Kind: ModPreventCombatDamage, Next: true},
	} {
		if multiplyDamageModProblem(m) == "" {
			t.Errorf("%q carrying a multiplyDamage field passes", m.Kind)
		}
	}
}

// Restore refuses a multiplier vocabulary value this binary does not
// know (ADR 0041 P4), and round-trips a live one, spent "next time"
// state included.
func TestMultiplyDamageSnapshotRoundTripAndRefusal(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushToughCreature(g, me, "Source")
	multiplier(t, g, DamageMultiplier{Controller: me.ID, Factor: 3, Sources: DamageSourcesYours,
		Recipients: DamageRecipientsOpponentsAndTheirPermanents, Label: "Isengard Unleashed"})
	g.WithWriteLock(func() {
		ref, zone, _ := g.DamageSourceRefLocked(src)
		g.MultiplyDamageForEffect(DamageMultiplier{Controller: me.ID, Factor: 2, Source: ref, SourceZone: zone, Next: true})
		_ = g.DealDamageToPlayerForEffect(src, opp.ID, 1)
	})
	snap := g.CaptureSnapshot()
	raw, err := json.Marshal(snap)
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
	if len(restored.ScopedEffects) != 2 || restored.ScopedEffects[1].Mods[0].SpentInstance == 0 {
		t.Fatalf("restored records %+v", restored.ScopedEffects)
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

// The banner line (ADR 0108 §3 decision 4).
func TestDamageMultiplierLines(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	multiplier(t, g, DamageMultiplier{Controller: me.ID, Factor: 2, Sources: DamageSourcesYours, Label: "Insult"})
	multiplier(t, g, DamageMultiplier{Controller: me.ID, Factor: 3, Sources: DamageSourcesYours,
		Recipients: DamageRecipientsOpponentsAndTheirPermanents, Label: "Isengard Unleashed"})
	multiplier(t, g, DamageMultiplier{Controller: me.ID, Factor: 2, Recipients: DamageRecipientsPlayerAndTheirPermanents,
		Player: opp.ID, UntilNextTurnOf: me.ID, Label: "Lightning"})
	var lines []string
	g.WithWriteLock(func() { lines = g.DamageMultiplierLines() })
	want := []string{
		me.Name + "'s sources deal double damage this turn — Insult",
		me.Name + "'s sources deal triple damage to " + me.Name + "'s opponents and their permanents this turn — Isengard Unleashed",
		"Sources deal double damage to " + opp.Name + " and their permanents until " + me.Name + "'s next turn — Lightning",
	}
	if !slices.Equal(lines, want) {
		t.Errorf("lines\n%q\nwant\n%q", lines, want)
	}
}

// Desperate Gambit's "choose a source you control": the choose_source
// candidates narrowed to the chooser's own sources (CR 108.4a: a card
// nobody controls answers by its owner).
func TestChooseSourceControllerFilter(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushToughCreature(g, me, "Mine")
	theirs := pushToughCreature(g, opp, "Theirs")
	var all, own []uuid.UUID
	g.WithWriteLock(func() {
		all = g.DamageSourceCandidatesLocked(nil)
		own = g.damageSourceCandidatesLocked(nil, me.ID)
	})
	if !slices.Contains(all, theirs) || !slices.Contains(all, mine) {
		t.Fatalf("unfiltered candidates %v miss a permanent", all)
	}
	if slices.Contains(own, theirs) || !slices.Contains(own, mine) {
		t.Errorf("filtered candidates %v: want yours only", own)
	}
	for _, id := range own {
		var c Card
		var ok bool
		g.WithWriteLock(func() { c, ok = g.LookupCardForEffect(id) })
		if ok && c.Owner != me.ID && c.Controller != me.ID {
			t.Errorf("candidate %s is neither controlled nor owned by the chooser", c.Name)
		}
	}
}

// Impulsive Maneuvers' losing flip: "the next time that creature would
// deal COMBAT damage this turn, prevent that damage". Non-combat damage
// passes the shield by and does not spend it.
func TestNextDamageShieldCombatOnly(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := pushToughCreature(g, me, "Attacker")
	g.WithWriteLock(func() {
		ref, zone, _ := g.DamageSourceRefLocked(attacker)
		g.PreventNextDamageFromSourceForEffect(NextDamageShield{Controller: me.ID, Source: ref, SourceZone: zone,
			CombatOnly: true, Label: "Impulsive Maneuvers"})
	})
	start := lifeOf(g, opp.ID)
	dealToPlayer(t, g, attacker, opp.ID, 2)
	if got := lifeOf(g, opp.ID); got != start-2 {
		t.Fatalf("life %d, want %d: non-combat damage is not prevented", got, start-2)
	}
	if len(g.ScopedEffects) != 1 || g.ScopedEffects[0].Mods[0].SpentInstance != 0 {
		t.Fatalf("non-combat damage spent the shield: %+v", g.ScopedEffects)
	}
	g.WithWriteLock(func() { g.markCombatDamageToPlayerLocked(opp.ID, attacker, 3, CombatStepRegular) })
	if got := lifeOf(g, opp.ID); got != start-2 {
		t.Errorf("life %d, want %d: the combat damage is prevented", got, start-2)
	}
}
