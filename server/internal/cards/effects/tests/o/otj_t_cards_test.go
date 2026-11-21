package testso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// otj_t_cards_test.go — Outlaws of Thunder Junction (OTJ), the T cards
// (#4468). Each test plays the real catalog entry through the engine and
// asserts what the printed text says.

const (
	otjTTakeTheFallOracle   = "279f3aa5-5807-4797-8cd1-e009984a3691"
	otjTThisTownOracle      = "2550c2ae-834d-46e7-9ab4-d7bf7bfa6a78"
	otjTThrowOracle         = "d6780d29-0dd7-4f04-95b8-4ca6a55308f2"
	otjTLassoOracle         = "d98cc18b-6ebb-4f55-9d0d-b8fbe9a92bc1"
	otjTSalvoOracle         = "d3ce7c83-986f-410a-ae00-6c1c1f918a2b"
	otjTTinybonesJoinsUp    = "3d2bd4f6-174d-411a-b930-9161aeefd33a"
	otjTTombTrawlerOracle   = "d22d3934-8437-4a18-81c1-4e9c501846ca"
	otjTTrashTheTownOracle  = "b47f7f1e-2d81-4b1b-870d-4c1a4f020505"
	otjTDredgerOracle       = "9a691e96-fb47-4c23-a063-3b287c51147b"
	otjTTrickShotOracle     = "843c8485-21b8-45ef-9a0c-4967f07278bb"
	otjTTumbleweedOracle    = "ef574346-1d98-447d-acb6-4d8d53c5f651"
	otjTThisTownTypeLine    = "Instant"
	otjTThisTownPrintedCost = "{4}{U}"
)

func TestOtjTEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		otjTTakeTheFallOracle:  "Take the Fall",
		otjTThisTownOracle:     "This Town Ain't Big Enough",
		otjTThrowOracle:        "Throw from the Saddle",
		otjTLassoOracle:        "Thunder Lasso",
		otjTSalvoOracle:        "Thunder Salvo",
		otjTTinybonesJoinsUp:   "Tinybones Joins Up",
		otjTTombTrawlerOracle:  "Tomb Trawler",
		otjTTrashTheTownOracle: "Trash the Town",
		otjTDredgerOracle:      "Treasure Dredger",
		otjTTrickShotOracle:    "Trick Shot",
		otjTTumbleweedOracle:   "Tumbleweed Rising",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
	if sa := effectstest.PlotOf(otjTTumbleweedOracle); sa == nil || sa.Cost != "{2}{G}" {
		t.Errorf("Tumbleweed Rising: plot %+v, want {2}{G}", sa)
	}
}

// --- Take the Fall -----------------------------------------------------------

func TestOtjTTakeTheFallShrinksByOneOrFourWithAnOutlawAndDraws(t *testing.T) {
	for _, tc := range []struct {
		name   string
		outlaw bool
		want   int
	}{
		{"no outlaw", false, 4},
		{"an outlaw", true, 1},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.SeedLibrary(me, "One", "Two")
		if tc.outlaw {
			effectstest.B12Creature(g, me.ID, "My Rogue", "Creature — Human Rogue", 1, 1)
		}
		theirs := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)
		effectstest.AdvanceToMain(t, g)
		effectstest.CastCatalogSpell(t, g, "Take the Fall", "Instant", otjTTakeTheFallOracle, effectstest.B16TargetCard(theirs))
		hand := me.Hand.Size()
		effectstest.PassPriorityAroundTable(t, g)
		if p, tough := effectstest.EffectivePTOf(t, g, theirs); p != tc.want || tough != 5 {
			t.Errorf("%s: the Giant is %d/%d, want %d/5", tc.name, p, tough, tc.want)
		}
		if got := me.Hand.Size() - hand; got != 1 {
			t.Errorf("%s: drew %d, want 1", tc.name, got)
		}
	}
}

// --- This Town Ain't Big Enough ----------------------------------------------

func TestOtjTThisTownCostsThreeLessWhenItTargetsYourPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.AdvanceToMain(t, g)
	price := func(targets ...uuid.UUID) int {
		return effectstest.SelfPriced(t, g, me, otjTThisTownOracle, otjTThisTownTypeLine, otjTThisTownPrintedCost, game.ZoneHand, effectstest.CardRefs(targets...)).ManaValue()
	}
	if got := price(theirs); got != 5 {
		t.Errorf("only an opponent's permanent: %d, want 5", got)
	}
	if got := price(theirs, mine); got != 2 {
		t.Errorf("one of yours among the targets: %d, want 2", got)
	}
}

