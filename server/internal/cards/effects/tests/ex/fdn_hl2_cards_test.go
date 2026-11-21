package testsex

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fdn_hl2_cards_test.go — Foundations, H–L, second batch (#4465): the
// cards the first H–L batch left un-attempted that existing seams
// express. Each test plays the real catalog entry through the engine and
// asserts what its text says.

const (
	fdnHL2HalanaOracle      = "b347be2d-9aa1-42a9-b652-8fb032832f09"
	fdnHL2HeroicOracle      = "c442e8b0-a0a4-4839-8049-125b91b15c99"
	fdnHL2HordeOracle       = "b3c69bf4-880f-4198-a67c-87d23d9c3d41"
	fdnHL2ImmersturmOracle  = "fcb6c6d4-ea8d-49cf-b6df-377084bdb757"
	fdnHL2IncineratOracle   = "49f0da4a-e046-4220-b90e-e105a5f2ef00"
	fdnHL2IngeniousOracle   = "99853d2e-2223-4c05-bbd7-0a746f8f810d"
	fdnHL2InspirationOracle = "be87e410-4302-43ee-b880-56eedf2b3648"
	fdnHL2PaladinOracle     = "95efbb81-2418-42ed-ae34-5759efff8253"
	fdnHL2JazalOracle       = "609d13dc-c850-4877-8ad3-e84171b58069"
	fdnHL2JoragaOracle      = "216d3eed-cb07-40c5-869a-962e074e757d"
	fdnHL2KomaOracle        = "1ebf2142-f171-40cd-8143-15f50084b707"
	fdnHL2KykarOracle       = "0156c94b-e64c-49de-a46c-a203c5e05e98"
	fdnHL2LathrilOracle     = "9b568c34-ba0a-44a7-a0ee-bee3a37b3396"
	fdnHL2VanguardOracle    = "de1e1352-adb2-4ac1-bb2a-ee654d5c727d"
	fdnHL2LindenOracle      = "dbc56f3b-83b3-41c8-a8ae-d3774fe6ee01"
)

func TestFdnHL2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		fdnHL2HalanaOracle:      "Halana and Alena, Partners",
		fdnHL2HeroicOracle:      "Heroic Reinforcements",
		fdnHL2HordeOracle:       "Homunculus Horde",
		fdnHL2ImmersturmOracle:  "Immersturm Predator",
		fdnHL2IncineratOracle:   "Incinerating Blast",
		fdnHL2IngeniousOracle:   "Ingenious Leonin",
		fdnHL2InspirationOracle: "Inspiration from Beyond",
		fdnHL2PaladinOracle:     "Inspiring Paladin",
		fdnHL2JazalOracle:       "Jazal Goldmane",
		fdnHL2JoragaOracle:      "Joraga Invocation",
		fdnHL2KomaOracle:        "Koma, World-Eater",
		fdnHL2KykarOracle:       "Kykar, Zephyr Awakener",
		fdnHL2LathrilOracle:     "Lathril, Blade of the Elves",
		fdnHL2VanguardOracle:    "Leonin Vanguard",
		fdnHL2LindenOracle:      "Linden, the Steadfast Queen",
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

func TestFdnHL2KeywordBodiesCarryTheirPrintedKeywords(t *testing.T) {
	for _, c := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		keywords               []string
	}{
		{"Halana and Alena, Partners", "Legendary Creature — Human Ranger", fdnHL2HalanaOracle, 2, 3, []string{"reach", "first strike"}},
		{"Immersturm Predator", "Creature — Vampire Dragon", fdnHL2ImmersturmOracle, 3, 3, []string{"flying"}},
		{"Jazal Goldmane", "Legendary Creature — Cat Warrior", fdnHL2JazalOracle, 4, 4, []string{"first strike"}},
		{"Koma, World-Eater", "Legendary Creature — Serpent", fdnHL2KomaOracle, 8, 12, []string{"trample"}},
		{"Kykar, Zephyr Awakener", "Legendary Creature — Bird Wizard", fdnHL2KykarOracle, 3, 4, []string{"flying"}},
		{"Lathril, Blade of the Elves", "Legendary Creature — Elf Noble", fdnHL2LathrilOracle, 2, 3, []string{"menace"}},
		{"Linden, the Steadfast Queen", "Legendary Creature — Human Noble", fdnHL2LindenOracle, 3, 3, []string{"vigilance"}},
	} {
		g := effectstest.NewCatalogGame(t)
		id := effectstest.B12Push(g, g.Seats[0].ID, c.name, c.typeLine, c.oracle, c.power, c.toughness)
		for _, kw := range c.keywords {
			if !effectstest.HasKeywordOnBattlefield(t, g, id, kw) {
				t.Errorf("%s has no %s", c.name, kw)
			}
		}
	}
}

