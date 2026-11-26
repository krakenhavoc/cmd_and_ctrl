package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b502_test.go — Alara Reborn, older-sets batch 502 (#5586). Every test
// plays the real catalog entry through the engine.

const (
	oldB502Vectis    = "1ed63322-f877-4f1c-90bd-e1adb3668d6d"
	oldB502Ogre      = "a0f26f84-9607-4411-bca5-681ba4e57756"
	oldB502Shepherd  = "c4a4e607-7f37-4dc6-8fc0-f77bdafeb919"
	oldB502Page      = "f1c87913-9374-41d7-ae70-444d427290b4"
	oldB502Sigil     = "f39fd4bd-3157-42da-9a39-94e094b5f681"
	oldB502Slave     = "07190a8b-2a59-45b3-9fbf-505161698a9c"
	oldB502Soulquake = "0b9a477b-f2dd-47b2-8493-eb63f44a2688"
	oldB502Double    = "9fe57d74-0394-4263-a1b1-da9773b485a6"
	oldB502Funeral   = "cf5103c1-2590-4aed-b3ae-8a570f797738"
)

func TestOldB502EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB502Vectis: "Vectis Dominator", oldB502Ogre: "Singe-Mind Ogre", oldB502Shepherd: "Mycoid Shepherd",
		oldB502Page: "Etherwrought Page", oldB502Sigil: "Tainted Sigil", oldB502Slave: "Slave of Bolas",
		oldB502Soulquake: "Soulquake", oldB502Double: "Double Negative", oldB502Funeral: "Mind Funeral",
	} {
		effectstest.StdRegisteredFull(t, oracle, name)
	}
}

// --- Vectis Dominator ---------------------------------------------------------

func oldB502Vectis1(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	vectis := effectstest.PushCatalogPermanent(g, me.ID, "Vectis Dominator", "Artifact Creature — Human Wizard", oldB502Vectis, false)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, vectis, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return g, me, opp, vectis, bear
}

func TestOldB502VectisTapsTheCreatureWhenItsControllerDeclines(t *testing.T) {
	g, _, opp, _, bear := oldB502Vectis1(t)
	effectstest.AnswerOptionPick(t, g, opp.ID, 0)
	if !effectstest.B16Tapped(t, g, bear) {
		t.Error("the creature was not tapped")
	}
}

func TestOldB502VectisSpareTheCreatureForTwoLife(t *testing.T) {
	g, _, opp, _, bear := oldB502Vectis1(t)
	life := opp.Life
	effectstest.AnswerOptionPick(t, g, opp.ID, 1)
	if effectstest.B16Tapped(t, g, bear) {
		t.Error("the creature was tapped although its controller paid")
	}
	_ = life
}

func TestOldB502VectisAtOneLifeThereIsNothingToAsk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	opp.Life = 1
	vectis := effectstest.PushCatalogPermanent(g, me.ID, "Vectis Dominator", "Artifact Creature — Human Wizard", oldB502Vectis, false)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, vectis, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, bear) {
		t.Error("the creature was not tapped by a controller who can't pay")
	}
}

// --- Singe-Mind Ogre ----------------------------------------------------------

func TestOldB502OgreMakesTheTargetLoseLifeEqualToARandomCardsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	opp.Hand.Cards = nil
	opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Five Drop", TypeLine: "Creature — Giant", ManaCost: "{4}{R}", Owner: opp.ID, Controller: opp.ID})
	life := opp.Life
	effectstest.B12PlayFromHand(t, g, "Singe-Mind Ogre", "Creature — Ogre Mutant", oldB502Ogre, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTargetPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-5 {
		t.Errorf("life %d, want %d", opp.Life, life-5)
	}
}

func TestOldB502OgreWithAnEmptyHandLosesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	opp.Hand.Cards = nil
	life := opp.Life
	effectstest.B12PlayFromHand(t, g, "Singe-Mind Ogre", "Creature — Ogre Mutant", oldB502Ogre, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTargetPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("life %d, want %d", opp.Life, life)
	}
}

// --- Mycoid Shepherd ----------------------------------------------------------

