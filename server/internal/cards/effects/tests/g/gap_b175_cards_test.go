package testsg

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gap_b175_cards_test.go — Jumpstart 2022 cards, S82 GAP batch 175
// (#5948). Every test plays the real catalog entry through the engine.

const (
	gapKasminaOracle          = "df561b38-fe2a-4db9-a9d4-ec25b1c7572a"
	gapTamiyoMoonSageOracle   = "f88be051-36ae-407e-bd5a-2b029c26414e"
	gapTezzeretArtificeOracle = "e22824cf-07a1-4c83-b6c4-9d8fcff3892f"
	gapAjaniStrengthOracle    = "c3974ec7-7af0-4b65-baf6-4c1c4fe8a5c4"
	gapBloodPriceOracle       = "93c6554f-096d-4556-81ca-e87e29265887"
	gapConsignToThePitOracle  = "b0c50079-5376-47ff-82c5-d52dbf49afdf"
	gapFeastOfBloodOracle     = "fed05740-01b2-4c7a-8b97-55e64837c07f"
	gapGiantGrowthOracle      = "5748ebf1-24e3-499d-ab7c-c2cebd462a24"
)

func TestGapB175CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapKasminaOracle: "Kasmina, Enigmatic Mentor", gapTamiyoMoonSageOracle: "Tamiyo, the Moon Sage",
		gapTezzeretArtificeOracle: "Tezzeret, Artifice Master", gapAjaniStrengthOracle: "Ajani, Strength of the Pride",
		gapBloodPriceOracle: "Blood Price", gapConsignToThePitOracle: "Consign to the Pit", gapFeastOfBloodOracle: "Feast of Blood",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registration %+v", name, spec)
		}
	}
}

func TestGapBloodPriceTakesTwoOfTheTopFourAndCostsTwoLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	for _, n := range []string{"One", "Two", "Three", "Four"} {
		effectstest.StapleLibraryCard(me, n, "Creature — Bear")
	}
	life, hand, library := me.Life, len(me.Hand.Cards), len(me.Library.Cards)
	effectstest.CastCatalogSpell(t, g, "Blood Price", "Sorcery", gapBloodPriceOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil || len(pick.ChooseCards) != 4 || pick.ChooseMin != 2 || pick.ChooseMax != 2 {
		t.Fatalf("want a mandatory pick of two from four, got %+v", pick)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, pick.ChooseCards[:1]); err == nil {
		t.Error("one card was accepted for a take of two")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, pick.ChooseCards[:2]); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if got := len(me.Hand.Cards); got != hand+2 {
		t.Errorf("hand = %d, want %d", got, hand+2)
	}
	if got := len(me.Library.Cards); got != library-2 {
		t.Errorf("library = %d, want %d (the rest went to the bottom)", got, library-2)
	}
	if me.Life != life-2 {
		t.Errorf("life = %d, want %d", me.Life, life-2)
	}
}

func TestGapConsignToThePitDestroysAndHitsTheController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp, _ := effectstest.BlbB1Seats(g)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	land := effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Land", TypeLine: "Land"})
	if err := effectstest.CastCatalogSpellErr(t, g, "Consign to the Pit", "Sorcery", gapConsignToThePitOracle, effectstest.CardTarget(land)); err == nil {
		t.Fatal("Consign to the Pit targeted a land")
	}
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Consign to the Pit", "Sorcery", gapConsignToThePitOracle, effectstest.CardTarget(ogre))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Error("the creature survived")
	}
	if opp.Life != life-2 {
		t.Errorf("the controller's life = %d, want %d", opp.Life, life-2)
	}
}

func TestGapFeastOfBloodNeedsTwoVampires(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.B12Creature(g, me.ID, "Vampire One", "Creature — Vampire", 2, 2)
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	spell := effectstest.PushCatalogHandCard(me, "Feast of Blood", "Sorcery", gapFeastOfBloodOracle)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(ogre)}); err == nil {
		t.Fatal("cast with a single Vampire")
	}
	effectstest.B12Creature(g, me.ID, "Vampire Two", "Creature — Vampire Rogue", 2, 2)
	life := me.Life
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(ogre)}); err != nil {
		t.Fatalf("cast with two Vampires: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Error("the creature survived")
	}
	if me.Life != life+4 {
		t.Errorf("life = %d, want %d", me.Life, life+4)
	}
}