// fdnHL2White pushes a white creature for a test that reads colour.
func fdnHL2White(g *game.Game, owner uuid.UUID, name, oracle string, power, toughness int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Human", OracleID: oracle,
		Power: power, Toughness: toughness, Colors: []string{"W"}, Owner: owner, Controller: owner,
	})
}

// --- Halana and Alena, Partners ---------------------------------------------------

func TestFdnHL2HalanaAndAlenaGiveTheirPowerInCountersAndHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	halana := effectstest.B12Push(g, me.ID, "Halana and Alena, Partners", "Legendary Creature — Human Ranger", fdnHL2HalanaOracle, 2, 3)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	ch := effectstest.PickTargetPrompt(t, g)
	for _, bad := range []uuid.UUID{halana, theirs} {
		if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Fatal("Halana and Alena themselves, or an opponent's creature, was a legal target")
		}
	}
	effectstest.AnswerPickTarget(t, g, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, mine, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters, want 2 (their power)", got)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, mine, "haste") {
		t.Error("the creature did not gain haste")
	}
}

// --- Heroic Reinforcements --------------------------------------------------------

func TestFdnHL2HeroicReinforcementsMakesTwoSoldiersAndPumpsTheTeam(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Heroic Reinforcements", "Sorcery", fdnHL2HeroicOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.NamedOnBattlefield(g, "Soldier", me.ID); n != 2 {
		t.Fatalf("%d Soldier tokens, want 2", n)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, mine); p != 3 || tough != 3 || !effectstest.HasKeywordOnBattlefield(t, g, mine, "haste") {
		t.Errorf("your creature is %d/%d, haste %v; want 3/3 with haste", p, tough, effectstest.HasKeywordOnBattlefield(t, g, mine, "haste"))
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Soldier" {
			continue
		}
		if p, tough := effectstest.EffectivePTOf(t, g, c.InstanceID); p != 2 || tough != 2 || !effectstest.HasKeywordOnBattlefield(t, g, c.InstanceID, "haste") {
			t.Errorf("a Soldier is %d/%d; want a 2/2 with haste", p, tough)
		}
	}
	if p, _ := effectstest.EffectivePTOf(t, g, theirs); p != 2 || effectstest.HasKeywordOnBattlefield(t, g, theirs, "haste") {
		t.Error("an opponent's creature was pumped")
	}
}

// --- Homunculus Horde -------------------------------------------------------------