func TestOtjTThisTownReturnsUpToTwoNonlandPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	land := effectstest.B12Permanent(g, opp.ID, "Their Land", "Land")
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.CastCatalogSpellErr(t, g, "This Town Ain't Big Enough", otjTThisTownTypeLine, otjTThisTownOracle, effectstest.CardRefs(land)); err == nil {
		t.Error("a land was accepted as a nonland permanent")
	}
	effectstest.CastCatalogSpell(t, g, "This Town Ain't Big Enough", otjTThisTownTypeLine, otjTThisTownOracle, effectstest.CardRefs(mine, theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(mine) || !opp.Hand.Contains(theirs) {
		t.Error("each permanent returns to its owner's hand")
	}
	if _, ok := effectstest.BattlefieldCard(g, land); !ok {
		t.Error("the land stays")
	}
}

// --- Throw from the Saddle ---------------------------------------------------

func TestOtjTThrowFromTheSaddlePumpsOrCountersThenBites(t *testing.T) {
	for _, tc := range []struct {
		name     string
		typeLine string
		counters int
	}{
		{"not a Mount", "Creature — Bear", 0},
		{"a Mount", "Creature — Horse Mount", 1},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mine := effectstest.B12Creature(g, me.ID, "My Beast", tc.typeLine, 2, 2)
		theirs := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)
		effectstest.AdvanceToMain(t, g)
		effectstest.CastCatalogSpell(t, g, "Throw from the Saddle", "Sorcery", otjTThrowOracle, []game.TargetRef{
			{Kind: game.TargetCard, ID: mine, Slot: 0}, {Kind: game.TargetCard, ID: theirs, Slot: 1},
		})
		effectstest.PassPriorityAroundTable(t, g)
		if p, tough := effectstest.PtOf(t, g, mine); p != 3 || tough != 3 {
			t.Errorf("%s: the biter is %d/%d, want 3/3", tc.name, p, tough)
		}
		if got := effectstest.PlusOneCounters(g, mine); got != tc.counters {
			t.Errorf("%s: %d +1/+1 counters, want %d", tc.name, got, tc.counters)
		}
		if got := otjAC2Damage(t, g, theirs); got != 3 {
			t.Errorf("%s: the Giant has %d damage, want 3", tc.name, got)
		}
		if got := otjAC2Damage(t, g, mine); got != 0 {
			t.Errorf("%s: one-sided, but the biter took %d", tc.name, got)
		}
	}
}

// --- Thunder Lasso -----------------------------------------------------------

func TestOtjTThunderLassoAttachesOnEntryAndTapsAnAttackedPlayersCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	bear := effectstest.SeedBear(g, me.ID)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	bystander := effectstest.B12Creature(g, third.ID, "Third Bear", "Creature — Bear", 2, 2)

	lasso := effectstest.CastCatalogSpell(t, g, "Thunder Lasso", "Artifact — Equipment", otjTLassoOracle, nil)
	effectstest.B04WaitForPick(t, g, me.ID)
	if p := effectstest.LatestPickTarget(g, me.ID); effectstest.HasID(p.PickTargetCards, theirs) || !effectstest.HasID(p.PickTargetCards, bear) {
		t.Fatalf("the entry trigger targets a creature you control: %+v", p)
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if host := effectstest.AttachmentHostOf(t, g, lasso); host.ID != bear {
		t.Fatalf("the Lasso did not attach to the Bear: %+v", host)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("equipped P/T %d/%d, want 3/3", p, tough)
	}

	effectstest.DeclareAttack(t, g, opp.ID, bear)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if effectstest.HasID(p.PickTargetCards, bystander) || effectstest.HasID(p.PickTargetCards, bear) {
			t.Errorf("offered a creature the defending player does not control: %v", p.PickTargetCards)
		}
		effectstest.PickCard(t, g, me.ID, theirs)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.FindBattlefieldCardForTest(g, theirs).Tapped {
		t.Error("the defending player's creature was not tapped")
	}
	if effectstest.FindBattlefieldCardForTest(g, bystander).Tapped {
		t.Error("another player's creature was tapped")
	}
	spec, _ := effects.Lookup(otjTLassoOracle)
	if len(spec.Activated) != 1 || spec.Activated[0].Label != "Equip {2}" {
		t.Errorf("activated rows %+v, want Equip {2}", spec.Activated)
	}
}