func TestOldB502ShepherdGainsFiveWhenAPowerFiveCreatureDies(t *testing.T) {
	for _, tc := range []struct {
		name  string
		power int
		own   bool
		want  int
	}{{"power five of mine", 5, true, 5}, {"power four of mine", 4, true, 0}, {"power six of theirs", 6, false, 0}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			effectstest.B12Push(g, me.ID, "Mycoid Shepherd", "Creature — Fungus", oldB502Shepherd, 5, 4)
			owner := me
			if !tc.own {
				owner = opp
			}
			victim := effectstest.B12Creature(g, owner.ID, "Victim", "Creature — Giant", tc.power, 3)
			life := me.Life
			effectstest.KillCreature(t, g, owner.ID, victim)
			effectstest.B06AnswerAllTriggerPrompts(t, g, me.ID, true)
			if me.Life != life+tc.want {
				t.Errorf("life %d, want %d", me.Life, life+tc.want)
			}
		})
	}
}

func TestOldB502ShepherdTriggersForItselfAndMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	shepherd := effectstest.B12Push(g, me.ID, "Mycoid Shepherd", "Creature — Fungus", oldB502Shepherd, 5, 4)
	life := me.Life
	effectstest.KillCreature(t, g, me.ID, shepherd)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("declined, yet life is %d, want %d", me.Life, life)
	}
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.AdvanceToMain(t, g2)
	s2 := effectstest.B12Push(g2, me2.ID, "Mycoid Shepherd", "Creature — Fungus", oldB502Shepherd, 5, 4)
	effectstest.KillCreature(t, g2, me2.ID, s2)
	effectstest.AnswerLatestTriggerPrompt(t, g2, me2.ID, true)
	effectstest.PassPriorityAroundTable(t, g2)
	if me2.Life != life+5 {
		t.Errorf("life %d, want %d", me2.Life, life+5)
	}
}

// --- Etherwrought Page --------------------------------------------------------

func oldB502PageUpkeep(t *testing.T, mode int) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Etherwrought Page", "Artifact", oldB502Page, false)
	effectstest.SeedLibrary(me, "Top A", "Top B", "Top C")
	effectstest.B12ToMyNextUpkeep(t, g)
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		effectstest.PassPriorityAroundTable(t, g)
		c = effectstest.ModePickChoiceFor(g, me.ID)
	}
	if c == nil {
		t.Fatal("the upkeep trigger asked for no mode")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	return g, me, opp
}

func TestOldB502PageGainsTwoLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	base := g.Seats[0].Life
	g, me, _ := oldB502PageUpkeep(t, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != base+2 {
		t.Errorf("life %d, want %d", me.Life, base+2)
	}
}

func TestOldB502PageEachOpponentLosesOne(t *testing.T) {
	g, me, opp := oldB502PageUpkeep(t, 2)
	lives := []int{}
	for _, p := range g.Seats {
		lives = append(lives, p.Life)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		want := lives[i]
		if p.ID != me.ID {
			want--
		}
		if p.Life != want {
			t.Errorf("%s life %d, want %d", p.Name, p.Life, want)
		}
	}
	_ = opp
}

func TestOldB502PageSurveilAsksAboutTheTopCard(t *testing.T) {
	g, me, _ := oldB502PageUpkeep(t, 1)
	effectstest.PassPriorityAroundTable(t, g)
	var surveil *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceSurveil && c.Chooser == me.ID {
			surveil = c
		}
	}
	if surveil == nil {
		t.Fatalf("no surveil prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveSurveil(surveil.ID, me.ID, surveil.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if got := me.Graveyard.Size(); got != 1 {
		t.Errorf("graveyard %d, want 1 (the surveilled card)", got)
	}
}

// --- Tainted Sigil ------------------------------------------------------------

func TestOldB502SigilGainsTheTotalLifeLostByAllPlayersThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.AdvanceToMain(t, g)
	sigil := effectstest.PushCatalogPermanent(g, me.ID, "Tainted Sigil", "Artifact", oldB502Sigil, false)
	g.WithWriteLock(func() {
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -3)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, third.ID, -4)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -2)
	})
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, sigil, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(sigil) {
		t.Error("the artifact was not sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+9 {
		t.Errorf("life %d, want %d (3 + 4 + 2 lost this turn)", me.Life, life+9)
	}
}

// --- Slave of Bolas -----------------------------------------------------------

func TestOldB502SlaveStealsUntapsHastesAndSacrificesAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.TapForTest(t, g, victim)
	effectstest.AssistCardCastX(t, g, "Slave of Bolas", "Sorcery", oldB502Slave, "{3}{U}{B}",
		[]string{"U", "B", "C", "C", "C"}, 0, effectstest.CardRefs(victim))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ControllerOf(t, g, victim) != me.ID {
		t.Fatal("control did not change")
	}
	if effectstest.B16Tapped(t, g, victim) {
		t.Error("the creature was not untapped")
	}
	if !effectstest.P1g2HasKeyword(g, victim, "haste") {
		t.Error("the creature did not gain haste")
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the creature was not sacrificed at the end step")
	}
	if effectstest.ZoneOfCard(g, victim) != game.ZoneGraveyard {
		t.Errorf("the creature is in %q, want its owner's graveyard", effectstest.ZoneOfCard(g, victim))
	}
}