func TestFdnHL2HomunculusHordeCopiesItselfOnTheSecondDrawOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, other := g.Seats[1], g.Seats[2]
	effectstest.PushCatalogPermanent(g, me.ID, "Homunculus Horde", "Creature — Homunculus", fdnHL2HordeOracle, false)
	effectstest.HobA5Draw(t, g, other, nil)
	effectstest.HobA5Draw(t, g, other, nil)
	if n := effectstest.NamedOnBattlefield(g, "Homunculus Horde", me.ID); n != 1 {
		t.Fatalf("%d Hordes after another player drew twice, want 1", n)
	}
	effectstest.HobA5Draw(t, g, me, nil)
	if n := effectstest.NamedOnBattlefield(g, "Homunculus Horde", me.ID); n != 1 {
		t.Fatalf("%d Hordes after the first draw, want 1", n)
	}
	effectstest.HobA5Draw(t, g, me, nil)
	if n := effectstest.NamedOnBattlefield(g, "Homunculus Horde", me.ID); n != 2 {
		t.Fatalf("%d Hordes after the second draw, want 2", n)
	}
	effectstest.HobA5Draw(t, g, me, nil)
	if n := effectstest.NamedOnBattlefield(g, "Homunculus Horde", me.ID); n != 2 {
		t.Fatalf("%d Hordes after the third draw, want it to stay 2", n)
	}
}

// --- Immersturm Predator ----------------------------------------------------------

func TestFdnHL2ImmersturmPredatorSacrificesForIndestructibleAndEatsAGraveyardCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	pred := effectstest.B12Push(g, me.ID, "Immersturm Predator", "Creature — Vampire Dragon", fdnHL2ImmersturmOracle, 3, 3)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, pred, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{pred}}); err == nil {
		t.Fatal("the Predator sacrificed itself to its \"another creature\" cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, pred, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.IsTapped(g, pred) {
		t.Fatal("the Predator was not tapped")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, pred, "indestructible") {
		t.Error("the Predator did not gain indestructible")
	}
	// Tapping it is a "becomes tapped": the trigger asks for up to one
	// card from a graveyard, and the sacrificed creature is there.
	effectstest.AnswerPickTarget(t, g, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(fodder) {
		t.Error("the chosen graveyard card was not exiled")
	}
	if got := effectstest.CountersOn(g, pred, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
}

// --- Incinerating Blast -----------------------------------------------------------

func TestFdnHL2IncineratingBlastDealsSixThenMayRummage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Drawn", TypeLine: "Instant"})
	fodder := effectstest.HandCard(me, "Fodder", "Instant")
	big := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 6, 6)
	effectstest.CastCatalogSpell(t, g, "Incinerating Blast", "Sorcery", fdnHL2IncineratOracle, effectstest.CardRefs(big))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("no optional discard after the damage")
	}
	hand := me.Hand.Size()
	effectstest.AnswerDiscard(t, g, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("the 6/6 survived 6 damage")
	}
	if me.Hand.Contains(fodder) || me.Hand.Size() != hand {
		t.Errorf("hand %d → %d: one discarded and one drawn should leave it the same", hand, me.Hand.Size())
	}
}

func TestFdnHL2IncineratingBlastDrawsNothingWithoutADiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.HandCard(me, "Fodder", "Instant")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Incinerating Blast", "Sorcery", fdnHL2IncineratOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.AnswerDiscard(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d → %d after declining: no card is drawn without a discard", hand, me.Hand.Size())
	}
}

// --- Ingenious Leonin -------------------------------------------------------------

func TestFdnHL2IngeniousLeoninCountersAnAttackerAndGivesACatFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	leonin := effectstest.B12Push(g, me.ID, "Ingenious Leonin", "Creature — Cat Soldier", fdnHL2IngeniousOracle, 4, 4)
	cat := effectstest.B12Creature(g, me.ID, "My Cat", "Creature — Cat", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	home := effectstest.B12Creature(g, me.ID, "Stays Home", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, leonin, cat, bear)
	for _, bad := range []uuid.UUID{leonin, home} {
		effectstest.FillPoolColored(me, "W", 4)
		if err := g.ActivateCatalogAbility(me.ID, leonin, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bad)}); err == nil {
			t.Fatal("the Leonin itself, or a creature that isn't attacking, was a legal target")
		}
	}
	for _, id := range []uuid.UUID{cat, bear} {
		effectstest.FillPoolColored(me, "W", 4)
		if err := g.ActivateCatalogAbility(me.ID, leonin, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(id)}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if effectstest.CountersOn(g, cat, game.CounterPlusOne) != 1 || effectstest.CountersOn(g, bear, game.CounterPlusOne) != 1 {
		t.Error("each target should have one +1/+1 counter")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, cat, "first strike") {
		t.Error("the Cat did not gain first strike")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, bear, "first strike") {
		t.Error("a non-Cat gained first strike")
	}
}

