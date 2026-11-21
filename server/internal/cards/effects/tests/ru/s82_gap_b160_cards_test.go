package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b160_cards_test.go — GAP batch 160 (#5948): five Jumpstart 2022 blue
// creatures that were not in the catalog. Each test plays the real catalog
// entry through the engine and asserts what its text says.

const (
	gapB160BarrinOracle    = "f98f7cb7-90ca-41e7-9f1d-8054931ecf0e"
	gapB160StriderOracle   = "9bdd68e7-7787-4579-a128-09e78fdbf0e2"
	gapB160KrakenOracle    = "a1bdba37-f034-412f-8395-adbaed95ad11"
	gapB160ChillerOracle   = "5607bc18-f3b7-40b0-b97d-2e06affa99ac"
	gapB160YetiOracle      = "0363363b-a538-4cf6-ae6b-6b45cdabd99a"
	gapB160SnowSubtype     = "Island"
	gapB160SnowCreatureTyp = "Snow Creature — "
)

func TestGapB160EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB160BarrinOracle:  "Barrin, Tolarian Archmage",
		gapB160StriderOracle: "Berg Strider",
		gapB160KrakenOracle:  "Biblioplex Kraken",
		gapB160ChillerOracle: "Chillerpillar",
		gapB160YetiOracle:    "Frostpeak Yeti",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v registered=%v, want Full with no caveats", name, spec, ok)
		}
	}
}

func gapB160Lands(g *game.Game, owner uuid.UUID, n int, snow bool) {
	for i := 0; i < n; i++ {
		if snow {
			effectstest.BlbB1Push(g, owner, "Snow-Covered Island", "Basic Snow Land — Island", "", 0, 0)
		} else {
			effectstest.BlbB1Push(g, owner, "Island", "Basic Land — Island", "", 0, 0)
		}
	}
}

func TestGapB160BarrinBouncesThenDrawsOnYourEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	barrin := effectstest.CastHoldCreature(t, g, "Barrin, Tolarian Archmage", "Legendary Creature — Human Wizard", gapB160BarrinOracle, 2, 2)
	ch := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: barrin}); err == nil {
		t.Fatal("Barrin was accepted as the other target")
	}
	effectstest.AnswerPickTarget(t, g, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(mine) {
		t.Fatal("the creature was not returned to its owner's hand")
	}
	hand := me.Hand.Size()
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d at the end step, want one card drawn", hand, got)
	}

	// A permanent put into an opponent's hand draws nothing for you, and
	// neither does a turn with no bounce.
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[g2.Turn.ActiveSeat], g2.Seats[(g2.Turn.ActiveSeat+1)%len(g2.Seats)]
	theirs := effectstest.PushVanillaCreature(g2, opp2.ID, "Their Bear", 2, 2)
	effectstest.CastHoldCreature(t, g2, "Barrin, Tolarian Archmage", "Legendary Creature — Human Wizard", gapB160BarrinOracle, 2, 2)
	effectstest.AnswerPickTarget(t, g2, theirs)
	effectstest.PassPriorityAroundTable(t, g2)
	if !opp2.Hand.Contains(theirs) {
		t.Fatal("the opponent's creature was not bounced")
	}
	hand2 := me2.Hand.Size()
	effectstest.AdvanceTo(t, g2, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g2)
	if got := me2.Hand.Size(); got != hand2 {
		t.Errorf("hand %d -> %d: a permanent put into an opponent's hand draws nothing", hand2, got)
	}
	_ = opp
}

