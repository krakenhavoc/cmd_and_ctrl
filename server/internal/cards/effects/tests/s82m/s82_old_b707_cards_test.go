package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b707_cards_test.go — S82 OLD batch 707 (#5621, Urza's Destiny):
// Opalescence, Metalworker, Powder Keg, Aura Thief, Phyrexian Negator. Each
// test plays the real catalog entry.

const (
	oldB707Opalescence = "59489b46-9d02-4f3c-bcd0-884e7605e9a5"
	oldB707Metalworker = "9ed9accb-9f0c-40fa-aeea-2d1e1c9d1bd1"
	oldB707PowderKeg   = "ca1144c2-49a3-49b4-9085-3094141769ea"
	oldB707AuraThief   = "cf0936ea-2741-43a4-ad6a-a1258b54945b"
	oldB707Negator     = "f22fcd17-bc24-41df-81ce-6d11208861b9"
)

func TestOldB707EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB707Opalescence: "Opalescence", oldB707Metalworker: "Metalworker", oldB707PowderKeg: "Powder Keg",
		oldB707AuraThief: "Aura Thief", oldB707Negator: "Phyrexian Negator",
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

func oldB707Has(types []string, want string) bool {
	for _, s := range types {
		if s == want {
			return true
		}
	}
	return false
}

// --- Opalescence -------------------------------------------------------------

func TestOldB707OpalescenceAnimatesEveryOtherNonAuraEnchantmentAtItsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	op := effectstest.PushCatalogPermanent(g, me.ID, "Opalescence", "Enchantment", oldB707Opalescence, false)
	mine := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "My Shrine", TypeLine: "Enchantment", ManaCost: "{2}{G}"})
	theirs := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Their Shrine", TypeLine: "Enchantment", ManaCost: "{4}{W}{W}"})
	aura := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Their Aura", TypeLine: "Enchantment — Aura", ManaCost: "{1}{W}"})
	rock := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{3}"})

	if p, tough := effectstest.EffectivePTOf(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("my enchantment is %d/%d, want 3/3 (its mana value)", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, theirs); p != 6 || tough != 6 {
		t.Errorf("their enchantment is %d/%d, want 6/6: Opalescence affects every player's", p, tough)
	}
	for name, id := range map[string]uuid.UUID{"my enchantment": mine, "their enchantment": theirs} {
		types := effectstest.EffectiveTypes(t, g, id)
		if !oldB707Has(types, "Creature") || !oldB707Has(types, "Enchantment") {
			t.Errorf("%s has types %v, want creature in addition to enchantment", name, types)
		}
	}
	if oldB707Has(effectstest.EffectiveTypes(t, g, aura), "Creature") {
		t.Error("an Aura became a creature")
	}
	if oldB707Has(effectstest.EffectiveTypes(t, g, rock), "Creature") {
		t.Error("a non-enchantment became a creature")
	}
	if oldB707Has(effectstest.EffectiveTypes(t, g, op), "Creature") {
		t.Error("Opalescence animated itself: it says each OTHER enchantment")
	}
}

func TestOldB707OpalescenceStopsWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	op := effectstest.PushCatalogPermanent(g, me.ID, "Opalescence", "Enchantment", oldB707Opalescence, false)
	mine := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "My Shrine", TypeLine: "Enchantment", ManaCost: "{2}{G}"})
	if !oldB707Has(effectstest.EffectiveTypes(t, g, mine), "Creature") {
		t.Fatal("the Shrine was not animated")
	}
	effectstest.P1Destroy(t, g, op)
	if oldB707Has(effectstest.EffectiveTypes(t, g, mine), "Creature") {
		t.Error("the Shrine is still a creature with Opalescence gone")
	}
}

// --- Metalworker -------------------------------------------------------------