// --- Inspiration from Beyond ------------------------------------------------------

func TestFdnHL2InspirationFromBeyondMillsThenReturnsAnInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Library.Cards = nil
	for i := 0; i < 3; i++ {
		effectstest.PushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Land"})
	}
	milled := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: milled, Name: "Milled Bolt", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	bear := effectstest.B17GraveyardCard(me, "Old Bear", "Creature — Bear", "{1}{G}")
	effectstest.CastCatalogSpell(t, g, "Inspiration from Beyond", "Sorcery", fdnHL2InspirationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no choose-cards prompt after the mill")
	}
	if !effectstest.HasID(pick.ChooseCards, milled) || effectstest.HasID(pick.ChooseCards, bear) {
		t.Errorf("offered %v, want the milled instant and not the creature", pick.ChooseCards)
	}
	if pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want exactly one", pick.ChooseMin, pick.ChooseMax)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, milled)
	if !me.Hand.Contains(milled) {
		t.Error("the instant was not returned to hand")
	}
}

func TestFdnHL2InspirationFromBeyondFlashesBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	top := effectstest.SeedLibrary(me, "A", "B", "C", "D")
	id := effectstest.SeedGraveyardCard(t, g, "Inspiration from Beyond", "Sorcery", fdnHL2InspirationOracle)
	effectstest.FillPoolColored(me, "U", 7)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no choose-cards prompt after the mill")
	}
	if effectstest.HasID(pick.ChooseCards, id) || !effectstest.HasID(pick.ChooseCards, top[0]) {
		t.Errorf("offered %v: want the milled sorceries, never the spell itself", pick.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, top[0])
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(top[0]) {
		t.Error("the milled sorcery was not returned")
	}
	if !g.Exile.Contains(id) {
		t.Error("the flashed-back sorcery was not exiled")
	}
}

// --- Inspiring Paladin ------------------------------------------------------------

func TestFdnHL2InspiringPaladinGrantsFirstStrikeOnlyDuringYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	paladin := effectstest.B12Push(g, me.ID, "Inspiring Paladin", "Creature — Human Knight", fdnHL2PaladinOracle, 3, 3)
	countered := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Countered", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Counters: map[string]int{game.CounterPlusOne: 1}, Owner: me.ID, Controller: me.ID,
	})
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	if !effectstest.HasKeywordOnBattlefield(t, g, paladin, "first strike") || !effectstest.HasKeywordOnBattlefield(t, g, countered, "first strike") {
		t.Fatal("during your turn the Paladin and a creature with a +1/+1 counter have first strike")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, plain, "first strike") {
		t.Fatal("a creature with no counter has first strike")
	}
	effectstest.AdvanceToUpkeepOf(t, g, (seat+1)%len(g.Seats))
	if effectstest.HasKeywordOnBattlefield(t, g, paladin, "first strike") || effectstest.HasKeywordOnBattlefield(t, g, countered, "first strike") {
		t.Error("first strike on an opponent's turn")
	}
}

// --- Jazal Goldmane ---------------------------------------------------------------

func TestFdnHL2JazalGoldmanePumpsAttackersByTheAttackerCount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	jazal := effectstest.B12Push(g, me.ID, "Jazal Goldmane", "Legendary Creature — Cat Warrior", fdnHL2JazalOracle, 4, 4)
	a := effectstest.B12Creature(g, me.ID, "Attacker A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Attacker B", "Creature — Bear", 2, 2)
	home := effectstest.B12Creature(g, me.ID, "Stays Home", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.FillPoolColored(me, "W", 5)
	if err := g.ActivateCatalogAbility(me.ID, jazal, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if p, tough := effectstest.EffectivePTOf(t, g, id); p != 4 || tough != 4 {
			t.Errorf("an attacker is %d/%d, want 4/4 (two attacking creatures)", p, tough)
		}
	}
	if p, _ := effectstest.EffectivePTOf(t, g, home); p != 2 {
		t.Error("a creature that isn't attacking was pumped")
	}
	if p, _ := effectstest.EffectivePTOf(t, g, jazal); p != 4 {
		t.Error("Jazal, not attacking, was pumped")
	}
}