// --- Thunder Salvo -----------------------------------------------------------

func TestOtjTThunderSalvoCountsYourOtherSpellsThisTurn(t *testing.T) {
	for _, tc := range []struct {
		name   string
		before int
		want   int
	}{
		{"its own cast only", 0, 2},
		{"after one other spell", 1, 3},
	} {
		g := effectstest.NewCatalogGame(t)
		opp := g.Seats[1]
		big := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 9, 9)
		effectstest.AdvanceToMain(t, g)
		for range tc.before {
			otjB1BoltOpponent(t, g, opp.ID)
		}
		effectstest.CastCatalogSpell(t, g, "Thunder Salvo", "Instant", otjTSalvoOracle, effectstest.B16TargetCard(big))
		effectstest.PassPriorityAroundTable(t, g)
		if got := otjAC2Damage(t, g, big); got != tc.want {
			t.Errorf("%s: %d damage, want %d", tc.name, got, tc.want)
		}
	}
}

// --- Tinybones Joins Up ------------------------------------------------------

func TestOtjTTinybonesJoinsUpMakesTheChosenPlayersDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third, fourth := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	for _, p := range []*game.Player{me, opp, third, fourth} {
		effectstest.SeedLibrary(p, "Card in hand")
		p.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Held", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Tinybones Joins Up", "Legendary Enchantment", otjTTinybonesJoinsUp, nil)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetPlayer, ID: third.ID},
	}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, opp.ID) == nil || effectstest.DiscardChoiceFor(g, third.ID) == nil {
		t.Fatal("each chosen player is asked to discard")
	}
	if effectstest.DiscardChoiceFor(g, me.ID) != nil || effectstest.DiscardChoiceFor(g, fourth.ID) != nil {
		t.Error("an unchosen player was asked to discard")
	}
}

func TestOtjTTinybonesJoinsUpMillsAndDrainsOnALegendaryCreatureOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.SeedLibrary(opp, "A", "B", "C")
	effectstest.SeedLibrary(third, "A", "B", "C")
	effectstest.PushCatalogPermanent(g, me.ID, "Tinybones Joins Up", "Legendary Enchantment", otjTTinybonesJoinsUp, false)
	effectstest.AdvanceToMain(t, g)

	effectstest.Msc17Play(t, g, me, game.Card{Name: "Plain Bear", TypeLine: "Creature — Bear", ManaCost: "{0}", Power: 2, Toughness: 2})
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Fatalf("a nonlegendary creature triggered it: %+v", p)
	}

	oppLib, thirdLib, oppLife, thirdLife, myLife := opp.Library.Size(), third.Library.Size(), opp.Life, third.Life, me.Life
	effectstest.Msc17Play(t, g, me, game.Card{Name: "A Legend", TypeLine: "Legendary Creature — Human Rogue", ManaCost: "{0}", Power: 1, Toughness: 1})
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := oppLib - opp.Library.Size(); got != 1 {
		t.Errorf("the chosen player milled %d, want 1", got)
	}
	if got := oppLife - opp.Life; got != 1 {
		t.Errorf("the chosen player lost %d life, want 1", got)
	}
	if third.Library.Size() != thirdLib || third.Life != thirdLife || me.Life != myLife {
		t.Error("an unchosen player milled or lost life")
	}
}

// --- Tomb Trawler ------------------------------------------------------------

func TestOtjTTombTrawlerBottomsACardFromYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	trawler := effectstest.B12Push(g, me.ID, "Tomb Trawler", "Artifact Creature — Golem", otjTTombTrawlerOracle, 0, 4)
	effectstest.SeedLibrary(me, "Top")
	dead := effectstest.PushGraveyardPermanent(me, "Dead Thing", "Creature — Bear", "{1}{G}")
	theirs := effectstest.PushGraveyardPermanent(opp, "Their Dead Thing", "Creature — Bear", "{1}{G}")
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, trawler, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(theirs)}); err == nil {
		t.Error("a card in an opponent's graveyard was accepted")
	}
	effectstest.B16Activate(t, g, me.ID, trawler, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(dead)})
	if me.Graveyard.Contains(dead) {
		t.Error("the card leaves the graveyard")
	}
	if got := effectstest.LibraryBottom(me, 1); len(got) != 1 || got[0] != dead {
		t.Errorf("library bottom %v, want the card", got)
	}
}

