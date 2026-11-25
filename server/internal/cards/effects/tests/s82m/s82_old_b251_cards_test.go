package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b251_cards_test.go — Aether Revolt cards, S82 OLD batch 251
// (#5536). Every test plays the real catalog entry through the engine.

const (
	b251AjanisComrade     = "15b86440-2040-47ae-ab4e-8445b3e24462"
	b251OathOfAjani       = "1ec39c8f-eb1e-4881-a749-0c188788b612"
	b251DarkIntimations   = "228cb63f-0b93-4be5-94e5-4eaae9112eeb"
	b251IroncladRevolt    = "382109c5-fd08-4e82-b898-2a387181b470"
	b251MaulfistRevolt    = "65d41af0-978e-4071-9f64-0e3b5826763a"
	b251LifecraftAwaken   = "8919c8dd-d3b3-406f-a387-44cc1b6d92f3"
	b251FoundryHornet     = "e5342529-dc1a-4ebb-a0fd-843e01147402"
	b251CreatureBear      = "Creature — Bear"
	b251PlaneswalkerKind  = "Legendary Planeswalker — "
	b251ArtifactTypeLine  = "Artifact"
	b251ConstructTypeLine = "Artifact Creature — Construct"
)

func b251Walker(g *game.Game, owner uuid.UUID, kind string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: kind + " Test", TypeLine: b251PlaneswalkerKind + kind,
		Counters: map[string]int{game.CounterLoyalty: 4}, Owner: owner, Controller: owner,
	})
}

func TestB251AjanisComradeGrowsOnlyWithAnAjani(t *testing.T) {
	for _, ajani := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		comrade := effectstest.B12Push(g, me.ID, "Ajani's Comrade", "Creature — Elf Soldier", b251AjanisComrade, 2, 2)
		effectstest.AssertKeywords(t, g, comrade, "trample")
		kind := "Jace"
		if ajani {
			kind = "Ajani"
		}
		b251Walker(g, me.ID, kind)
		effectstest.AdvanceTo(t, g, game.StepBeginCombat)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if ajani {
			want = 1
		}
		if got := effectstest.PlusOnes(g, comrade); got != want {
			t.Errorf("an Ajani %v: %d counters, want %d", ajani, got, want)
		}
	}
}

func TestB251OathOfAjaniCountersAndDiscount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", b251CreatureBear, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", b251CreatureBear, 2, 2)
	effectstest.B12PlayFromHand(t, g, "Oath of Ajani", "Legendary Enchantment", b251OathOfAjani, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, mine); got != 1 {
		t.Errorf("my creature has %d counters, want 1", got)
	}
	if got := effectstest.PlusOnes(g, theirs); got != 0 {
		t.Errorf("their creature has %d counters, want 0", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Walker", b251PlaneswalkerKind+"Test", "{2}{U}{U}"); got != 3 {
		t.Errorf("my planeswalker spell: %d, want 3", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Bear", b251CreatureBear, "{2}{G}"); got != 3 {
		t.Errorf("my creature spell: %d, want 3 (untouched)", got)
	}
	if got := effectstest.PriceInHand(t, g, opp, "Walker", b251PlaneswalkerKind+"Test", "{2}{U}{U}"); got != 4 {
		t.Errorf("their planeswalker spell: %d, want 4 (undiscounted)", got)
	}
}

func TestB251IroncladRevolutionarySacrificesAnArtifactForCountersAndDrain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := effectstest.B12Permanent(g, me.ID, "Mind Stone", b251ArtifactTypeLine)
	theirLife := opp.Life
	revolutionary := effectstest.CastHoldCreature(t, g, "Ironclad Revolutionary", "Creature — Aetherborn Artificer", b251IroncladRevolt, 4, 4)
	effectstest.AnswerOwnPermanents(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
	if got := effectstest.PlusOnes(g, revolutionary); got != 2 {
		t.Errorf("Ironclad Revolutionary has %d counters, want 2", got)
	}
	if opp.Life != theirLife-2 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, theirLife-2)
	}
}

func TestB251IroncladRevolutionaryDoesNothingWithoutTheSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Permanent(g, me.ID, "Mind Stone", b251ArtifactTypeLine)
	theirLife := opp.Life
	revolutionary := effectstest.CastHoldCreature(t, g, "Ironclad Revolutionary", "Creature — Aetherborn Artificer", b251IroncladRevolt, 4, 4)
	effectstest.AnswerOwnPermanents(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, revolutionary); got != 0 || opp.Life != theirLife {
		t.Errorf("declining still gave %d counters and left the opponent at %d", got, opp.Life)
	}
}

func TestB251MaulfistRevolutionaryAddsAnotherOfEachKind(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: b251CreatureBear, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterPlusOne: 1, game.CounterCharge: 2},
	})
	effectstest.CastHoldCreature(t, g, "Maulfist Revolutionary", "Creature — Human Warrior", b251MaulfistRevolt, 3, 3)
	effectstest.AnswerPickTargetIfOpen(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters: %d, want 2", got)
	}
	if got := effectstest.CounterCount(g, bear, game.CounterCharge); got != 3 {
		t.Errorf("charge counters: %d, want 3", got)
	}
}

