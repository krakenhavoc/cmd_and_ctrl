package testss82m

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b093_cards_test.go — Zendikar Rising, batch 093 (#5508). Every test
// plays the real catalog entry through the engine.

const (
	oldB093AcquisitionsExpertOracle = "07e5f59e-3a98-4e01-a368-0be60e0463ba"
	oldB093CascadeSeerOracle        = "08992e72-49b5-4d40-8b44-cb5abfe5ce0d"
	oldB093TazeemRaptorOracle       = "3111f8b7-36af-4390-af99-5b8976b2f241"
	oldB093LegionAngelOracle        = "843c9962-dee1-449d-8106-6383df32bcf6"
	oldB093DranasSilencerOracle     = "8e993c7a-606d-47f0-bdea-6ad197a59426"
	oldB093ZulaportDuelistOracle    = "e8237d12-2860-4cc8-85a4-458e4c68e69a"
)

func TestOldB093EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		oldB093AcquisitionsExpertOracle, oldB093CascadeSeerOracle, oldB093TazeemRaptorOracle,
		oldB093LegionAngelOracle, oldB093DranasSilencerOracle, oldB093ZulaportDuelistOracle,
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

// --- Acquisitions Expert ------------------------------------------------------

func oldB093Hand(opp *game.Player, n int) []uuid.UUID {
	opp.Hand.Cards = nil
	var ids []uuid.UUID
	for i := 1; i <= n; i++ {
		id := uuid.New()
		opp.Hand.PushTop(game.Card{InstanceID: id, Name: fmt.Sprintf("Their Card %d", i), TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
		ids = append(ids, id)
	}
	return ids
}

func TestOldB093AcquisitionsExpertRevealsOnePerPartyMemberAndDiscardsTheChosenOne(t *testing.T) {
	for _, extra := range []int{0, 1} {
		t.Run(fmt.Sprintf("extra=%d", extra), func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
			if extra == 1 {
				effectstest.B12Creature(g, me.ID, "Healer", "Creature — Human Cleric", 1, 1)
			}
			hand := oldB093Hand(opp, 4)
			effectstest.CastHoldCreature(t, g, "Acquisitions Expert", "Creature — Human Rogue", oldB093AcquisitionsExpertOracle, 1, 2)
			effectstest.PickPlayer(t, g, me.ID, opp.ID)
			effectstest.PassPriorityAroundTable(t, g)
			n := 1 + extra
			ask := effectstest.ChooseCardsChoiceFor(g, opp.ID)
			if ask == nil || ask.ChooseMin != n || ask.ChooseMax != n {
				t.Fatalf("the opponent chooses exactly %d cards to reveal, got %+v", n, ask)
			}
			reveal := hand[:n]
			effectstest.AnswerChooseCards(t, g, opp.ID, reveal...)
			rp := effectstest.RevealPickFor(g, me.ID)
			if rp == nil || len(rp.ChooseCards) != n {
				t.Fatalf("want %d revealed cards for the Expert's controller, got %+v", n, rp)
			}
			if err := g.ResolveRevealPick(rp.ID, me.ID, []uuid.UUID{hand[3]}); err == nil {
				t.Fatal("the controller picked a card that was not revealed")
			}
			if err := g.ResolveRevealPick(rp.ID, me.ID, []uuid.UUID{reveal[n-1]}); err != nil {
				t.Fatalf("ResolveRevealPick: %v", err)
			}
			if !opp.Graveyard.Contains(reveal[n-1]) || opp.Hand.Size() != 3 {
				t.Errorf("the chosen card was not discarded (hand %d)", opp.Hand.Size())
			}
		})
	}
}

func TestOldB093AcquisitionsExpertAgainstAnEmptyHandDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	oldB093Hand(opp, 0)
	effectstest.CastHoldCreature(t, g, "Acquisitions Expert", "Creature — Human Rogue", oldB093AcquisitionsExpertOracle, 1, 2)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, opp.ID) != nil || effectstest.RevealPickFor(g, me.ID) != nil {
		t.Fatal("an empty hand raised a prompt")
	}
}

// --- Cascade Seer -------------------------------------------------------------

func TestOldB093CascadeSeerScriesThePartySize(t *testing.T) {
	for _, extra := range []int{0, 2} {
		t.Run(fmt.Sprintf("extra=%d", extra), func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := effectstest.Top100ActiveSeat(g)
			if extra == 2 {
				effectstest.B12Creature(g, me.ID, "Healer", "Creature — Human Cleric", 1, 1)
				effectstest.B12Creature(g, me.ID, "Thief", "Creature — Human Rogue", 1, 1)
			}
			effectstest.SeedLibrary(me, "A", "B", "C", "D")
			effectstest.CastHoldCreature(t, g, "Cascade Seer", "Creature — Merfolk Wizard", oldB093CascadeSeerOracle, 3, 3)
			c := effectstest.ScryChoiceFor(g, me.ID)
			if c == nil {
				t.Fatal("no scry prompt")
			}
			if want := 1 + extra; len(c.ScryCards) != want {
				t.Errorf("scrying %d cards, want %d", len(c.ScryCards), want)
			}
		})
	}
}