// --- Trash the Town ----------------------------------------------------------

func TestOtjTTrashTheTownAllThreeBullets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastModal(t, g, "Trash the Town", "Instant", otjTTrashTheTownOracle, []int{0, 1, 2}, []game.TargetRef{
		effectstest.ModeRef(game.TargetCard, bear, 0, 0),
		effectstest.ModeRef(game.TargetCard, bear, 1, 0),
		effectstest.ModeRef(game.TargetCard, bear, 2, 0),
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOneCounters(g, bear); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the creature gains trample")
	}
	hand := me.Hand.Size()
	effectstest.LciB1Attack(t, g, bear, opp.ID)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("combat damage to a player drew %d, want 2", got)
	}
}

func TestOtjTTrashTheTownGrantsTheDrawToItsTargetOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "One", "Two", "Three", "Four")
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastModal(t, g, "Trash the Town", "Instant", otjTTrashTheTownOracle, []int{2}, []game.TargetRef{
		effectstest.ModeRef(game.TargetCard, bear, 0, 0),
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOneCounters(g, bear); got != 0 {
		t.Errorf("an unchosen bullet put %d counters", got)
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("an unchosen bullet granted trample")
	}
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, bear, other)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("two creatures connected, only one with the grant: drew %d, want 2", got)
	}
}

// --- Treasure Dredger --------------------------------------------------------

func TestOtjTTreasureDredgerPaysOneLifeForATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dredger := effectstest.B12Push(g, me.ID, "Treasure Dredger", "Creature — Human Rogue", otjTDredgerOracle, 2, 2)
	effectstest.AdvanceToMain(t, g)
	life := me.Life
	effectstest.B06AddMana(me, "C")
	effectstest.B16Activate(t, g, me.ID, dredger, 0, game.ActivateAbilityParams{})
	if got := effectstest.TreasuresOf(g, me.ID); got != 1 {
		t.Errorf("Treasures = %d, want 1", got)
	}
	if got := life - me.Life; got != 1 {
		t.Errorf("paid %d life, want 1", got)
	}
	if !effectstest.B16Tapped(t, g, dredger) {
		t.Error("{T} is part of the cost")
	}
}

// --- Trick Shot --------------------------------------------------------------

func TestOtjTTrickShotHitsACreatureAndAnotherToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Beast", 7, 7)
	tok := effects.TokenCard("1/1 blue Fish")
	tok.Power, tok.Toughness = 3, 3
	fish := effectstest.PushToken(g, opp.ID, tok)
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.CastCatalogSpellErr(t, g, "Trick Shot", "Instant", otjTTrickShotOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: fish, Slot: 0}, {Kind: game.TargetCard, ID: fish, Slot: 1},
	}); err == nil {
		t.Error("the same token was accepted for both targets; the second is \"other\"")
	}
	effectstest.CastCatalogSpell(t, g, "Trick Shot", "Instant", otjTTrickShotOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: big, Slot: 0}, {Kind: game.TargetCard, ID: fish, Slot: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := otjAC2Damage(t, g, big); got != 6 {
		t.Errorf("the creature has %d damage, want 6", got)
	}
	if got := otjAC2Damage(t, g, fish); got != 2 {
		t.Errorf("the token has %d damage, want 2", got)
	}
}

// --- Tumbleweed Rising -------------------------------------------------------

func TestOtjTTumbleweedRisingMakesAnElementalAsBigAsYourBiggest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Creature(g, me.ID, "My Giant", "Creature — Giant", 4, 6)
	effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Their Titan", "Creature — Giant", 9, 9)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Tumbleweed Rising", "Sorcery", otjTTumbleweedOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	var made []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Elemental" && c.IsToken() && c.Controller == me.ID {
			made = append(made, c.InstanceID)
		}
	}
	if len(made) != 1 {
		t.Fatalf("%d Elementals, want 1", len(made))
	}
	if p, tough := effectstest.EffectivePTOf(t, g, made[0]); p != 4 || tough != 4 {
		t.Errorf("the Elemental is %d/%d, want 4/4", p, tough)
	}
	if c := effectstest.LayeredCard(t, g, made[0]); len(c.Colors) != 1 || c.Colors[0] != "G" {
		t.Errorf("the Elemental's colours are %v, want green", c.Colors)
	}
}