func TestB251LifecraftAwakeningAnimatesANoncreatureArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := effectstest.B12Permanent(g, me.ID, "Mind Stone", b251ArtifactTypeLine)
	if err := effectstest.CastX(t, g, "Lifecraft Awakening", "Instant", b251LifecraftAwaken, 2, effectstest.CardRefs(rock)); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, rock); got != 2 {
		t.Fatalf("the artifact has %d counters, want 2", got)
	}
	if p, tough := effectstest.PtOf(t, g, rock); p != 2 || tough != 2 {
		t.Errorf("the animated artifact is %d/%d, want 2/2", p, tough)
	}
	c := effectstest.B12Card(t, g, rock)
	if !c.IsCreature() || !c.HasSubtype("Construct") {
		t.Errorf("the artifact is %q, want a Construct creature", c.TypeLine)
	}
}

func TestB251LifecraftAwakeningLeavesACreatureAsItIs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bot := effectstest.B12Creature(g, me.ID, "Gearbot", b251ConstructTypeLine, 1, 1)
	if err := effectstest.CastX(t, g, "Lifecraft Awakening", "Instant", b251LifecraftAwaken, 1, effectstest.CardRefs(bot)); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bot); p != 2 || tough != 2 {
		t.Errorf("the creature is %d/%d, want 2/2 (its own size plus one counter)", p, tough)
	}
}

func TestB251FoundryHornetShrinksOpponentsOnlyWithACounteredCreature(t *testing.T) {
	for _, countered := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		mine := effectstest.B12Creature(g, me.ID, "Mine", b251CreatureBear, 2, 2)
		theirs := effectstest.B12Creature(g, opp.ID, "Theirs", b251CreatureBear, 2, 2)
		if countered {
			effectstest.PutCounters(t, g, mine, game.CounterPlusOne, 1)
		}
		effectstest.CastHoldCreature(t, g, "Foundry Hornet", "Creature — Insect", b251FoundryHornet, 2, 3)
		effectstest.PassPriorityAroundTable(t, g)
		wantTheirs := 2
		if countered {
			wantTheirs = 1
		}
		if p, _ := effectstest.PtOf(t, g, theirs); p != wantTheirs {
			t.Errorf("countered=%v: their creature has power %d, want %d", countered, p, wantTheirs)
		}
		wantMine := 2
		if countered {
			wantMine = 3
		}
		if p, _ := effectstest.PtOf(t, g, mine); p != wantMine {
			t.Errorf("countered=%v: my creature has power %d, want %d (untouched)", countered, p, wantMine)
		}
	}
}

func TestB251DarkIntimationsDoesEachThing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.SeedLibrary(me, "a", "b", "c")
	victim := effectstest.B12Creature(g, opp.ID, "Victim", b251CreatureBear, 1, 1)
	fallen := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: fallen, Name: "Fallen", TypeLine: b251CreatureBear, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	oppHand, myHand := opp.Hand.Size(), me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Dark Intimations", "Sorcery", b251DarkIntimations, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, opp.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		if s.ID != me.ID && effectstest.DiscardChoiceFor(g, s.ID) != nil {
			effectstest.DiscardFromHand(t, g, s.ID)
			effectstest.PassPriorityAroundTable(t, g)
		}
	}
	if c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID); c != nil {
		if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{fallen}); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if g.Battlefield.Contains(victim) {
		t.Error("the opponent kept the creature")
	}
	if got := oppHand - opp.Hand.Size(); got != 1 {
		t.Errorf("the opponent discarded %d cards, want 1", got)
	}
	if !me.Hand.Contains(fallen) {
		t.Error("the creature card did not come back to my hand")
	}
	// +1 returned, +1 drawn (the spell itself went in and out of the hand).
	if got := me.Hand.Size() - myHand; got != 2 {
		t.Errorf("my hand grew by %d, want 2", got)
	}
}

func TestB251DarkIntimationsExilesItselfForABolasAndAddsLoyalty(t *testing.T) {
	for _, bolas := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		card := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: card, Name: "Dark Intimations", TypeLine: "Sorcery", OracleID: b251DarkIntimations, Owner: me.ID, Controller: me.ID})
		kind := "Jace"
		if bolas {
			kind = "Bolas"
		}
		walker := effectstest.B12PlayFromHand(t, g, kind+" Walker", b251PlaneswalkerKind+kind, "", game.CastSpellParams{})
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		inYard := effectstest.GraveyardHas(me, card)
		if bolas == inYard {
			t.Errorf("bolas=%v: Dark Intimations in the graveyard = %v", bolas, inYard)
		}
		got := effectstest.CountersOn(g, walker, game.CounterLoyalty)
		if bolas && got != 1 {
			t.Errorf("the Bolas planeswalker entered with %d loyalty, want 1 (its 0 printed plus the extra)", got)
		}
	}
}
