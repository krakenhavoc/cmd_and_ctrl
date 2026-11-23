package testsds

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dsk_b1_cards_test.go — Duskmourn: House of Horror, batch 1 (#4466).
// Each test plays the real catalog entry and asserts what its text says.

const (
	dskB1CampgroundOracle     = "0eec9984-cd11-4a52-9234-469c6a5fb9aa"
	dskB1BleedingWoodsOracle  = "47b6d2ae-d3d7-41eb-9172-2076eb8d028d"
	dskB1AmalgamOracle        = "9fb95813-05e5-4aa0-bab9-53d8a071fe6a"
	dskB1ArabellaOracle       = "7c4bbb1b-29c4-4e06-aed0-b361293a585b"
	dskB1BearTrapOracle       = "f7afb2c3-50ee-40f5-8b14-32f80c2bacd0"
	dskB1SlasherOracle        = "b9671f4b-72d2-48d7-a497-2ea0ab72cd0b"
	dskB1CautiousOracle       = "28eb696f-a42c-4444-8139-dfd62029e020"
	dskB1ProwlerOracle        = "afe4a60c-747a-439b-bf61-cd323b8dffe7"
	dskB1CommuneOracle        = "0248bf01-14e0-4a39-ba3a-5177fe272fe9"
	dskB1CynicalLonerOracle   = "9d5ff568-1c55-4757-94d0-5743ca674bab"
	dskB1StartingLife         = 40
	dskB1CautiousSurvivorGain = 2
)

func TestDskB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		dskB1CampgroundOracle:    "Abandoned Campground",
		dskB1BleedingWoodsOracle: "Bleeding Woods",
		dskB1AmalgamOracle:       "Appendage Amalgam",
		dskB1ArabellaOracle:      "Arabella, Abandoned Doll",
		dskB1BearTrapOracle:      "Bear Trap",
		dskB1SlasherOracle:       "Cackling Slasher",
		dskB1CautiousOracle:      "Cautious Survivor",
		dskB1ProwlerOracle:       "Clammy Prowler",
		dskB1CommuneOracle:       "Commune with Evil",
		dskB1CynicalLonerOracle:  "Cynical Loner",
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

// dskB1CastAtDrawStep casts a card from the active seat's hand while the
// table is still in the draw step, where only an instant-speed spell may
// be cast.
func dskB1CastAtDrawStep(t *testing.T, g *game.Game, name, typeLine, oracle string) error {
	t.Helper()
	if g.Turn.Step != game.StepDraw {
		t.Fatalf("setup: at %s, want the draw step", g.Turn.Step)
	}
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.PushCatalogHandCard(me, name, typeLine, oracle)
	return g.CastSpell(me.ID, id, game.CastSpellParams{})
}

// --- Abandoned Campground, Bleeding Woods ----------------------------------------

func TestDskB1TaplandsEnterTappedUnlessAPlayerIsAtThirteen(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Abandoned Campground", dskB1CampgroundOracle},
		{"Bleeding Woods", dskB1BleedingWoodsOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			if !effectstest.B13Tapped(t, g, effectstest.B12PlayFromHand(t, g, tc.name, "Land", tc.oracle, game.CastSpellParams{})) {
				t.Error("everyone at 40: the land entered untapped")
			}
			g.Seats[1].Life = 13
			if effectstest.B13Tapped(t, g, effectstest.B12PlayFromHand(t, g, tc.name, "Land", tc.oracle, game.CastSpellParams{})) {
				t.Error("an opponent at 13: the land entered tapped")
			}
		})
	}
}

func TestDskB1TaplandsTapForEitherColour(t *testing.T) {
	for _, tc := range []struct{ name, oracle, produced string }{
		{"Abandoned Campground", dskB1CampgroundOracle, "{W|U}"},
		{"Bleeding Woods", dskB1BleedingWoodsOracle, "{R|G}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			land := effectstest.B12PlayFromHand(t, g, tc.name, "Land", tc.oracle, game.CastSpellParams{})
			if got := effectstest.LtProduced(t, g, land); !slices.Equal(got, []string{tc.produced}) {
				t.Errorf("the land adds %v, want %s", got, tc.produced)
			}
		})
	}
}

// --- Appendage Amalgam ------------------------------------------------------------

func TestDskB1AppendageAmalgamHasFlashAndSurveilsWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	if err := dskB1CastAtDrawStep(t, g, "Appendage Amalgam", "Enchantment Creature — Horror", dskB1AmalgamOracle); err != nil {
		t.Fatalf("Appendage Amalgam could not be cast at instant speed: %v", err)
	}

	g = effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	amalgam := effectstest.DskB1Creature(g, me.ID, "Appendage Amalgam", "Enchantment Creature — Horror", dskB1AmalgamOracle, 3, 2)
	top := effectstest.SeedLibrary(me, "Top Card")[0]
	effectstest.DeclareAttack(t, g, opp.ID, amalgam)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("attacking did not surveil")
	}
	if err := g.ResolveSurveil(c.ID, me.ID, []uuid.UUID{top}, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if !me.Graveyard.Contains(top) {
		t.Error("the surveilled card was not put into the graveyard")
	}
}

