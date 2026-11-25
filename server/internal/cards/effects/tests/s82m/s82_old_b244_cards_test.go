package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b244_cards_test.go — Amonkhet Auras, batch 244 (S82 older sets,
// #5535). Every test plays the real catalog entry through the engine.

const (
	oldB244CompulsoryRest  = "020225db-5623-49c2-8cdf-1ec2a05b8b0a"
	oldB244ConsumingFervor = "16dddf03-449c-4df0-9902-f8901f05656e"
	oldB244SixthSense      = "25d35e3e-ece8-4c94-9dbc-d076a52fca0d"
	oldB244TrespassersCurs = "2769751d-ca67-4d6e-93c9-9587367ae8ab"
	oldB244CruelReality    = "2e3cc9b6-7849-4385-a487-c8f5364174e7"
	oldB244CartoucheAmbit  = "33fe5484-7cc7-45c7-a349-b15c0454ab08"
	oldB244IllusoryWrap    = "bda363b6-4be1-4415-8b3e-92882df1f630"
	oldB244CartoucheStr    = "cba6563e-96e5-4cd3-bddf-0cf1368f9e15"
)

func TestOldB244EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		oldB244CompulsoryRest, oldB244ConsumingFervor, oldB244SixthSense, oldB244TrespassersCurs,
		oldB244CruelReality, oldB244CartoucheAmbit, oldB244IllusoryWrap, oldB244CartoucheStr,
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

// oldB244Aura casts an Aura onto a creature and resolves it (and its enters trigger's target pick,
// when `pick` is set).
func oldB244Aura(t *testing.T, g *game.Game, name, oracle string, host uuid.UUID) uuid.UUID {
	t.Helper()
	id := effectstest.CastCatalogSpell(t, g, name, "Enchantment — Aura", oracle, effectstest.CardTarget(host))
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestOldB244CompulsoryRestStopsAttackAndBlockAndGrantsASacrificeOutlet(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	oldB244Aura(t, g, "Compulsory Rest", oldB244CompulsoryRest, bear)
	var r game.Restriction
	g.ReadSnapshot(func() { c, _ := g.LookupCardForEffect(bear); r = c.Effective().Restrictions })
	if r&game.CantAttack == 0 || r&game.CantBlock == 0 {
		t.Fatalf("restrictions = %v, want can't attack and can't block", r)
	}
	if r&(game.CantActivate|game.CantActivateMana) != 0 {
		t.Error("its own abilities were stopped; only attacking and blocking are")
	}
	life := effectstest.LifeOf(g, me.ID)
	effectstest.FloatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("the granted ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature was not sacrificed")
	}
	if got := effectstest.LifeOf(g, me.ID) - life; got != 2 {
		t.Errorf("gained %d life, want 2", got)
	}
}

func TestOldB244ConsumingFervorPumpsAndShrinksEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	oldB244Aura(t, g, "Consuming Fervor", oldB244ConsumingFervor, bear)
	if p, tough := effectstest.PtOf(t, g, bear); p != 5 || tough != 5 {
		t.Fatalf("Bear is %d/%d, want 5/5", p, tough)
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, game.CounterMinusOne); n != 1 {
		t.Errorf("the creature has %d -1/-1 counters after my upkeep, want 1", n)
	}
}

func TestOldB244SixthSenseDrawsOnlyWhenYouSayYes(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := effectstest.SeatsForCombat(g)
		effectstest.SeedLibrary(me, "A", "B")
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		oldB244Aura(t, g, "Sixth Sense", oldB244SixthSense, bear)
		hand := me.Hand.Size()
		effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 2)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		want := hand
		if yes {
			want++
		}
		if me.Hand.Size() != want {
			t.Errorf("answer %v: hand %d, want %d", yes, me.Hand.Size(), want)
		}
	}
}

func TestOldB244IllusoryWrappingsSetsBaseToZeroTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.SeatsForCombat(g)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 5, 5)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ogre, game.CounterPlusOne, 1) })
	oldB244Aura(t, g, "Illusory Wrappings", oldB244IllusoryWrap, ogre)
	if p, tough := effectstest.PtOf(t, g, ogre); p != 1 || tough != 3 {
		t.Errorf("Ogre is %d/%d, want 1/3 (base 0/2 plus its +1/+1 counter)", p, tough)
	}
}