// --- Joraga Invocation ------------------------------------------------------------

func TestFdnHL2JoragaInvocationPumpsYourTeamAndMakesItMustBeBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := []uuid.UUID{
		effectstest.B12Creature(g, me.ID, "Mine A", "Creature — Elf", 1, 1),
		effectstest.B12Creature(g, me.ID, "Mine B", "Creature — Elf", 2, 2),
	}
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Joraga Invocation", "Sorcery", fdnHL2JoragaOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i, id := range mine {
		if p, tough := effectstest.EffectivePTOf(t, g, id); p != i+4 || tough != i+4 {
			t.Errorf("your creature is %d/%d, want %d/%d", p, tough, i+4, i+4)
		}
		if len(effectstest.EffectiveOf(t, g, id).BlockRequirements) == 0 {
			t.Error("your creature carries no must-be-blocked requirement")
		}
	}
	if p, _ := effectstest.EffectivePTOf(t, g, theirs); p != 2 || len(effectstest.EffectiveOf(t, g, theirs).BlockRequirements) != 0 {
		t.Error("an opponent's creature was affected")
	}
}

// --- Koma, World-Eater ------------------------------------------------------------

func TestFdnHL2KomaCantBeCounteredHasWardAndMakesFourCoils(t *testing.T) {
	spec, _ := effects.Lookup(fdnHL2KomaOracle)
	if !spec.CantBeCountered {
		t.Error("Koma can be countered")
	}
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	koma := effectstest.B12Push(g, me.ID, "Koma, World-Eater", "Legendary Creature — Serpent", fdnHL2KomaOracle, 8, 12)
	effectstest.DealCombatDamageToPlayer(g, koma, opp.ID, 8)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.NamedOnBattlefield(g, "Koma's Coil", me.ID); n != 4 {
		t.Fatalf("%d Koma's Coils, want 4", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Koma's Coil" && (c.Power != 3 || c.Toughness != 3 || !c.HasSubtype("Serpent")) {
			t.Errorf("a Coil is %d/%d %s; want a 3/3 Serpent", c.Power, c.Toughness, c.TypeLine)
		}
	}
}

// --- Kykar, Zephyr Awakener -------------------------------------------------------

func TestFdnHL2KykarMakesASpiritForANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Kykar, Zephyr Awakener", "Legendary Creature — Bird Wizard", fdnHL2KykarOracle, 3, 4)
	effectstest.CastCatalogSpell(t, g, "Test Bear", "Creature — Bear", "", nil)
	effectstest.NoModePick(t, g, me.ID, "a creature spell does not trigger it")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Test Instant", "Instant", "", nil)
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no mode pick for a noncreature spell")
	}
	effectstest.AnswerMode(t, g, c, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.NamedOnBattlefield(g, "Spirit", me.ID); n != 1 {
		t.Errorf("%d Spirit tokens, want 1", n)
	}
}

func TestFdnHL2KykarFlickersAnotherCreatureYouControlAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Kykar, Zephyr Awakener", "Legendary Creature — Bird Wizard", fdnHL2KykarOracle, 3, 4)
	pal := effectstest.B12Creature(g, me.ID, "Pal", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Test Instant", "Instant", "", nil)
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no mode pick for a noncreature spell")
	}
	effectstest.AnswerMode(t, g, c, me.ID, 0)
	effectstest.PickCard(t, g, me.ID, pal)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(pal) || !g.Exile.Contains(pal) {
		t.Fatal("the creature was not exiled")
	}
	for g.Turn.Step != game.StepEnd {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep toward the end step: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Exile.Contains(pal) || effectstest.NamedOnBattlefield(g, "Pal", me.ID) != 1 {
		t.Error("the creature did not return at the next end step")
	}
}