func TestGapKasminaTaxesSpellsThatTargetYourCreaturesAndMakesAWizard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	kasmina := effectstest.PushWalkerForTest(g, me.ID, "Kasmina, Enigmatic Mentor", gapKasminaOracle, 5)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.EndTurn(t, g)
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	cast := func(target uuid.UUID, mana int) error {
		spell := uuid.New()
		opp.Hand.PushTop(game.Card{InstanceID: spell, Name: "Giant Growth", TypeLine: "Instant", OracleID: gapGiantGrowthOracle,
			ManaCost: "{G}", Owner: opp.ID, Controller: opp.ID})
		opp.ManaPool.EmptyPool()
		opp.ManaPool.AddMana(game.ManaToken{Color: "G"})
		for i := 1; i < mana; i++ {
			opp.ManaPool.AddMana(game.ManaToken{Color: "C"})
		}
		return g.CastSpell(opp.ID, spell, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(target)})
	}
	if err := cast(theirs, 1); err != nil {
		t.Fatalf("a spell at the opponent's own creature costs {G}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := cast(mine, 1); err == nil {
		t.Error("a spell at Kasmina's controller's creature cost only {G}")
	}
	if err := cast(mine, 3); err != nil {
		t.Errorf("the taxed spell for {2}{G}: %v", err)
	}
	_ = kasmina
}

func TestGapKasminaMinusTwoMakesAWizardAndLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	kasmina := effectstest.PushWalkerForTest(g, me.ID, "Kasmina, Enigmatic Mentor", gapKasminaOracle, 5)
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	hand := len(me.Hand.Cards)
	if err := g.ActivateCatalogAbility(me.ID, kasmina, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("-2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var wizard uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Wizard" {
			wizard = c.InstanceID
		}
	}
	if wizard == uuid.Nil {
		t.Fatal("no Wizard token")
	}
	if p, tou := effectstest.PtOf(t, g, wizard); p != 2 || tou != 2 {
		t.Errorf("Wizard = %d/%d, want 2/2", p, tou)
	}
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("no discard after the draw")
	}
	if len(me.Hand.Cards) != hand+1 {
		t.Errorf("hand = %d, want %d before discarding", len(me.Hand.Cards), hand+1)
	}
	if got := effectstest.LoyaltyOf(g, kasmina); got != 3 {
		t.Errorf("loyalty = %d, want 3", got)
	}
}

func TestGapTamiyoThePlusOneTapsAndFreezes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	tamiyo := effectstest.PushWalkerForTest(g, me.ID, "Tamiyo, the Moon Sage", gapTamiyoMoonSageOracle, 4)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	if err := g.ActivateCatalogAbility(me.ID, tamiyo, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B13Tapped(t, g, bear) {
		t.Fatal("the permanent was not tapped")
	}
	// Through its controller's next untap step it stays tapped.
	for g.Turn.ActiveSeat != 1 || g.Turn.Step != game.StepUpkeep {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if !effectstest.B13Tapped(t, g, bear) {
		t.Error("it untapped in its controller's next untap step")
	}
}

func TestGapTamiyoTheMinusTwoDrawsForTappedCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	tamiyo := effectstest.PushWalkerForTest(g, me.ID, "Tamiyo, the Moon Sage", gapTamiyoMoonSageOracle, 4)
	for i := 0; i < 3; i++ {
		id := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
		if i < 2 {
			g.WithWriteLock(func() {
				for j := range g.Battlefield.Cards {
					if g.Battlefield.Cards[j].InstanceID == id {
						g.Battlefield.Cards[j].Tapped = true
					}
				}
			})
		}
	}
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	hand := len(me.Hand.Cards)
	if err := g.ActivateCatalogAbility(me.ID, tamiyo, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("-2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != hand+2 {
		t.Errorf("hand = %d, want %d (two tapped creatures)", got, hand+2)
	}
}

func TestGapTamiyoTheEmblemKeepsYourHandAndReturnsGraveyardCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	tamiyo := effectstest.PushWalkerForTest(g, me.ID, "Tamiyo, the Moon Sage", gapTamiyoMoonSageOracle, 8)
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	if err := g.ActivateCatalogAbility(me.ID, tamiyo, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("-8: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.EmblemsOf(g, me.ID)); got != 1 {
		t.Fatalf("emblems = %d, want 1", got)
	}
	if got := effectstest.MhMax(g, me); got != game.NoMaxHandSize {
		t.Errorf("max hand size = %d, want no maximum", got)
	}
	// A discarded card arrives in the graveyard: the emblem offers it back.
	me.Hand.Cards = nil
	discard := effectstest.PushCatalogHandCard(me, "Discard Me", "Creature — Bear", "")
	g.WithWriteLock(func() {
		g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: me.ID, Source: tamiyo, N: 1})
	})
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestTriggerPromptFor(g, me.ID)
	if prompt == nil {
		t.Fatal("the emblem's trigger asked nothing")
	}
	if err := g.ResolveTriggerPrompt(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveTriggerPrompt: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(discard) {
		t.Error("the card did not return to hand")
	}
}