func TestOldB244CartoucheOfStrengthPumpsAndMayFight(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := effectstest.SeatsForCombat(g)
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		foe := effectstest.B12Creature(g, opp.ID, "Goblin", "Creature — Goblin", 1, 3)
		effectstest.CastCatalogSpell(t, g, "Cartouche of Strength", "Enchantment — Aura Cartouche", oldB244CartoucheStr, effectstest.CardTarget(bear))
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.PickCard(t, g, me.ID, foe)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 3 {
			t.Errorf("answer %v: Bear is %d/%d, want 3/3", yes, p, tough)
		}
		if !effectstest.HasKeywordOf(g, bear, "trample") {
			t.Errorf("answer %v: the Bear has no trample", yes)
		}
		if died := !g.Battlefield.Contains(foe); died != yes {
			t.Errorf("answer %v: the 1/3 died = %v", yes, died)
		}
	}
}

func TestOldB244CartoucheOfAmbitionPumpsAndMayShrinkACreature(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := effectstest.SeatsForCombat(g)
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		foe := effectstest.B12Creature(g, opp.ID, "Goblin", "Creature — Goblin", 3, 3)
		effectstest.CastCatalogSpell(t, g, "Cartouche of Ambition", "Enchantment — Aura Cartouche", oldB244CartoucheAmbit, effectstest.CardTarget(bear))
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.PickCard(t, g, me.ID, foe)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 3 {
			t.Errorf("answer %v: Bear is %d/%d, want 3/3", yes, p, tough)
		}
		if !effectstest.HasKeywordOf(g, bear, "lifelink") {
			t.Errorf("answer %v: the Bear has no lifelink", yes)
		}
		want := 0
		if yes {
			want = 1
		}
		if n := effectstest.CountersOn(g, foe, game.CounterMinusOne); n != want {
			t.Errorf("answer %v: the target has %d -1/-1 counters, want %d", yes, n, want)
		}
	}
}

func TestOldB244TrespassersCurseDrainsForCreaturesEnteringUnderTheEnchantedPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.CastCatalogSpell(t, g, "Trespasser's Curse", "Enchantment — Aura Curse", oldB244TrespassersCurs, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	card := game.NewCard("Their Bear", opp.ID)
	card.TypeLine, card.Power, card.Toughness = "Creature — Bear", 2, 2
	opp.Graveyard.PushTop(card)
	mine := game.NewCard("My Bear", me.ID)
	mine.TypeLine, mine.Power, mine.Toughness = "Creature — Bear", 2, 2
	me.Graveyard.PushTop(mine)
	myLife, oppLife := effectstest.LifeOf(g, me.ID), effectstest.LifeOf(g, opp.ID)
	effectstest.B30Reanimate(t, g, mine.InstanceID, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LifeOf(g, me.ID) != myLife || effectstest.LifeOf(g, opp.ID) != oppLife {
		t.Fatal("my own creature entering triggered the Curse")
	}
	effectstest.B30Reanimate(t, g, card.InstanceID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID) - oppLife; got != -1 {
		t.Errorf("the enchanted player's life changed by %d, want -1", got)
	}
	if got := effectstest.LifeOf(g, me.ID) - myLife; got != 1 {
		t.Errorf("my life changed by %d, want +1", got)
	}
}

func TestOldB244CruelRealitySacrificesOrLosesFive(t *testing.T) {
	for _, hasCreature := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		_, opp := effectstest.SeatsForCombat(g)
		var bear uuid.UUID
		if hasCreature {
			bear = effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
		}
		effectstest.CastCatalogSpell(t, g, "Cruel Reality", "Enchantment — Aura Curse", oldB244CruelReality, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
		effectstest.PassPriorityAroundTable(t, g)
		life := effectstest.LifeOf(g, opp.ID)
		effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
		effectstest.PassPriorityAroundTable(t, g)
		if hasCreature {
			if p := effectstest.ChooseCardsChoiceFor(g, opp.ID); p != nil {
				effectstest.AnswerChooseCards(t, g, opp.ID, bear)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if g.Battlefield.Contains(bear) {
				t.Error("the creature was not sacrificed")
			}
			if effectstest.LifeOf(g, opp.ID) != life {
				t.Errorf("life changed to %d though a creature was sacrificed", effectstest.LifeOf(g, opp.ID))
			}
		} else if got := effectstest.LifeOf(g, opp.ID) - life; got != -5 {
			t.Errorf("with nothing to sacrifice life changed by %d, want -5", got)
		}
	}
}