// --- Lathril, Blade of the Elves --------------------------------------------------

func TestFdnHL2LathrilMakesThatManyElvesOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	lathril := effectstest.B12Push(g, me.ID, "Lathril, Blade of the Elves", "Legendary Creature — Elf Noble", fdnHL2LathrilOracle, 2, 3)
	effectstest.DealCombatDamageToPlayer(g, lathril, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.NamedOnBattlefield(g, "Elf Warrior", me.ID); n != 3 {
		t.Errorf("%d Elf Warrior tokens, want 3", n)
	}
}

func TestFdnHL2LathrilTapsTenElvesToDrainTen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	lathril := effectstest.B12Push(g, me.ID, "Lathril, Blade of the Elves", "Legendary Creature — Elf Noble", fdnHL2LathrilOracle, 2, 3)
	var elves []uuid.UUID
	for i := 0; i < 10; i++ {
		elves = append(elves, effectstest.B12Creature(g, me.ID, "Elf", "Creature — Elf Warrior", 1, 1))
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, lathril, 0, game.ActivateAbilityParams{TapIDs: append(append([]uuid.UUID(nil), elves[:9]...), bear)}); err == nil {
		t.Fatal("a non-Elf paid the cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, lathril, 0, game.ActivateAbilityParams{TapIDs: elves[:9]}); err == nil {
		t.Fatal("nine Elves paid the cost")
	}
	life := me.Life
	opps := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		if p.ID != me.ID {
			opps[p.ID] = p.Life
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, lathril, 0, game.ActivateAbilityParams{TapIDs: elves}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+10 {
		t.Errorf("life %d, want %d", me.Life, life+10)
	}
	for _, p := range g.Seats {
		if before, ok := opps[p.ID]; ok && p.Life != before-10 {
			t.Errorf("an opponent's life %d, want %d", p.Life, before-10)
		}
	}
}

// --- Leonin Vanguard --------------------------------------------------------------

func TestFdnHL2LeoninVanguardGrowsAndGainsWithThreeCreatures(t *testing.T) {
	for _, tc := range []struct {
		name   string
		others int
		gain   int
		power  int
	}{
		{"three creatures", 2, 1, 2},
		{"two creatures", 1, 0, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			vanguard := effectstest.B12Push(g, me.ID, "Leonin Vanguard", "Creature — Cat Soldier", fdnHL2VanguardOracle, 1, 1)
			for i := 0; i < tc.others; i++ {
				effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
			}
			life := me.Life
			effectstest.AdvanceTo(t, g, game.StepBeginCombat)
			effectstest.PassPriorityAroundTable(t, g)
			if me.Life != life+tc.gain {
				t.Errorf("life %d, want %d", me.Life, life+tc.gain)
			}
			if p, _ := effectstest.EffectivePTOf(t, g, vanguard); p != tc.power {
				t.Errorf("power %d, want %d", p, tc.power)
			}
		})
	}
}

// --- Linden, the Steadfast Queen --------------------------------------------------

func TestFdnHL2LindenGainsOneForEachWhiteAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	linden := fdnHL2White(g, me.ID, "Linden, the Steadfast Queen", fdnHL2LindenOracle, 3, 3)
	knight := fdnHL2White(g, me.ID, "White Knight", "", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Green Bear", "Creature — Bear", 2, 2)
	life := me.Life
	effectstest.DeclareAttack(t, g, opp.ID, linden, knight, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d, want %d (two white attackers)", me.Life, life+2)
	}
	if effectstest.IsTapped(g, linden) {
		t.Error("Linden tapped to attack despite vigilance")
	}
}
