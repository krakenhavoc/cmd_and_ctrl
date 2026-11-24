package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB28StickerOID   = "f55a5b8a-ee73-4248-90c7-d2456928448e"
	oldB28RoseNobleOID = "cf66067b-8f5e-46f6-9d61-cffb8875d858"
	oldB28EvinOID      = "f688f839-c7c2-45a9-8c60-9fdb030803e6"
	oldB28MeepOID      = "ff2448d0-7a45-45be-bf02-19650f490b54"
	oldB28DenialOID    = "3adc3af3-5086-4d61-b38c-c1dddeaeebe6"
	oldB28LapseOID     = "bbfb3e4a-b389-4391-8141-13b68c0ef2e0"
	oldB28AmyPondOID   = "aa7d3e99-fc14-4f62-bf00-c0428c11e4f2"
	oldB28Bear         = "Creature — Bear"
)

func TestOldB28CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB28StickerOID:   "Sticker sheet",
		oldB28RoseNobleOID: "Rose Noble", oldB28EvinOID: "Evin, Waterdeep Opportunist",
		oldB28MeepOID: "The Meep", oldB28DenialOID: "Stubborn Denial", oldB28LapseOID: "Memory Lapse",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB28Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, tgh int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: p, Toughness: tgh, Owner: owner, Controller: owner,
	})
}

func oldB28Treasure(g *game.Game, owner uuid.UUID) uuid.UUID {
	return oldB28Push(g, owner, "Treasure", "Token Artifact — Treasure", "", 0, 0)
}

// oldB28Ward casts a Doom Blade at `victim` from `caster` and passes until the
// ward trigger asks, reporting whether it asked the caster for payment.
func oldB28WardPrompt(t *testing.T, g *game.Game, caster *game.Player, victim uuid.UUID) (spell uuid.UUID, asked bool) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	spell = effectstest.CastAtWardedCreature(t, g, caster, victim)
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, caster.ID) && effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceConfirm, caster.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	return spell, effectstest.HasPayUnlessFor(g, caster.ID) || effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceConfirm, caster.ID) != nil
}

// --- the Secret Lair sticker sheet ----------------------------------------------

func TestOldB28TheStickerSheetHasNoAbilities(t *testing.T) {
	for _, id := range []string{oldB28StickerOID} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Fatalf("%s is not registered", id)
		}
		if len(spec.Triggered) != 0 || len(spec.Activated) != 0 || len(spec.Static) != 0 || len(spec.ManaAbilities) != 0 {
			t.Errorf("%s declares abilities, but the card has no rules text", spec.Name)
		}
	}
}

// --- Rose Noble -----------------------------------------------------------------

func TestOldB28RoseNobleDrawsForDoctorsAndDoctorsCompanionCreatures(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		draws                  bool
	}{
		{"The Doctor", "Legendary Creature — Time Lord Doctor", "", true},
		{"Amy Pond", "Legendary Creature — Human", oldB28AmyPondOID, true},
		{"Rose Noble", "Legendary Creature — Human", oldB28RoseNobleOID, true},
		{"Plain Bear", oldB28Bear, "", false},
		{"Plain Sorcery", "Sorcery", "", false},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		oldB28Push(g, me.ID, "Rose Noble", "Legendary Creature — Human", oldB28RoseNobleOID, 2, 3)
		effectstest.AdvanceToMain(t, g)
		before := me.Hand.Size()
		effectstest.CastCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		want := before
		if tc.draws {
			want++
		}
		if got := me.Hand.Size(); got != want {
			t.Errorf("%s: hand %d, want %d", tc.name, got, want)
		}
	}
}

func TestOldB28RoseNobleHasWardTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rose := oldB28Push(g, me.ID, "Rose Noble", "Legendary Creature — Human", oldB28RoseNobleOID, 2, 3)
	spell, asked := oldB28WardPrompt(t, g, opp, rose)
	if !asked {
		t.Fatal("no ward payment prompt for the spell's controller")
	}
	if c := effectstest.PendingOfKind(g, game.PendingChoicePayUnless); c == nil || c.PayCost != "{2}" {
		t.Errorf("the ward cost is %+v, want {2}", c)
	}
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rose) || !opp.Graveyard.Contains(spell) {
		t.Error("declining the ward must counter the spell")
	}
}