func gapB160Strider(t *testing.T, snow bool) (*game.Game, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.ToMainForCost(t, g)
	gapB160Lands(g, me.ID, 5, snow)
	id := effectstest.HandCardFull(me, "Berg Strider", "Snow Creature — Giant Wizard", "{4}{U}", gapB160StriderOracle, []string{"U"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	return g, victim
}

func TestGapB160BergStriderTapsAndFreezesOnlyWithSnowMana(t *testing.T) {
	for _, snow := range []bool{true, false} {
		g, victim := gapB160Strider(t, snow)
		if c := effectstest.FindBattlefieldCardByID(g, victim); c == nil || !c.Tapped {
			t.Fatalf("snow=%v: the target was not tapped", snow)
		}
		effectstest.AdvanceToNextSeatsTurn(t, g)
		effectstest.AdvanceTo(t, g, game.StepUpkeep)
		c := effectstest.FindBattlefieldCardByID(g, victim)
		if c == nil {
			t.Fatalf("snow=%v: the target left the battlefield", snow)
		}
		if c.Tapped != snow {
			t.Errorf("snow=%v: tapped after its controller's untap step = %v, want %v", snow, c.Tapped, snow)
		}
	}
}

func TestGapB160BiblioplexKrakenScriesAndMayBounceForEvasion(t *testing.T) {
	for _, decline := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		kraken := effectstest.B12Push(g, me.ID, "Biblioplex Kraken", "Creature — Kraken", gapB160KrakenOracle, 4, 4)
		other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
		blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
		effectstest.DeclareAttack(t, g, opp.ID, kraken)
		effectstest.PassPriorityAroundTable(t, g)
		pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
		if pick == nil {
			t.Fatal("no creature choice")
		}
		if effectstest.HasID(pick.ChooseCards, kraken) || !effectstest.HasID(pick.ChooseCards, other) {
			t.Fatalf("offered %v", pick.ChooseCards)
		}
		picks := []uuid.UUID{other}
		if decline {
			picks = nil
		}
		if err := g.ResolveOwnPermanents(pick.ID, me.ID, picks); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := me.Hand.Contains(other); got == decline {
			t.Errorf("decline=%v: the creature is in hand = %v", decline, got)
		}
		c, _ := effectstest.BattlefieldCard(g, kraken)
		if got := c.Effective().Restrictions&game.CantBeBlocked != 0; got == decline {
			t.Errorf("decline=%v: can't be blocked = %v", decline, got)
		}
		_ = blocker
	}
}

func TestGapB160BiblioplexKrakenScriesThreeWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := me.Library.Size()
	effectstest.CastHoldCreature(t, g, "Biblioplex Kraken", "Creature — Kraken", gapB160KrakenOracle, 4, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID); got == nil {
		t.Errorf("no scry prompt after the Kraken entered (library %d)", before)
	}
}

func TestGapB160ChillerpillarGoesMonstrousForSnowAndFlies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.BlbB1Push(g, me.ID, "Chillerpillar", gapB160SnowCreatureTyp+"Insect", gapB160ChillerOracle, 3, 3)
	effectstest.AdvanceToMain(t, g)
	if effectstest.HasKeywordOnBattlefield(t, g, id, "flying") {
		t.Fatal("flying before it is monstrous")
	}
	gapB160Lands(g, me.ID, 4, false)
	gapB160Lands(g, me.ID, 1, true)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err == nil {
		t.Fatal("activated with only one snow mana for {S}{S}")
	}
	gapB160Lands(g, me.ID, 1, true)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, id, game.CounterPlusOne); n != 2 {
		t.Errorf("%d +1/+1 counters, want 2", n)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, id, "flying") {
		t.Error("no flying once monstrous")
	}
}

func TestGapB160FrostpeakYetiCantBeBlockedForOneAndSnow(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	yeti := effectstest.BlbB1Push(g, me.ID, "Frostpeak Yeti", gapB160SnowCreatureTyp+"Yeti", gapB160YetiOracle, 3, 3)
	effectstest.AdvanceToMain(t, g)
	gapB160Lands(g, me.ID, 2, false)
	if err := g.ActivateCatalogAbility(me.ID, yeti, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err == nil {
		t.Fatal("activated with no snow mana for {S}")
	}
	gapB160Lands(g, me.ID, 1, true)
	if err := g.ActivateCatalogAbility(me.ID, yeti, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCard(g, yeti)
	if c.Effective().Restrictions&game.CantBeBlocked == 0 {
		t.Error("the Yeti can still be blocked")
	}
}