func TestOldB707MetalworkerAddsTwoColorlessForEachArtifactCardRevealed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.EmptyHand(g, me)
	id := effectstest.PushCatalogPermanent(g, me.ID, "Metalworker", "Artifact Creature — Construct", oldB707Metalworker, false)
	effectstest.Unsick(g, id)
	a1 := effectstest.AvB1Hand(me, game.Card{Name: "Rock A", TypeLine: "Artifact"})
	a2 := effectstest.AvB1Hand(me, game.Card{Name: "Golem B", TypeLine: "Artifact Creature — Golem", Power: 2, Toughness: 2})
	bear := effectstest.AvB1Hand(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{RevealIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("revealed a card that is not an artifact")
	}
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{RevealIDs: []uuid.UUID{a1, a2}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	got := 0
	for _, tok := range me.ManaPool {
		if tok.Color == "C" {
			got += tok.N()
		}
	}
	if got != 4 || me.Hand.Size() != 3 {
		t.Errorf("pool has %d {C}, hand %d; want 4 and the three cards still in hand", got, me.Hand.Size())
	}
	if c, _ := effectstest.BattlefieldCard(g, id); !c.Tapped {
		t.Error("Metalworker was not tapped for the cost")
	}
}

func TestOldB707MetalworkerRevealingNothingAddsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Metalworker", "Artifact Creature — Construct", oldB707Metalworker, false)
	effectstest.Unsick(g, id)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	for _, tok := range me.ManaPool {
		if tok.N() > 0 {
			t.Errorf("{%s} was added for no cards", tok.Color)
		}
	}
}

// --- Powder Keg --------------------------------------------------------------

func TestOldB707PowderKegGrowsAFuseCounterEachUpkeepIfYouWish(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	keg := effectstest.PushCatalogPermanent(g, me.ID, "Powder Keg", "Artifact", oldB707PowderKeg, false)
	for _, yes := range []bool{true, false, true} {
		g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventBeginUpkeep, Actor: me.ID}) })
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if n := effectstest.AangCounterCount(g, keg, "fuse"); n != 2 {
		t.Errorf("fuse counters = %d, want 2 (two yeses, one decline)", n)
	}
	// It does not grow on an opponent's upkeep.
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventBeginUpkeep, Actor: g.Seats[1].ID}) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceTriggerPrompt, me.ID) != nil {
		t.Error("the Keg offered a counter on an opponent's upkeep")
	}
}

func TestOldB707PowderKegDestroysArtifactsAndCreaturesOfExactlyThatManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	keg := effectstest.PushCatalogPermanent(g, me.ID, "Powder Keg", "Artifact", oldB707PowderKeg, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == keg {
				g.Battlefield.Cards[i].Counters = map[string]int{"fuse": 2}
			}
		}
	})
	bear2 := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Two Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	mine2 := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Two Rock", TypeLine: "Artifact", ManaCost: "{2}"})
	bear3 := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Three Bear", TypeLine: "Creature — Bear", ManaCost: "{2}{G}", Power: 3, Toughness: 3})
	rock3 := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Three Rock", TypeLine: "Artifact", ManaCost: "{3}"})
	ench2 := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Two Shrine", TypeLine: "Enchantment", ManaCost: "{1}{W}"})
	land := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.Pr7Activate(t, g, me.ID, keg, 0, game.ActivateAbilityParams{Strict: true})
	for name, id := range map[string]uuid.UUID{"their mana value 2 creature": bear2, "my mana value 2 artifact": mine2} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s survived", name)
		}
	}
	for name, id := range map[string]uuid.UUID{"a mana value 3 creature": bear3, "a mana value 3 artifact": rock3,
		"a mana value 2 enchantment": ench2, "a land": land} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s was destroyed", name)
		}
	}
	if g.Battlefield.Contains(keg) {
		t.Error("the Keg was not sacrificed")
	}
}

// --- Aura Thief --------------------------------------------------------------