// --- Evin, Waterdeep Opportunist --------------------------------------------------

func TestOldB28EvinGetsPlusTwoForEachTreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	evin := oldB28Push(g, me.ID, "Evin, Waterdeep Opportunist", "Legendary Creature — Human Rogue", oldB28EvinOID, 2, 4)
	oldB28Treasure(g, me.ID)
	oldB28Treasure(g, me.ID)
	oldB28Treasure(g, opp.ID)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if p, tg := effectstest.EffectivePower(t, g, evin), effectstest.EffectiveToughness(t, g, evin); p != 6 || tg != 4 {
		t.Errorf("Evin is %d/%d, want 6/4 (two Treasures of mine)", p, tg)
	}
	if !effectstest.HasEffectiveKeyword(t, g, evin, "menace") {
		t.Error("Evin has no menace")
	}
}

func oldB28Treasures(g *game.Game, controller uuid.UUID, tapped bool) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == controller && c.Name == "Treasure" && effects.IsToken(c) && c.Tapped == tapped {
				n++
			}
		}
	})
	return n
}

func TestOldB28EvinMakesATappedTreasureOnceEachTurnWhenAnyoneSacrificesACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB28Push(g, me.ID, "Evin, Waterdeep Opportunist", "Legendary Creature — Human Rogue", oldB28EvinOID, 2, 4)
	mine := oldB28Push(g, me.ID, "My Bear", oldB28Bear, "", 2, 2)
	theirs := oldB28Push(g, opp.ID, "Their Bear", oldB28Bear, "", 2, 2)
	rock := oldB28Treasure(g, opp.ID)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(rock); err != nil {
			t.Fatalf("sacrificing a Treasure: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB28Treasures(g, me.ID, true); got != 0 {
		t.Errorf("sacrificing a non-creature made %d Treasures", got)
	}
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(theirs); err != nil {
			t.Fatalf("sacrificing their creature: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB28Treasures(g, me.ID, true); got != 1 {
		t.Fatalf("tapped Treasures = %d, want 1 (a creature any player sacrifices)", got)
	}
	if got := oldB28Treasures(g, opp.ID, true) + oldB28Treasures(g, opp.ID, false); got != 0 {
		t.Errorf("the opponent has %d Treasures; the sacrificed one was theirs and nothing gave them another", got)
	}
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(mine); err != nil {
			t.Fatalf("sacrificing my creature: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB28Treasures(g, me.ID, true); got != 1 {
		t.Errorf("tapped Treasures = %d, want still 1: it triggers only once each turn", got)
	}
}

func TestOldB28EvinsWardIsTheSacrificeOfACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	evin := oldB28Push(g, me.ID, "Evin, Waterdeep Opportunist", "Legendary Creature — Human Rogue", oldB28EvinOID, 2, 4)
	// The opponent controls no creature, so the ward can't be paid and the spell is countered.
	spell, _ := oldB28WardPrompt(t, g, opp, evin)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(evin) || !opp.Graveyard.Contains(spell) {
		t.Error("a payer with no creature to sacrifice must have the spell countered")
	}
}

// --- The Meep -------------------------------------------------------------------

func TestOldB28TheMeepHasWardPayThreeLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	meep := oldB28Push(g, me.ID, "The Meep", "Legendary Creature — Alien", oldB28MeepOID, 0, 4)
	spell, asked := oldB28WardPrompt(t, g, opp, meep)
	if !asked {
		t.Fatal("no ward payment prompt for the spell's controller")
	}
	life := effectstest.LifeOf(g, opp.ID)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceConfirm, opp.ID)
	if c == nil {
		t.Fatal("the ward asked no yes/no question")
	}
	if err := g.ResolveConfirm(c.ID, opp.ID, false); err != nil {
		t.Fatalf("declining: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LifeOf(g, opp.ID) != life {
		t.Error("declining the ward cost life")
	}
	if !g.Battlefield.Contains(meep) || !opp.Graveyard.Contains(spell) {
		t.Error("declining the ward must counter the spell")
	}
}

