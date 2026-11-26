package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b390_cards_test.go — Theros Auras, batch 390 (#5561, S82 OLD).
// Every test plays the real catalog entry through the engine.

const (
	oldB390Viper    = "35dead9a-20ab-45e9-be3e-36857c928ab9"
	oldB390Thassa   = "38e7be61-944e-46f2-be76-fde8b0c19905"
	oldB390Presence = "93fe5eed-520f-467a-8e27-6f40eaa727e8"
	oldB390Heliod   = "a6861402-4819-4d85-8b60-677471f40d7d"
	oldB390Speed    = "aa21315d-4da9-4bbd-9604-c406b83c0db2"
	oldB390Chained  = "afb0c14c-41ac-4536-bfce-f1c36066b278"
	oldB390Fate     = "d578ef54-6a6b-4839-a059-261728fdaef1"
	oldB390Erebos   = "e12497ba-2c49-4609-befa-ebba828bedd6"
	oldB390Aura     = "Enchantment — Aura"
	oldB390Bear     = "Creature — Bear"
)

func TestOldB390CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB390Viper: "Viper's Kiss", oldB390Thassa: "Ordeal of Thassa", oldB390Presence: "Nylea's Presence",
		oldB390Heliod: "Ordeal of Heliod", oldB390Speed: "Messenger's Speed", oldB390Chained: "Chained to the Rocks",
		oldB390Fate: "Fate Foretold", oldB390Erebos: "Ordeal of Erebos",
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

func oldB390Table(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	return g, g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// oldB390Enchant casts the Aura on the host and resolves it.
func oldB390Enchant(t *testing.T, g *game.Game, name, oracle string, host uuid.UUID) uuid.UUID {
	t.Helper()
	aura := effectstest.CastCatalogSpell(t, g, name, oldB390Aura, oracle, effectstest.CardTargets(host))
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardByID(g, aura); c == nil || !c.IsAttachedTo(host) {
		t.Fatalf("%s did not attach to its host", name)
	}
	return aura
}

// oldB390OrdealAtThree enchants a creature that already has two +1/+1
// counters, attacks with it and returns the Aura: the third counter
// sacrifices it.
func oldB390OrdealAtThree(t *testing.T, g *game.Game, me, opp *game.Player, name, oracle string) uuid.UUID {
	t.Helper()
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.FindBattlefieldCardForTest(g, bear).Counters = map[string]int{game.CounterPlusOne: 2}
	aura := oldB390Enchant(t, g, name, oracle, bear)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 3 {
		t.Fatalf("%d +1/+1 counters, want 3", got)
	}
	if effectstest.FindBattlefieldCardByID(g, aura) != nil {
		t.Fatal("the Aura was not sacrificed at three counters")
	}
	return aura
}

func oldB390HasSubtype(g *game.Game, id uuid.UUID, subtype string) bool {
	found := false
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				eff := c.Effective()
				for _, st := range eff.Subtypes {
					if st == subtype {
						found = true
					}
				}
			}
		}
	})
	return found
}

func TestOldB390OrdealOfThassaDrawsTwoWhenSacrificed(t *testing.T) {
	g, me, opp := oldB390Table(t)
	before := me.Hand.Size()
	oldB390OrdealAtThree(t, g, me, opp, "Ordeal of Thassa", oldB390Thassa)
	// The Aura left the hand when it was cast: a net of two cards drawn.
	if got := me.Hand.Size(); got != before+2 {
		t.Errorf("hand %d, want %d (two cards drawn)", got, before+2)
	}
}

func TestOldB390OrdealOfThassaDoesNothingBelowThreeCounters(t *testing.T) {
	g, me, opp := oldB390Table(t)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	aura := oldB390Enchant(t, g, "Ordeal of Thassa", oldB390Thassa, bear)
	before := me.Hand.Size()
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters after one attack, want 1", got)
	}
	if effectstest.FindBattlefieldCardByID(g, aura) == nil {
		t.Error("the Aura was sacrificed with one counter")
	}
	if me.Hand.Size() != before {
		t.Errorf("hand changed from %d to %d", before, me.Hand.Size())
	}
}

func TestOldB390OrdealOfHeliodGainsTenLifeWhenSacrificed(t *testing.T) {
	g, me, opp := oldB390Table(t)
	life := me.Life
	oldB390OrdealAtThree(t, g, me, opp, "Ordeal of Heliod", oldB390Heliod)
	if me.Life != life+10 {
		t.Errorf("life %d, want %d", me.Life, life+10)
	}
}