// --- Arabella, Abandoned Doll ---------------------------------------------------

func TestDskB1ArabellaDrainsByCreaturesWithPowerTwoOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arabella := effectstest.DskB1Creature(g, me.ID, "Arabella, Abandoned Doll", "Legendary Artifact Creature — Toy", dskB1ArabellaOracle, 1, 3)
	effectstest.B12Creature(g, me.ID, "Small Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Big Bear", "Creature — Bear", 3, 3)
	effectstest.B12Creature(g, opp.ID, "Their Squirrel", "Creature — Squirrel", 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, arabella)
	effectstest.PassPriorityAroundTable(t, g)

	// X is Arabella and the 2/2: two creatures with power 2 or less.
	if me.Life != dskB1StartingLife+2 {
		t.Errorf("my life %d, want %d", me.Life, dskB1StartingLife+2)
	}
	for _, p := range g.Seats[1:] {
		if p.Life != dskB1StartingLife-2 {
			t.Errorf("opponent %s life %d, want %d", p.Name, p.Life, dskB1StartingLife-2)
		}
	}
}

// --- Bear Trap --------------------------------------------------------------------

func TestDskB1BearTrapHasFlash(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	if err := dskB1CastAtDrawStep(t, g, "Bear Trap", "Artifact", dskB1BearTrapOracle); err != nil {
		t.Fatalf("Bear Trap could not be cast at instant speed: %v", err)
	}
	g = effectstest.NewCatalogGame(t)
	if err := dskB1CastAtDrawStep(t, g, "Plain Trinket", "Artifact", ""); err == nil {
		t.Fatal("control: an artifact without flash was cast in the draw step")
	}
}

func TestDskB1BearTrapTapsAndSacrificesToDealThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	trap := effectstest.PushCatalogPermanent(g, me.ID, "Bear Trap", "Artifact", dskB1BearTrapOracle, false)
	victim := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	if err := effectstest.HobB4Activate(t, g, me.ID, trap, 0, "{C}{C}", effectstest.CardRefs(victim)); err == nil {
		t.Fatal("activated for {2}; the cost is {3}")
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, trap, 0, "{C}", effectstest.CardRefs(victim)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(trap) {
		t.Error("Bear Trap was not sacrificed")
	}
	if g.Battlefield.Contains(victim) {
		t.Error("the 3/3 survived 3 damage")
	}
}

func TestDskB1BearTrapNeedsToUntap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	trap := effectstest.PushCatalogPermanent(g, me.ID, "Bear Trap", "Artifact", dskB1BearTrapOracle, false)
	victim := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	effectstest.DskB1Tap(g, trap)
	if err := effectstest.HobB4Activate(t, g, me.ID, trap, 0, "{C}{C}{C}", effectstest.CardRefs(victim)); err == nil {
		t.Fatal("a tapped Bear Trap paid its {T}")
	}
}

// --- Cackling Slasher ------------------------------------------------------------

func TestDskB1CacklingSlasherGetsACounterOnlyIfACreatureDied(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	slasher := effectstest.CastHoldCreature(t, g, "Cackling Slasher", "Creature — Human Assassin", dskB1SlasherOracle, 3, 3)
	if got := effectstest.PlusOnes(g, slasher); got != 0 {
		t.Errorf("nothing died: %d +1/+1 counters, want 0", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, slasher, "deathtouch") {
		t.Error("Cackling Slasher has no deathtouch")
	}

	g = effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.TrekB18BDestroy(t, g, victim)
	slasher = effectstest.CastHoldCreature(t, g, "Cackling Slasher", "Creature — Human Assassin", dskB1SlasherOracle, 3, 3)
	if got := effectstest.PlusOnes(g, slasher); got != 1 {
		t.Errorf("an opponent's creature died: %d +1/+1 counters, want 1", got)
	}
}

// --- Cautious Survivor -----------------------------------------------------------

func TestDskB1CautiousSurvivorGainsTwoLifeOnlyIfTapped(t *testing.T) {
	for _, tapped := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		survivor := effectstest.PushCatalogPermanent(g, me.ID, "Cautious Survivor", "Creature — Elf Survivor", dskB1CautiousOracle, false)
		if tapped {
			effectstest.DskB1Tap(g, survivor)
		}
		effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
		effectstest.PassPriorityAroundTable(t, g)
		want := dskB1StartingLife
		if tapped {
			want += dskB1CautiousSurvivorGain
		}
		if me.Life != want {
			t.Errorf("tapped=%v: life %d, want %d", tapped, me.Life, want)
		}
	}
}