// --- Tazeem Raptor ------------------------------------------------------------

func TestOldB093TazeemRaptorMayReturnALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	land := effectstest.B12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	raptor := effectstest.CastHoldCreature(t, g, "Tazeem Raptor", "Creature — Bird", oldB093TazeemRaptorOracle, 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, raptor, "flying") {
		t.Error("no flying")
	}
	effectstest.AnswerOwnPermanents(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(land) || g.Battlefield.Contains(land) {
		t.Error("the land did not return to its owner's hand")
	}
}

func TestOldB093TazeemRaptorMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	land := effectstest.B12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	effectstest.CastHoldCreature(t, g, "Tazeem Raptor", "Creature — Bird", oldB093TazeemRaptorOracle, 2, 2)
	effectstest.AnswerOwnPermanents(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(land) {
		t.Error("declining still returned the land")
	}
}

// --- Legion Angel -------------------------------------------------------------

func TestOldB093LegionAngelFetchesOnlyACardNamedLegionAngel(t *testing.T) {
	g := effectstest.NewSideboardCatalogGame(t, 2,
		effectstest.OutsideCard("Lightning Bolt", "Instant"), effectstest.OutsideCard("Legion Angel", "Creature — Angel Warrior"))
	me := g.Seats[0]
	hand := me.Hand.Size()
	effectstest.TdCastFromHand(t, g, game.Card{Name: "Legion Angel", TypeLine: "Creature — Angel Warrior",
		OracleID: oldB093LegionAngelOracle, Power: 4, Toughness: 3}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatalf("no outside-the-game prompt: %+v", g.PendingChoices)
	}
	if got := effectstest.CandidateNames(g, c); !effectstest.EqualNames(got, []string{"Legion Angel"}) {
		t.Errorf("candidates %v, want only Legion Angel", got)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{effectstest.SideCardID(me, "Legion Angel")}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OutsideHandHas(me, "Legion Angel") || me.Hand.Size() != hand+1 {
		t.Errorf("hand %d (had %d before the cast, want one more), has the Angel %v", me.Hand.Size(), hand, effectstest.OutsideHandHas(me, "Legion Angel"))
	}
	if got := effectstest.SideNamed(me); !effectstest.EqualNames(got, []string{"Lightning Bolt"}) {
		t.Errorf("sideboard %v, want only the Bolt left", got)
	}
}

// --- Drana's Silencer ---------------------------------------------------------

func TestOldB093DranasSilencerShrinksByThePartySize(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	effectstest.B12Creature(g, me.ID, "Healer", "Creature — Human Cleric", 1, 1)
	victim := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 4, 4)
	effectstest.CastHoldCreature(t, g, "Drana's Silencer", "Creature — Vampire Rogue", oldB093DranasSilencerOracle, 3, 2)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
		t.Error("the Silencer accepted a creature its controller controls")
	}
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if pw, tough := effectstest.EffectivePTOf(t, g, victim); pw != 2 || tough != 2 {
		t.Errorf("the Ogre is %d/%d, want 2/2 (a party of two)", pw, tough)
	}
}

// --- Zulaport Duelist ---------------------------------------------------------

func TestOldB093ZulaportDuelistShrinksAndMillsTheTargetsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	victim := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	library, mine := opp.Library.Size(), me.Library.Size()
	duelist := effectstest.CastHoldCreature(t, g, "Zulaport Duelist", "Creature — Human Rogue", oldB093ZulaportDuelistOracle, 1, 1)
	if !effectstest.HasEffectiveKeyword(t, g, duelist, "flash") {
		t.Error("no flash")
	}
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, victim); p != 2 || tough != 4 {
		t.Errorf("the Ogre is %d/%d, want 2/4", p, tough)
	}
	if got := library - opp.Library.Size(); got != 2 {
		t.Errorf("the opponent milled %d, want 2", got)
	}
	if me.Library.Size() != mine {
		t.Error("the Duelist's controller milled")
	}
}

func TestOldB093ZulaportDuelistCanTargetYourOwnCreatureAndMillYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	library := me.Library.Size()
	effectstest.CastHoldCreature(t, g, "Zulaport Duelist", "Creature — Human Rogue", oldB093ZulaportDuelistOracle, 1, 1)
	effectstest.PickCard(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if got := library - me.Library.Size(); got != 2 {
		t.Errorf("you milled %d, want 2 (its controller is you)", got)
	}
}

func TestOldB093ZulaportDuelistWithNoTargetMillsNobody(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	lib, theirs := me.Library.Size(), opp.Library.Size()
	effectstest.CastHoldCreature(t, g, "Zulaport Duelist", "Creature — Human Rogue", oldB093ZulaportDuelistOracle, 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != lib || opp.Library.Size() != theirs {
		t.Error("a library changed with no target chosen")
	}
}