func TestOldB390OrdealOfErebosMakesTheTargetPlayerDiscardTwo(t *testing.T) {
	g, me, opp := oldB390Table(t)
	for opp.Hand.Size() > 0 {
		_, _ = opp.Hand.PopTop()
	}
	var kept uuid.UUID
	for i := 0; i < 3; i++ {
		kept = effectstest.EoHand(opp, "Card", "Sorcery")
	}
	_ = kept
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.FindBattlefieldCardForTest(g, bear).Counters = map[string]int{game.CounterPlusOne: 2}
	oldB390Enchant(t, g, "Ordeal of Erebos", oldB390Erebos, bear)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickPlayerTarget(t, g, me.ID, opp.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardOwed(g, opp.ID) != 2 {
		t.Fatalf("the opponent owes %d discards, want 2", effectstest.DiscardOwed(g, opp.ID))
	}
	effectstest.DiscardFromHand(t, g, opp.ID)
	if c := effectstest.DiscardChoiceFor(g, opp.ID); c != nil {
		effectstest.DiscardFromHand(t, g, opp.ID)
	}
	if opp.Hand.Size() != 1 {
		t.Errorf("the opponent's hand is %d, want 1", opp.Hand.Size())
	}
}

func TestOldB390VipersKissShrinksAndStopsActivatedAbilities(t *testing.T) {
	g, me, _ := oldB390Table(t)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 3, 3)
	oldB390Enchant(t, g, "Viper's Kiss", oldB390Viper, bear)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("%d/%d, want 2/2", p, tough)
	}
	r := effectstest.AuraRestrictions(t, g, bear)
	if r&game.CantActivate == 0 || r&game.CantActivateMana == 0 {
		t.Errorf("restrictions %v lack the activation bits", r)
	}
}

func TestOldB390MessengersSpeedGrantsTrampleAndHaste(t *testing.T) {
	g, me, _ := oldB390Table(t)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	other := effectstest.PushVanillaCreature(g, me.ID, "Other Bear", 2, 2)
	oldB390Enchant(t, g, "Messenger's Speed", oldB390Speed, bear)
	for _, kw := range []string{"trample", "haste"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, bear, kw) {
			t.Errorf("the enchanted creature lacks %s", kw)
		}
		if effectstest.HasKeywordOnBattlefield(t, g, other, kw) {
			t.Errorf("an unenchanted creature has %s", kw)
		}
	}
}

func TestOldB390FateForetoldDrawsOnEntryAndWhenTheHostDies(t *testing.T) {
	g, me, opp := oldB390Table(t)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	mine := me.Hand.Size()
	theirHand := opp.Hand.Size()
	oldB390Enchant(t, g, "Fate Foretold", oldB390Fate, theirs)
	if got := me.Hand.Size(); got != mine+1 {
		t.Errorf("hand %d, want %d (one card drawn)", got, mine+1)
	}
	effectstest.B27Kill(g, theirs)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Hand.Size(); got != theirHand+1 {
		t.Errorf("the creature's controller's hand %d, want %d", got, theirHand+1)
	}
	if got := me.Hand.Size(); got != mine+1 {
		t.Errorf("the Aura's controller drew for the opponent's creature: hand %d, want %d", got, mine+1)
	}
}

func TestOldB390NyleasPresenceDrawsAndMakesTheLandEveryBasicType(t *testing.T) {
	g, me, _ := oldB390Table(t)
	land := effectstest.PushLandFor(g, me.ID, "Wasteland Test", "Land")
	before := me.Hand.Size()
	oldB390Enchant(t, g, "Nylea's Presence", oldB390Presence, land)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d, want %d (one card drawn)", got, before+1)
	}
	for _, want := range []string{"Plains", "Island", "Swamp", "Mountain", "Forest"} {
		if !oldB390HasSubtype(g, land, want) {
			t.Errorf("the enchanted land is not a %s", want)
		}
	}
}

func TestOldB390ChainedToTheRocksExilesUntilItLeaves(t *testing.T) {
	g, me, opp := oldB390Table(t)
	mountain := effectstest.PushLandFor(g, me.ID, "Mountain", "Basic Land — Mountain")
	plains := effectstest.PushLandFor(g, me.ID, "Plains", "Basic Land — Plains")
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB390Bear, 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Chained to the Rocks", oldB390Aura, oldB390Chained, effectstest.CardTargets(plains)); err == nil {
		t.Fatal("the Aura was cast on a Plains")
	}
	aura := effectstest.CastCatalogSpell(t, g, "Chained to the Rocks", oldB390Aura, oldB390Chained, effectstest.CardTargets(mountain))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetIfOpen(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Fatal("the creature is still on the battlefield")
	}
	if !g.Exile.Contains(theirs) {
		t.Fatal("the creature is not in exile")
	}
	effectstest.B27Kill(g, aura)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.NamedOnBattlefield(g, "Their Bear", opp.ID); n != 1 {
		t.Errorf("after the Aura left, %d of the exiled creature under its owner's control, want 1", n)
	}
}