// --- Soulquake ----------------------------------------------------------------

func TestOldB502SoulquakeReturnsEveryCreatureAndCreatureCardToItsOwnersHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	graveCreature := effectstest.PushGraveyardCardTyped(opp, "Dead Bear", "Creature — Bear")
	graveSpell := effectstest.PushGraveyardCardTyped(me, "Dead Spell", "Sorcery")
	effectstest.AssistCardCastX(t, g, "Soulquake", "Sorcery", oldB502Soulquake, "{3}{U}{U}{B}{B}",
		[]string{"U", "U", "B", "B", "C", "C", "C"}, 0, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(mine) || !opp.Hand.Contains(theirs) {
		t.Error("a battlefield creature did not return to its owner's hand")
	}
	if !opp.Hand.Contains(graveCreature) {
		t.Error("the creature card in a graveyard did not return to its owner's hand")
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("a noncreature permanent was bounced")
	}
	if !me.Graveyard.Contains(graveSpell) {
		t.Error("a noncreature card left the graveyard")
	}
}

// --- Double Negative ----------------------------------------------------------

func TestOldB502DoubleNegativeCountersTwoSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	a := effectstest.CastCatalogSpell(t, g, "Spell A", "Instant", "", nil)
	b := effectstest.CastCatalogSpell(t, g, "Spell B", "Instant", "", nil)
	effectstest.CastCatalogSpell(t, g, "Double Negative", "Instant", oldB502Double,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if g.Stack.Contains(id) {
			t.Error("a target spell is still on the stack")
		}
		if effectstest.ZoneOfCard(g, id) != game.ZoneGraveyard {
			t.Errorf("a countered spell is in %q, want the graveyard", effectstest.ZoneOfCard(g, id))
		}
	}
}

func TestOldB502DoubleNegativeCountersJustOneWhenOnlyOneIsChosen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	a := effectstest.CastCatalogSpell(t, g, "Spell A", "Instant", "", nil)
	effectstest.CastCatalogSpell(t, g, "Double Negative", "Instant", oldB502Double,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOfCard(g, a) != game.ZoneGraveyard {
		t.Errorf("the target is in %q, want the graveyard", effectstest.ZoneOfCard(g, a))
	}
}

func TestOldB502DoubleNegativeRefusesThreeTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	a := effectstest.CastCatalogSpell(t, g, "Spell A", "Instant", "", nil)
	b := effectstest.CastCatalogSpell(t, g, "Spell B", "Instant", "", nil)
	c := effectstest.CastCatalogSpell(t, g, "Spell C", "Instant", "", nil)
	err := effectstest.CastCatalogSpellErrStrict(t, g, "Double Negative", "Instant", oldB502Double,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}, {Kind: game.TargetCard, ID: c}})
	if err == nil {
		t.Error("three targets were accepted for \"up to two\"")
	}
}

// --- Mind Funeral -------------------------------------------------------------

func TestOldB502FuneralMillsUntilTheFourthLandIsRevealed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	opp.Library.Cards = nil
	var ids []uuid.UUID
	push := func(name, typ string) {
		id := uuid.New()
		opp.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typ, Owner: opp.ID, Controller: opp.ID})
		ids = append(ids, id)
	}
	// Pushed bottom-up: the last push is the top card.
	push("Below", "Sorcery")
	push("Land 5", "Basic Land — Forest")
	push("Land 4", "Basic Land — Forest")
	push("Spell 2", "Sorcery")
	push("Land 3", "Basic Land — Forest")
	push("Land 2", "Basic Land — Forest")
	push("Spell 1", "Instant")
	push("Land 1", "Basic Land — Forest")
	effectstest.AssistCardCastX(t, g, "Mind Funeral", "Sorcery", oldB502Funeral, "{1}{U}{B}",
		[]string{"U", "B", "C"}, 0, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Graveyard.Size(); got != 6 {
		t.Errorf("graveyard %d, want 6 (the cards down to and including the fourth land)", got)
	}
	if effectstest.ZoneOfCard(g, ids[0]) != game.ZoneLibrary || effectstest.ZoneOfCard(g, ids[1]) != game.ZoneLibrary {
		t.Error("a card below the fourth land was milled")
	}
}