func TestOldB28TheMeepSacrificeSetsBasePowerAndToughnessToTheSacrificedManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	meep := oldB28Push(g, me.ID, "The Meep", "Legendary Creature — Alien", oldB28MeepOID, 0, 4)
	fodder := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Costly Bear", TypeLine: oldB28Bear, ManaCost: "{2}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	other := oldB28Push(g, me.ID, "Other Bear", oldB28Bear, "", 1, 1)
	theirs := oldB28Push(g, opp.ID, "Their Bear", oldB28Bear, "", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, meep)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID); c != nil {
		if err := g.ResolveOwnPermanents(c.ID, me.ID, []uuid.UUID{fodder}); err != nil {
			t.Fatalf("choosing the creature to sacrifice: %v", err)
		}
	} else if c := effectstest.PendingOfKind(g, game.PendingChoiceTriggerPrompt); c != nil {
		t.Fatalf("an unexpected trigger prompt: %+v", c)
	} else {
		t.Fatal("the Meep asked for no creature to sacrifice")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Fatal("the creature was not sacrificed")
	}
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	for name, id := range map[string]uuid.UUID{"The Meep": meep, "the other Bear": other} {
		if p, tg := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id); p != 3 || tg != 3 {
			t.Errorf("%s is %d/%d, want 3/3", name, p, tg)
		}
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 2 {
		t.Errorf("an opposing creature has power %d, want 2", p)
	}
}

// --- Stubborn Denial ------------------------------------------------------------

func oldB28Denial(t *testing.T, ferocious bool) (g *game.Game, me *game.Player, victim uuid.UUID, asked bool) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	them := g.Seats[1]
	if ferocious {
		oldB28Push(g, them.ID, "Big Bear", oldB28Bear, "", 4, 4)
	} else {
		oldB28Push(g, them.ID, "Small Bear", oldB28Bear, "", 3, 3)
	}
	effectstest.ToMainForCost(t, g)
	victim = effectstest.CastCatalogSpell(t, g, "Victim Sorcery", "Sorcery", "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	effectstest.CastFromSeat(t, g, them, "Stubborn Denial", "Instant", oldB28DenialOID, []game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, me.ID) && g.Stack.Contains(victim); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	return g, me, victim, effectstest.HasPayUnlessFor(g, me.ID)
}

func TestOldB28StubbornDenialCountersUnlessOneIsPaid(t *testing.T) {
	g, me, victim, asked := oldB28Denial(t, false)
	if !asked {
		t.Fatal("no {1} payment prompt without ferocious")
	}
	if !g.Stack.Contains(victim) {
		t.Fatal("the spell resolved while the {1} was unanswered")
	}
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	if g.Stack.Contains(victim) {
		t.Error("declining the tax must counter the spell")
	}
}

func TestOldB28StubbornDenialWithFerociousJustCounters(t *testing.T) {
	g, _, victim, asked := oldB28Denial(t, true)
	if asked {
		t.Error("ferocious must counter the spell outright, with no payment offered")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Stack.Contains(victim) {
		t.Error("the spell was not countered")
	}
}

func TestOldB28StubbornDenialRefusesACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	them := g.Seats[1]
	effectstest.ToMainForCost(t, g)
	victim := effectstest.CastCatalogSpell(t, g, "Their Bear", oldB28Bear, "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	id := uuid.New()
	them.Hand.PushTop(game.Card{InstanceID: id, Name: "Stubborn Denial", TypeLine: "Instant", OracleID: oldB28DenialOID, Owner: them.ID, Controller: them.ID})
	if err := g.CastSpell(them.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err == nil {
		t.Error("Stubborn Denial targeted a creature spell")
	}
}

// --- Memory Lapse ---------------------------------------------------------------

func TestOldB28MemoryLapsePutsTheCounteredSpellOnTopOfItsOwnersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	victim := effectstest.CastCatalogSpell(t, g, "Victim Sorcery", "Sorcery", "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	effectstest.CastFromSeat(t, g, them, "Memory Lapse", "Instant", oldB28LapseOID, []game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Stack.Contains(victim) {
		t.Fatal("the spell was not countered")
	}
	if me.Graveyard.Contains(victim) {
		t.Error("the spell went to the graveyard")
	}
	if size := me.Library.Size(); size == 0 || me.Library.Cards[size-1].InstanceID != victim {
		t.Error("the spell is not on top of its owner's library")
	}
	if them.Library.Contains(victim) {
		t.Error("the spell went into the caster's library")
	}
}