func TestOldB707AuraThiefTakesEveryEnchantmentWhenItDiesAuraStaysAttached(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	thief := effectstest.PushCatalogPermanent(g, me.ID, "Aura Thief", "Creature — Illusion", oldB707AuraThief, false)
	shrine := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Their Shrine", TypeLine: "Enchantment", ManaCost: "{2}{W}"})
	host := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	aura := effectstest.Enchant(t, g, "Their Aura", "", host)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == aura {
				g.Battlefield.Cards[i].Owner, g.Battlefield.Cards[i].Controller = opp.ID, opp.ID
			}
		}
	})
	rock := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Their Rock", TypeLine: "Artifact", ManaCost: "{2}"})
	effectstest.P1Destroy(t, g, thief)
	effectstest.PassPriorityAroundTable(t, g)
	controls := func(id uuid.UUID) bool {
		c := effectstest.FindBattlefieldCardByID(g, id)
		return c != nil && c.Controller == me.ID
	}
	if !controls(shrine) {
		t.Error("the Shrine was not taken")
	}
	if !controls(aura) {
		t.Error("the Aura was not taken")
	}
	if c := effectstest.FindBattlefieldCardByID(g, aura); c == nil || !c.IsAttachedTo(host) {
		t.Error("the Aura moved: it must stay attached to what it enchants")
	}
	if controls(rock) || controls(host) {
		t.Error("a non-enchantment was taken")
	}
}

// --- Phyrexian Negator -------------------------------------------------------

func TestOldB707NegatorSacrificesThatManyPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	negator := effectstest.B12Push(g, me.ID, "Phyrexian Negator", "Creature — Phyrexian Horror", oldB707Negator, 5, 5)
	a := effectstest.B12Permanent(g, me.ID, "Rock A", "Artifact")
	b := effectstest.B12Permanent(g, me.ID, "Rock B", "Artifact")
	c := effectstest.B12Permanent(g, me.ID, "Rock C", "Artifact")
	d := effectstest.B12Permanent(g, me.ID, "Rock D", "Artifact")
	effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "", effectstest.B16TargetCard(negator))
	effectstest.PassPriorityAroundTable(t, g)
	prompts := 0
	for _, ch := range g.PendingChoices {
		if ch != nil && ch.Kind == game.PendingChoiceSacrifice && ch.Chooser == me.ID {
			prompts++
		}
	}
	if prompts != 3 {
		t.Fatalf("3 damage: its controller owes three sacrifices, %d prompts queued", prompts)
	}
	effectstest.AnswerSacrifice(t, g, me.ID, a)
	effectstest.AnswerSacrifice(t, g, me.ID, b)
	effectstest.AnswerSacrifice(t, g, me.ID, c)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) || g.Battlefield.Contains(c) || !g.Battlefield.Contains(d) {
		t.Error("three permanents of their choice are gone, the fourth stays")
	}
	if !g.Battlefield.Contains(negator) {
		t.Error("3 damage does not kill a 5/5")
	}
	if effectstest.SacrificeChoiceFor(g, me.ID) != nil {
		t.Error("nothing more is owed")
	}
}

func TestOldB707NegatorStillTriggersWhenTheDamageKillsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	negator := effectstest.B12Push(g, me.ID, "Phyrexian Negator", "Creature — Phyrexian Horror", oldB707Negator, 2, 2)
	a := effectstest.B12Permanent(g, me.ID, "Rock A", "Artifact")
	b := effectstest.B12Permanent(g, me.ID, "Rock B", "Artifact")
	effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "", effectstest.B16TargetCard(negator))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(negator) {
		t.Fatal("the 2/2 Negator survived a Bolt")
	}
	// Only two permanents remain to sacrifice: the owed three are trimmed to what exists.
	effectstest.AnswerSacrifice(t, g, me.ID, a)
	effectstest.AnswerSacrifice(t, g, me.ID, b)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("the dead Negator's trigger did not make its controller sacrifice")
	}
}
