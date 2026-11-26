package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 set AVR, batch 434 (#5570): Avacyn Restored, life gain and loss.

const (
	avr434HomicidalSeclusion = "3bc51b45-1f96-4277-8b80-bb0a528c6ac2"
	avr434ShelteringWord     = "77aaaad5-7960-453d-b024-6dacbca3d606"
	avr434EssenceHarvest     = "ae9a398e-f3a7-4878-a14d-0f6c2dac3d83"
	avr434Scroll             = "c961373a-be0e-467c-92e0-33a0c953736a"
	avr434MentalAgony        = "e80a7118-dd6b-4907-ad69-09387e0c2114"
)

func avr434Creature(g *game.Game, owner uuid.UUID, typeLine string, p, tough int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test " + typeLine, TypeLine: typeLine,
		Power: p, Toughness: tough, Owner: owner, Controller: owner,
	})
}

func TestAvrB434RegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		avr434HomicidalSeclusion: "Homicidal Seclusion", avr434ShelteringWord: "Sheltering Word",
		avr434EssenceHarvest: "Essence Harvest", avr434Scroll: "Scroll of Griselbrand",
		avr434MentalAgony: "Mental Agony",
	} {
		effectstest.StdRegisteredFull(t, oracle, name)
	}
}

func TestAvrB434HomicidalSeclusionOnlyWithExactlyOneCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Homicidal Seclusion", "Enchantment", avr434HomicidalSeclusion, false)
	solo := avr434Creature(g, me.ID, effectstest.TestCreatureTypeLine, 2, 2)
	if p, tough := effectstest.EffectivePTOf(t, g, solo); p != 5 || tough != 3 {
		t.Fatalf("the lone creature is %d/%d, want 5/3", p, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, solo, "lifelink") {
		t.Error("the lone creature should have lifelink")
	}
	second := avr434Creature(g, me.ID, effectstest.TestCreatureTypeLine, 2, 2)
	if p, tough := effectstest.EffectivePTOf(t, g, solo); p != 2 || tough != 2 {
		t.Errorf("with two creatures the first is %d/%d, want 2/2", p, tough)
	}
	if effectstest.HasKeywordOnBattlefield(t, g, second, "lifelink") {
		t.Error("with two creatures nobody has lifelink")
	}
	opp := g.Seats[1]
	other := avr434Creature(g, opp.ID, effectstest.TestCreatureTypeLine, 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(second) })
	if p, _ := effectstest.EffectivePTOf(t, g, solo); p != 5 {
		t.Errorf("back to one creature: power %d, want 5", p)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, other); p != 2 {
		t.Errorf("an opponent's creature is %d power, want 2: it is not affected", p)
	}
}

func TestAvrB434ShelteringWordGivesHexproofAndLifeEqualToToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	wall := avr434Creature(g, me.ID, "Creature — Wall", 0, 5)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Sheltering Word", "Instant", avr434ShelteringWord,
		[]game.TargetRef{{Kind: game.TargetCard, ID: wall}})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, wall, "hexproof") {
		t.Error("the creature should have hexproof")
	}
	if me.Life != life+5 {
		t.Errorf("life = %d, want %d (its toughness)", me.Life, life+5)
	}
	// "Target creature you control": an opponent's creature is refused.
	theirs := avr434Creature(g, g.Seats[1].ID, "Creature — Wall", 0, 5)
	if err := effectstest.CastCatalogSpellErr(t, g, "Sheltering Word", "Instant", avr434ShelteringWord,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}}); err == nil {
		t.Error("an opponent's creature is not a legal target")
	}
}

func TestAvrB434EssenceHarvestDrainsByGreatestPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	avr434Creature(g, me.ID, effectstest.TestCreatureTypeLine, 2, 2)
	avr434Creature(g, me.ID, effectstest.TestCreatureTypeLine, 4, 1)
	avr434Creature(g, opp.ID, effectstest.TestCreatureTypeLine, 9, 9)
	mine, theirs := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "Essence Harvest", "Sorcery", avr434EssenceHarvest, effectstest.AcPlayerRef(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirs-4 || me.Life != mine+4 {
		t.Errorf("life: opponent %d (want %d), me %d (want %d)", opp.Life, theirs-4, me.Life, mine+4)
	}
}

func TestAvrB434MentalAgonyDiscardsTwoAndLosesTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.PushHandCard(g, opp)
	effectstest.PushHandCard(g, opp)
	effectstest.PushHandCard(g, opp)
	life, hand := opp.Life, opp.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Mental Agony", "Sorcery", avr434MentalAgony, effectstest.AcPlayerRef(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.DiscardChoiceFor(g, opp.ID)
	if c == nil || c.ChooseMax != 2 {
		t.Fatalf("want a two-card discard prompt for the target, got %+v", c)
	}
	effectstest.DiscardFromHand(t, g, opp.ID)
	if opp.Hand.Size() != hand-2 {
		t.Errorf("hand = %d, want %d", opp.Hand.Size(), hand-2)
	}
	if opp.Life != life-2 {
		t.Errorf("life = %d, want %d", opp.Life, life-2)
	}
}

func avr434UseScroll(t *testing.T, g *game.Game, demon bool) (me, opp *game.Player) {
	t.Helper()
	me, opp = g.Seats[0], g.Seats[1]
	scroll := effectstest.PushCatalogPermanent(g, me.ID, "Scroll of Griselbrand", "Artifact", avr434Scroll, false)
	if demon {
		avr434Creature(g, me.ID, "Creature — Demon", 7, 4)
	}
	effectstest.PushHandCard(g, opp)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, scroll, 0, game.ActivateAbilityParams{Strict: true,
		Targets: effectstest.AcPlayerRef(opp.ID)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(scroll) {
		t.Error("the Scroll is sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	return me, opp
}

func TestAvrB434ScrollOfGriselbrandDiscardsWithoutADemon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := g.Seats[0], g.Seats[1]
	life, hand := opp.Life, opp.Hand.Size()
	avr434UseScroll(t, g, false)
	effectstest.DiscardFromHand(t, g, opp.ID)
	if opp.Hand.Size() != hand {
		t.Errorf("hand = %d, want %d (one seeded, one discarded)", opp.Hand.Size(), hand)
	}
	if opp.Life != life {
		t.Errorf("life = %d, want %d without a Demon", opp.Life, life)
	}
}

func TestAvrB434ScrollOfGriselbrandDrainsWithADemon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	life := opp.Life
	avr434UseScroll(t, g, true)
	effectstest.DiscardFromHand(t, g, opp.ID)
	if opp.Life != life-3 {
		t.Errorf("life = %d, want %d with a Demon", opp.Life, life-3)
	}
}