// --- Clammy Prowler --------------------------------------------------------------

func TestDskB1ClammyProwlerMakesAnotherAttackerUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	prowler := effectstest.DskB1Creature(g, me.ID, "Clammy Prowler", "Enchantment Creature — Horror", dskB1ProwlerOracle, 2, 5)
	other := effectstest.B12Creature(g, me.ID, "Other Attacker", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "Stay Home", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, prowler, other)
	pick := effectstest.PickTargetPrompt(t, g)
	if slices.Contains(pick.PickTargetCards, prowler) {
		t.Error("the Prowler offered itself: the clause says another creature")
	}
	if slices.Contains(pick.PickTargetCards, idle) {
		t.Error("a creature that is not attacking was offered")
	}
	effectstest.AnswerPickTarget(t, g, other)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertRestrictions(t, g, other, game.CantBeBlocked)
	if r := effectstest.RestrictionsOf(t, g, prowler); r != 0 {
		t.Errorf("the Prowler itself got %v", r.Names())
	}
}

func TestDskB1ClammyProwlerAttackingAloneHasNoTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	prowler := effectstest.DskB1Creature(g, me.ID, "Clammy Prowler", "Enchantment Creature — Horror", dskB1ProwlerOracle, 2, 5)
	effectstest.DeclareAttack(t, g, opp.ID, prowler)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a target prompt was raised with no other attacking creature")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if r := effectstest.RestrictionsOf(t, g, prowler); r != 0 {
		t.Errorf("the Prowler made itself unblockable: %v", r.Names())
	}
}

// --- Commune with Evil -----------------------------------------------------------

func TestDskB1CommuneWithEvilTakesOneOfFourBinsTheRestThenGainsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SeedLibrary(me, "One", "Two", "Three", "Four", "Fifth")
	effectstest.CastCatalogSpell(t, g, "Commune with Evil", "Sorcery", dskB1CommuneOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick among the top four cards")
	}
	if len(pick.ChooseCards) != 4 || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Fatalf("offered %d cards in [%d,%d], want exactly one of the top four", len(pick.ChooseCards), pick.ChooseMin, pick.ChooseMax)
	}
	if slices.Contains(pick.ChooseCards, ids[4]) {
		t.Error("the fifth card was offered")
	}
	if me.Life != dskB1StartingLife {
		t.Errorf("life %d before the pick; the gain comes after the cards move", me.Life)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[2]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(ids[2]) {
		t.Error("the chosen card is not in hand")
	}
	for _, id := range []uuid.UUID{ids[0], ids[1], ids[3]} {
		if !me.Graveyard.Contains(id) {
			t.Errorf("a card that was not chosen is not in the graveyard")
		}
	}
	if !me.Library.Contains(ids[4]) {
		t.Error("the fifth card left the library")
	}
	if me.Life != dskB1StartingLife+3 {
		t.Errorf("life %d, want %d", me.Life, dskB1StartingLife+3)
	}
}

// --- Cynical Loner ---------------------------------------------------------------

func TestDskB1CynicalLonerCantBeBlockedByGlimmers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	loner := effectstest.DskB1Creature(g, me.ID, "Cynical Loner", "Creature — Human Survivor", dskB1CynicalLonerOracle, 3, 1)
	glimmer := effectstest.B12Creature(g, opp.ID, "Their Glimmer", "Enchantment Creature — Glimmer", 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(loner, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(glimmer, loner); err == nil {
		t.Error("a Glimmer blocked Cynical Loner")
	}
	if err := g.DeclareBlocker(bear, loner); err != nil {
		t.Errorf("a non-Glimmer could not block: %v", err)
	}
}

func TestDskB1CynicalLonerSurvivalMaySearchACardIntoTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Griselbrand", TypeLine: "Legendary Creature — Demon"},
	)
	loner := effectstest.PushCatalogPermanent(g, me.ID, "Cynical Loner", "Creature — Human Survivor", dskB1CynicalLonerOracle, false)
	effectstest.DskB1Tap(g, loner)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("a tapped Cynical Loner did not offer the search")
	}
	if c.SearchMin != 0 {
		t.Errorf("SearchMin = %d, want 0: the search is a \"you may\"", c.SearchMin)
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Griselbrand")
	if !effectstest.B02bGraveyardHasNamed(me, "Griselbrand") {
		t.Error("the chosen card did not go to the graveyard")
	}
}

func TestDskB1CynicalLonerUntappedDoesNotSearch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	effectstest.PushCatalogPermanent(g, me.ID, "Cynical Loner", "Creature — Human Survivor", dskB1CynicalLonerOracle, false)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Error("an untapped Cynical Loner offered the search")
	}
}
