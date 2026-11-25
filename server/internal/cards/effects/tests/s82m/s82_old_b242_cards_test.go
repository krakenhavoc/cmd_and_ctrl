package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b242_cards_test.go — Amonkhet, batch 242 (S82 older sets, #5535).
// Every test plays the real catalog entry through the engine.

const (
	oldB242RuthlessSniper = "02a506bd-203f-421c-b72e-e2ba77e922aa"
	oldB242Luxa           = "1eb64899-8865-48cc-bd58-529b1294ec99"
	oldB242Soulstinger    = "31505942-52c4-4485-9241-c9592e6839e4"
	oldB242Manticore      = "3ec1b67a-5ff8-4ce7-8388-d11a23f76f0e"
	oldB242LilianasInf    = "7fe67cef-38c8-4071-86df-71053818d908"
	oldB242QuarryHauler   = "a55bdb73-e17a-414d-b27f-dc53af2f345c"
	oldB242Decimator      = "c56e6715-80bc-4f9d-86ef-7c31b1ef0a17"
	oldB242Javelineer     = "dd2b59b4-281e-4b51-8998-9b20a846ddbf"
	oldB242Greatmaw       = "ee152ffa-9c39-48ed-9aca-e53c8ebb0231"
	oldB242Pyramid        = "efc66685-94cf-4d3f-82e0-9967c6ab2e5c"
	oldB242Exemplar       = "f732873e-0b34-4ccc-8b4c-9f02463ba57b"
	oldB242Bear           = "Creature — Bear"
)

func TestOldB242EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		oldB242RuthlessSniper, oldB242Luxa, oldB242Soulstinger, oldB242Manticore, oldB242LilianasInf, oldB242QuarryHauler,
		oldB242Decimator, oldB242Javelineer, oldB242Greatmaw, oldB242Pyramid, oldB242Exemplar,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

func oldB242Minus(g *game.Game, id uuid.UUID) int {
	return effectstest.CountersOn(g, id, game.CounterMinusOne)
}

func TestOldB242ExemplarPutsThreeOnYourCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", oldB242Bear, 5, 5)
	effectstest.RfdCast(t, g, "Exemplar of Strength", "Creature — Human Warrior", oldB242Exemplar, 4, 4)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := oldB242Minus(g, bear); n != 3 {
		t.Errorf("the creature has %d -1/-1 counters, want 3", n)
	}
}

func TestOldB242ExemplarAttackShedsACounterAndGainsLife(t *testing.T) {
	for _, tc := range []struct {
		counters, wantLife int
	}{{2, 1}, {0, 0}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := effectstest.SeatsForCombat(g)
		ex := effectstest.B12Push(g, me.ID, "Exemplar of Strength", "Creature — Human Warrior", oldB242Exemplar, 4, 4)
		if tc.counters > 0 {
			g.WithWriteLock(func() { _ = g.AddCounterForEffect(ex, game.CounterMinusOne, tc.counters) })
		}
		life := effectstest.LifeOf(g, me.ID)
		effectstest.AttackWithTrigger(t, g, ex, opp.ID, uuid.Nil)
		if n := oldB242Minus(g, ex); n != max(tc.counters-1, 0) {
			t.Errorf("%d counters: left with %d", tc.counters, n)
		}
		if got := effectstest.LifeOf(g, me.ID) - life; got != tc.wantLife {
			t.Errorf("%d counters: life changed by %d, want %d", tc.counters, got, tc.wantLife)
		}
	}
}

func TestOldB242QuarryHaulerPutsOrRemovesPerKind(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	target := effectstest.BigBear(g, g.Seats[1].ID, 2, false)
	effectstest.S82SetCounters(g, target, map[string]int{"charge": 3, "stun": 1})
	effectstest.RfdCast(t, g, "Quarry Hauler", "Creature — Camel", oldB242QuarryHauler, 4, 3)
	effectstest.PickCard(t, g, me.ID, target)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0) // charge: put another
	effectstest.AnswerOptionPick(t, g, me.ID, 1) // stun: remove it
	if n := effectstest.CountersOn(g, target, "charge"); n != 4 {
		t.Errorf("charge = %d, want 4", n)
	}
	if n := effectstest.CountersOn(g, target, "stun"); n != 0 {
		t.Errorf("stun = %d, want 0", n)
	}
	if effectstest.LatestOptionPickFor(g, me.ID) != nil {
		t.Error("asked more than once per kind")
	}
}

func TestOldB242ManticoreCountersMineAndShootsAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", oldB242Bear, 5, 5)
	life := opp.Life
	effectstest.RfdCast(t, g, "Manticore of the Gauntlet", "Creature — Manticore", oldB242Manticore, 5, 4)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if n := oldB242Minus(g, bear); n != 1 {
		t.Errorf("my creature has %d -1/-1 counters, want 1", n)
	}
	if opp.Life != life-3 {
		t.Errorf("opponent life = %d, want %d", opp.Life, life-3)
	}
}

func TestOldB242SoulstingerDiesWithItsCountersOntoATarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	foe := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB242Bear, 9, 9)
	sting := effectstest.RfdCast(t, g, "Soulstinger", "Creature — Scorpion Demon", oldB242Soulstinger, 4, 5)
	effectstest.PickCard(t, g, me.ID, sting)
	effectstest.PassPriorityAroundTable(t, g)
	if n := oldB242Minus(g, sting); n != 2 {
		t.Fatalf("Soulstinger has %d counters, want 2", n)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(sting, game.CounterMinusOne, 3) })
	g.RunStateChecksForTest()
	effectstest.PickCard(t, g, me.ID, foe)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceTriggerPrompt, me.ID)
	if prompt == nil {
		t.Fatal("the dies trigger did not ask whether to put the counters")
	}
	if err := g.ResolveTriggerPrompt(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveTriggerPrompt: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := oldB242Minus(g, foe); n != 5 {
		t.Errorf("the target has %d -1/-1 counters, want the 5 Soulstinger died with", n)
	}
}

func TestOldB242GreatmawRemovesACounterWhenItGetsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	ally := effectstest.B12Creature(g, me.ID, "My Bear", oldB242Bear, 5, 5)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ally, game.CounterMinusOne, 2) })
	maw := effectstest.RfdCast(t, g, "Defiant Greatmaw", "Creature — Hippo", oldB242Greatmaw, 4, 5)
	effectstest.PickCard(t, g, me.ID, maw)    // its own enters: two counters on itself
	effectstest.PassPriorityAroundTable(t, g) // the enters trigger resolves; then the Greatmaw trigger asks for another creature
	for effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickCard(t, g, me.ID, ally)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := oldB242Minus(g, maw); n != 2 {
		t.Errorf("Greatmaw has %d counters, want 2", n)
	}
	if n := oldB242Minus(g, ally); n != 1 {
		t.Errorf("the other creature has %d counters, want 1 (one removed)", n)
	}
}

func TestOldB242LilianasInfluenceShrinksOpponentsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", oldB242Bear, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB242Bear, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Liliana's Influence", "Sorcery", oldB242LilianasInf, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := oldB242Minus(g, theirs); n != 1 {
		t.Errorf("their creature has %d counters, want 1", n)
	}
	if n := oldB242Minus(g, mine); n != 0 {
		t.Errorf("my creature has %d counters, want 0", n)
	}
}