func TestGapTezzeretArtificeMasterPlusOneAndZero(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	tez := effectstest.PushWalkerForTest(g, me.ID, "Tezzeret, Artifice Master", gapTezzeretArtificeOracle, 5)
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	if err := g.ActivateCatalogAbility(me.ID, tez, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var thopter uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Thopter" {
			thopter = c.InstanceID
		}
	}
	if thopter == uuid.Nil {
		t.Fatal("no Thopter token")
	}
	if !effectstest.P1g2HasKeyword(g, thopter, "flying") || !effectstest.LayeredCard(t, g, thopter).IsArtifact() {
		t.Error("the Thopter is not a flying artifact creature")
	}
	// Zero: one card with fewer than three artifacts, two with three or more.
	// (Only one loyalty ability a turn: the next turn of mine.)
	for i := 0; i < 4; i++ {
		effectstest.EndTurn(t, g)
	}
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	hand := len(me.Hand.Cards)
	if err := g.ActivateCatalogAbility(me.ID, tez, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("0: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("with one artifact the hand = %d, want %d", got, hand+1)
	}
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
	for i := 0; i < 4; i++ {
		effectstest.EndTurn(t, g)
	}
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	hand = len(me.Hand.Cards)
	if err := g.ActivateCatalogAbility(me.ID, tez, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("0 (second): %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != hand+2 {
		t.Errorf("with three artifacts the hand = %d, want %d", got, hand+2)
	}
}

func TestGapTezzeretArtificeMasterEmblemFetchesAPermanentAtEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	tez := effectstest.PushWalkerForTest(g, me.ID, "Tezzeret, Artifice Master", gapTezzeretArtificeOracle, 9)
	spell := effectstest.StapleLibraryCard(me, "Some Instant", "Instant")
	bear := effectstest.StapleLibraryCard(me, "Library Bear", "Creature — Bear")
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	if err := g.ActivateCatalogAbility(me.ID, tez, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("-9: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.EmblemsOf(g, me.ID)); got != 1 {
		t.Fatalf("emblems = %d, want 1", got)
	}
	for g.Turn.Step != game.StepEnd {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	if _, ok := effectstest.AangCardOnBF(g, bear); !ok {
		t.Error("the permanent card was not put onto the battlefield")
	}
	if _, ok := effectstest.AangCardOnBF(g, spell); ok {
		t.Error("a nonpermanent card was fetched")
	}
}

func TestGapAjaniStrengthOfThePrideGainsLifeFromCreaturesAndWalkers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	ajani := effectstest.PushWalkerForTest(g, me.ID, "Ajani, Strength of the Pride", gapAjaniStrengthOracle, 5)
	effectstest.B12Creature(g, me.ID, "Bear One", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Bear Two", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.PushWalkerForTest(g, me.ID, "Other Walker", "", 3)
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, ajani, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("life = %d, want %d (two creatures and two planeswalkers)", me.Life, life+4)
	}
}

func TestGapAjaniStrengthOfThePrideMakesAPridemate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	ajani := effectstest.PushWalkerForTest(g, me.ID, "Ajani, Strength of the Pride", gapAjaniStrengthOracle, 5)
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	if err := g.ActivateCatalogAbility(me.ID, ajani, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("-2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pm, ok := effectstest.VisBattlefieldNamed(g, me.ID, "Ajani's Pridemate")
	if !ok {
		t.Fatal("no Ajani's Pridemate token")
	}
	if !slices.Contains(pm.Effective().Subtypes, "Cat") || !slices.Contains(pm.Effective().Subtypes, "Soldier") {
		t.Errorf("subtypes = %v", pm.Effective().Subtypes)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, pm.InstanceID, game.CounterPlusOne); n != 1 {
		t.Errorf("Pridemate counters = %d, want 1", n)
	}
}

func TestGapAjaniStrengthOfThePrideZeroNeedsFifteenAboveStarting(t *testing.T) {
	for _, tc := range []struct {
		name string
		life int
		wipe bool
	}{{"short", game.StartingLife + 14, false}, {"enough", game.StartingLife + 15, true}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp, third := effectstest.BlbB1Seats(g)
			ajani := effectstest.PushWalkerForTest(g, me.ID, "Ajani, Strength of the Pride", gapAjaniStrengthOracle, 5)
			theirCreature := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
			theirArtifact := effectstest.PushPermanent(g, third.ID, game.Card{Name: "Their Rock", TypeLine: "Artifact"})
			theirLand := effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Their Land", TypeLine: "Land"})
			mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
			me.Life = tc.life
			effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
			if err := g.ActivateCatalogAbility(me.ID, ajani, 2, game.ActivateAbilityParams{}); err != nil {
				t.Fatalf("0: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			gone := !g.Battlefield.Contains(theirCreature) && !g.Battlefield.Contains(theirArtifact) && !g.Battlefield.Contains(ajani)
			if gone != tc.wipe {
				t.Errorf("exiled = %v, want %v", gone, tc.wipe)
			}
			if !g.Battlefield.Contains(theirLand) || !g.Battlefield.Contains(mine) {
				t.Error("a land or my own creature was exiled")
			}
		})
	}
}