func TestOldB242LuxaRiverShrineBricksThenGainsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	shrine := effectstest.PushCatalogPermanent(g, me.ID, "Luxa River Shrine", "Artifact", oldB242Luxa, false)
	if err := g.ActivateCatalogAbility(me.ID, shrine, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the second ability worked with no brick counters")
	}
	for i := 0; i < 3; i++ {
		effectstest.FloatForTest(g, me, "C")
		before := effectstest.LifeOf(g, me.ID)
		if err := g.ActivateCatalogAbility(me.ID, shrine, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.LifeOf(g, me.ID) - before; got != 1 {
			t.Errorf("activation %d gained %d, want 1", i, got)
		}
		effectstest.Untap(g, shrine)
	}
	if n := effectstest.CountersOn(g, shrine, "brick"); n != 3 {
		t.Fatalf("brick counters = %d, want 3", n)
	}
	before := effectstest.LifeOf(g, me.ID)
	if err := g.ActivateCatalogAbility(me.ID, shrine, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("second ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID) - before; got != 2 {
		t.Errorf("gained %d, want 2", got)
	}
}

func TestOldB242PyramidBricksThenAddsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	pyr := effectstest.PushCatalogPermanent(g, me.ID, "Pyramid of the Pantheon", "Artifact", oldB242Pyramid, false)
	if err := g.ActivateManaAbility(me.ID, pyr, 1, game.ManaAbilityParams{Colors: []string{"G"}}); err == nil {
		t.Fatal("the three-mana ability worked with no brick counters")
	}
	for i := 0; i < 3; i++ {
		g.WithWriteLock(func() {
			me.ManaPool.EmptyPool()
			me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
		})
		effectstest.Untap(g, pyr)
		if err := g.ActivateManaAbility(me.ID, pyr, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
		if me.ManaPool.Total() != 1 {
			t.Errorf("activation %d: pool %d, want 1", i, me.ManaPool.Total())
		}
	}
	if n := effectstest.CountersOn(g, pyr, "brick"); n != 3 {
		t.Fatalf("brick counters = %d, want 3", n)
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.Untap(g, pyr)
	if err := g.ActivateManaAbility(me.ID, pyr, 1, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("three-mana ability: %v", err)
	}
	if me.ManaPool.Total() != 3 || me.ManaPool.CountColor("G") != 3 {
		t.Errorf("pool = %v, want GGG", me.ManaPool.Units())
	}
}

func TestOldB242JavelineerCountersAndStopsABlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	jav := effectstest.PushCatalogPermanent(g, me.ID, "Merciless Javelineer", "Creature — Minotaur Warrior", oldB242Javelineer, false)
	foe := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB242Bear, 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, jav, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: foe}}}); err == nil {
		t.Fatal("activated with no card to discard and no mana")
	}
	hand := game.NewCard("Spare Card", me.ID)
	hand.TypeLine = "Sorcery"
	me.Hand.PushTop(hand)
	effectstest.FloatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, jav, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: foe}}, DiscardIDs: []uuid.UUID{hand.InstanceID}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := oldB242Minus(g, foe); n != 1 {
		t.Errorf("the target has %d counters, want 1", n)
	}
	if !effectstest.MidB03CantBlock(t, g, foe) {
		t.Error("the target can still block")
	}
}

func TestOldB242RuthlessSniperPaysOneToShrink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Ruthless Sniper", "Creature — Human Archer", oldB242RuthlessSniper, false)
	foe := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB242Bear, 3, 3)
	card := game.NewCard("Spare Card", me.ID)
	card.TypeLine = "Sorcery"
	me.Hand.PushTop(card)
	effectstest.B30DiscardOne(t, g, me.ID)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickCard(t, g, me.ID, foe)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("the discard did not offer the payment")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := oldB242Minus(g, foe); n != 1 {
		t.Errorf("the target has %d counters, want 1", n)
	}
}

func TestOldB242DecimatorMovesACounterAcross(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	ally := effectstest.B12Creature(g, me.ID, "My Bear", oldB242Bear, 5, 5)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ally, game.CounterMinusOne, 2) })
	foe := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB242Bear, 5, 5)
	beetle := effectstest.PushCatalogPermanent(g, me.ID, "Decimator Beetle", "Creature — Insect", oldB242Decimator, false)
	effectstest.DeclareAttack(t, g, opp.ID, beetle)
	picked := false
	for effectstest.LatestPickTarget(g, me.ID) != nil {
		pick := ally
		if picked {
			pick = foe
		}
		picked = true
		effectstest.PickCard(t, g, me.ID, pick)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := oldB242Minus(g, ally); n != 1 {
		t.Errorf("my creature has %d counters, want 1", n)
	}
	if n := oldB242Minus(g, foe); n != 1 {
		t.Errorf("the defender's creature has %d counters, want 1", n)
	}
}
