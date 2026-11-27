package testsp

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pcy_b674_cards_test.go — Prophecy, S82 OLD batch 674 (#5617): Root Cage,
// Sheltering Prayers, Overburden, Brutal Suppression, Trenching Steed, Agent of
// Shauku, Fault Riders, Steal Strength, Silt Crawler, Gulf Squid and Keldon
// Firebombers, each through the real cast, activation and trigger paths.

const (
	pcy674RootCage    = "14c3a43c-cdb1-45d1-8eba-8e6d16bd7643"
	pcy674Prayers     = "44075eea-ab9d-4252-b77c-597d19a10fc4"
	pcy674Overburden  = "8104e454-77f5-4d38-894e-41ed0d74c17c"
	pcy674Suppression = "b4971aa3-32d1-4bc7-8133-ed67763d2dc6"
	pcy674Steed       = "31136103-535a-4d11-819a-1cbd57ad48a4"
	pcy674Agent       = "d2e0d739-b441-4173-a8f1-b831a19ea98e"
	pcy674Riders      = "df516046-7f83-4d64-aa0d-d687631d919a"
	pcy674Strength    = "e67434e6-5408-416a-be19-64291fca8e3e"
	pcy674Crawler     = "5ba1595a-6578-459d-96fe-fdfb84ac34d5"
	pcy674Squid       = "98a60a97-3a43-4544-8ea3-40d0bad79caf"
	pcy674Firebombers = "d5e02fa5-827e-414b-acb5-932de4bb3ee3"
)

func TestPcy674EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		pcy674RootCage: "Root Cage", pcy674Prayers: "Sheltering Prayers", pcy674Overburden: "Overburden",
		pcy674Suppression: "Brutal Suppression", pcy674Steed: "Trenching Steed", pcy674Agent: "Agent of Shauku",
		pcy674Riders: "Fault Riders", pcy674Strength: "Steal Strength", pcy674Crawler: "Silt Crawler",
		pcy674Squid: "Gulf Squid", pcy674Firebombers: "Keldon Firebombers",
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

func pcy674Land(g *game.Game, owner uuid.UUID, n int) []uuid.UUID {
	var ids []uuid.UUID
	for range n {
		ids = append(ids, effectstest.SeedLand(g, owner, "Plains", "Basic Land — Plains", ""))
	}
	return ids
}

// --- Root Cage ---

func TestPcy674RootCageKeepsMercenariesTappedForEveryController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Root Cage", "Enchantment", pcy674RootCage, false)
	mine := effectstest.B12Push(g, me.ID, "My Mercenary", "Creature — Human Mercenary", "", 1, 1)
	theirs := effectstest.B12Push(g, opp.ID, "Their Mercenary", "Creature — Human Mercenary", "", 1, 1)
	myBear := effectstest.B12Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	theirBear := effectstest.B12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	effectstest.TapForTest(t, g, mine, theirs, myBear, theirBear)
	effectstest.UpsAfterTheirUntap(t, g, opp.ID)
	if !effectstest.IsTapped(g, theirs) {
		t.Error("an opponent's Mercenary untapped during its controller's untap step")
	}
	if effectstest.IsTapped(g, theirBear) {
		t.Error("an opponent's non-Mercenary stayed tapped")
	}
	effectstest.UpsAfterTheirUntap(t, g, me.ID)
	if !effectstest.IsTapped(g, mine) {
		t.Error("your Mercenary untapped during your untap step")
	}
	if effectstest.IsTapped(g, myBear) {
		t.Error("your non-Mercenary stayed tapped")
	}
}

// --- Sheltering Prayers ---

func TestPcy674ShelteringPrayersGivesBasicLandsShroudWhileTheirControllerHasThreeOrFewer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Sheltering Prayers", "Enchantment", pcy674Prayers, false)
	mine := pcy674Land(g, me.ID, 3)
	theirs := pcy674Land(g, opp.ID, 2)
	for _, id := range append(append([]uuid.UUID{}, mine...), theirs...) {
		if !effectstest.HasEffectiveKeyword(t, g, id, "shroud") {
			t.Errorf("basic land %s lacks shroud with three or fewer lands", id)
		}
	}
	// A nonbasic land has none, and counts as a fourth land for its controller.
	nonbasic := effectstest.SeedLand(g, me.ID, "Wasteland", "Land", "")
	if effectstest.HasEffectiveKeyword(t, g, nonbasic, "shroud") {
		t.Error("a nonbasic land has shroud")
	}
	if effectstest.HasEffectiveKeyword(t, g, mine[0], "shroud") {
		t.Error("a basic land kept shroud while its controller controls four lands")
	}
	if !effectstest.HasEffectiveKeyword(t, g, theirs[0], "shroud") {
		t.Error("my land count switched off an opponent's shroud")
	}
	pcy674Land(g, opp.ID, 2)
	if effectstest.HasEffectiveKeyword(t, g, theirs[0], "shroud") {
		t.Error("an opponent's basic land kept shroud with four lands")
	}
}

// --- Overburden ---

func pcy674CastCreature(t *testing.T, g *game.Game, name string) uuid.UUID {
	t.Helper()
	return effectstest.TdCastFromHand(t, g, game.Card{Name: name, TypeLine: "Creature — Bear", Power: 2, Toughness: 2}, game.CastSpellParams{})
}

func TestPcy674OverburdenReturnsALandWhenAPlayerPutsANontokenCreatureOntoTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, g.Seats[1].ID, "Overburden", "Enchantment", pcy674Overburden, false)
	lands := pcy674Land(g, me.ID, 2)
	bear := pcy674CastCreature(t, g, "Bear")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPermanentPicks(t, g, me.ID, map[uuid.UUID][]uuid.UUID{me.ID: {lands[0]}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(lands[0]) || !me.Hand.Contains(lands[0]) {
		t.Error("the chosen land was not returned to its owner's hand")
	}
	if !g.Battlefield.Contains(lands[1]) || !g.Battlefield.Contains(bear) {
		t.Error("the other land or the creature left the battlefield")
	}
}

func TestPcy674OverburdenWithNoLandDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.PushCatalogPermanent(g, g.Seats[1].ID, "Overburden", "Enchantment", pcy674Overburden, false)
	bear := pcy674CastCreature(t, g, "Bear")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the creature left the battlefield")
	}
}

func TestPcy674OverburdenIgnoresNoncreaturesAndTokens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, g.Seats[1].ID, "Overburden", "Enchantment", pcy674Overburden, false)
	land := pcy674Land(g, me.ID, 1)[0]
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me.ID, effects.RuleDefinedToken(effects.TreasureToken()), 1); err != nil {
			t.Fatalf("token: %v", err)
		}
	})
	effectstest.TdCastFromHand(t, g, game.Card{Name: "Shrine", TypeLine: "Enchantment"}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(land) {
		t.Error("a land was returned for a noncreature or a token")
	}
}

// --- Brutal Suppression, Trenching Steed, Agent of Shauku ---

func TestPcy674BrutalSuppressionTaxesNontokenRebelsWithALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, g.Seats[1].ID, "Brutal Suppression", "Enchantment", pcy674Suppression, false)
	steed := effectstest.PushCatalogPermanent(g, me.ID, "Trenching Steed", "Creature — Horse Rebel", pcy674Steed, false)
	agent := effectstest.PushCatalogPermanent(g, me.ID, "Agent of Shauku", "Creature — Human Mercenary", pcy674Agent, false)
	lands := pcy674Land(g, me.ID, 3)
	if err := g.ActivateCatalogAbility(me.ID, steed, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: lands[:1]}); err == nil {
		t.Fatal("the Rebel's ability was activated with only one land sacrificed under Brutal Suppression")
	}
	if err := g.ActivateCatalogAbility(me.ID, steed, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: lands[:2]}); err != nil {
		t.Fatalf("the Rebel's ability with two lands: %v", err)
	}
	if g.Battlefield.Contains(lands[0]) || g.Battlefield.Contains(lands[1]) || !g.Battlefield.Contains(lands[2]) {
		t.Error("exactly the two named lands should have been sacrificed")
	}
	effectstest.DcMana(t, g, me.ID, "{B}{B}")
	if err := g.ActivateCatalogAbility(me.ID, agent, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: lands[2:], Targets: effectstest.CardTarget(steed)}); err != nil {
		t.Fatalf("a Mercenary's ability pays only its own land: %v", err)
	}
}

func TestPcy674TrenchingSteedGetsPlusZeroPlusThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	steed := effectstest.B12Push(g, me.ID, "Trenching Steed", "Creature — Horse Rebel", pcy674Steed, 2, 3)
	lands := pcy674Land(g, me.ID, 2)
	for i, land := range lands {
		if err := g.ActivateCatalogAbility(me.ID, steed, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{land}}); err != nil {
			t.Fatalf("activation %d: %v", i+1, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if p, tough := effectstest.PtOf(t, g, steed); p != 2 || tough != 3+3*(i+1) {
			t.Fatalf("after %d activations the Steed is %d/%d, want 2/%d", i+1, p, tough, 3+3*(i+1))
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, steed, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the ability was activated with no land to sacrifice")
	}
}

func TestPcy674AgentOfShaukuGivesPlusTwoPlusZeroToAnyCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	agent := effectstest.B12Push(g, me.ID, "Agent of Shauku", "Creature — Human Mercenary", pcy674Agent, 1, 1)
	theirs := effectstest.SeedCreature(g, "Their Bear", opp.ID)
	land := pcy674Land(g, me.ID, 1)[0]
	if err := g.ActivateCatalogAbility(me.ID, agent, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{land}, Targets: effectstest.CardTarget(theirs)}); err == nil {
		t.Fatal("activated with an empty mana pool")
	}
	effectstest.DcMana(t, g, me.ID, "{B}{B}")
	if err := g.ActivateCatalogAbility(me.ID, agent, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{land}, Targets: effectstest.CardTarget(theirs)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, theirs); p != 4 || tough != 2 {
		t.Errorf("the target is %d/%d, want 4/2", p, tough)
	}
	if g.Battlefield.Contains(land) {
		t.Error("the land was not sacrificed")
	}
}

// --- Fault Riders ---

func TestPcy674FaultRidersPumpsAndGainsFirstStrikeOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	riders := effectstest.B12Push(g, me.ID, "Fault Riders", "Creature — Human Soldier", pcy674Riders, 2, 2)
	lands := pcy674Land(g, me.ID, 2)
	if effectstest.HasEffectiveKeyword(t, g, riders, "first strike") {
		t.Fatal("first strike before the ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, riders, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: lands[:1]}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, riders); p != 4 || tough != 2 {
		t.Errorf("Fault Riders is %d/%d, want 4/2", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, riders, "first strike") {
		t.Error("no first strike after the ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, riders, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: lands[1:]}); err == nil {
		t.Error("the ability was activated a second time in one turn")
	}
}

func TestPcy674FaultRidersNeedsALandToSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	riders := effectstest.B12Push(g, me.ID, "Fault Riders", "Creature — Human Soldier", pcy674Riders, 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, riders, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with no land to sacrifice")
	}
}

// --- Steal Strength ---

func TestPcy674StealStrengthPumpsOneAndShrinksAnother(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	mine := effectstest.SeedCreature(g, "My Bear", me.ID)
	theirs := effectstest.SeedCreature(g, "Their Bear", opp.ID)
	effectstest.CastCatalogSpell(t, g, "Steal Strength", "Instant", pcy674Strength, effectstest.CardRefs(mine, theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("the first target is %d/%d, want 3/3", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, theirs); p != 1 || tough != 1 {
		t.Errorf("the second target is %d/%d, want 1/1", p, tough)
	}
}

func TestPcy674StealStrengthRefusesTheSameCreatureTwice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	bear := effectstest.SeedCreature(g, "My Bear", me.ID)
	if err := effectstest.CastCatalogSpellErr(t, g, "Steal Strength", "Instant", pcy674Strength, effectstest.CardRefs(bear, bear)); err == nil {
		t.Error("one creature was accepted as both targets")
	}
}

// --- Silt Crawler, Gulf Squid ---

func TestPcy674SiltCrawlerTapsYourLandsWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	mine := pcy674Land(g, me.ID, 2)
	theirs := pcy674Land(g, opp.ID, 1)[0]
	effectstest.TdCastFromHand(t, g, game.Card{Name: "Silt Crawler", TypeLine: "Creature — Beast", OracleID: pcy674Crawler, Power: 3, Toughness: 3}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range mine {
		if !effectstest.IsTapped(g, id) {
			t.Errorf("land %s is untapped", id)
		}
	}
	if effectstest.IsTapped(g, theirs) {
		t.Error("an opponent's land was tapped")
	}
}

func TestPcy674GulfSquidTapsTheTargetPlayersLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	mine := pcy674Land(g, me.ID, 1)[0]
	theirs := pcy674Land(g, opp.ID, 2)
	effectstest.TdCastFromHand(t, g, game.Card{Name: "Gulf Squid", TypeLine: "Creature — Squid Beast", OracleID: pcy674Squid, Power: 2, Toughness: 2}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTargetPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range theirs {
		if !effectstest.IsTapped(g, id) {
			t.Errorf("target player's land %s is untapped", id)
		}
	}
	if effectstest.IsTapped(g, mine) {
		t.Error("the controller's own land was tapped")
	}
}

// --- Keldon Firebombers ---

// pcy674Sacrifice answers a player's sacrifice prompts with the lands named,
// one at a time or all at once, whichever the engine asks.
func pcy674Sacrifice(t *testing.T, g *game.Game, who uuid.UUID, ids ...uuid.UUID) {
	t.Helper()
	for len(ids) > 0 {
		c := effectstest.SacrificeChoiceFor(g, who)
		if c == nil {
			t.Fatalf("no sacrifice prompt for the %d lands still to name", len(ids))
		}
		if c.Kind == game.PendingChoiceChooseCards {
			effectstest.AnswerSacrifice(t, g, who, ids...)
			return
		}
		effectstest.AnswerSacrifice(t, g, who, ids[0])
		ids = ids[1:]
	}
}

func TestPcy674KeldonFirebombersLeavesEachPlayerThreeLandsOfTheirChoice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[g.Turn.ActiveSeat], g.Seats[1], g.Seats[2]
	mine := pcy674Land(g, me.ID, 5)
	theirs := pcy674Land(g, opp.ID, 4)
	few := pcy674Land(g, third.ID, 2)
	effectstest.TdCastFromHand(t, g, game.Card{Name: "Keldon Firebombers", TypeLine: "Creature — Human Soldier", OracleID: pcy674Firebombers, Power: 3, Toughness: 3}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	pcy674Sacrifice(t, g, me.ID, mine[0], mine[1])
	pcy674Sacrifice(t, g, opp.ID, theirs[3])
	effectstest.PassPriorityAroundTable(t, g)
	for i, id := range mine {
		if g.Battlefield.Contains(id) != (i >= 2) {
			t.Errorf("my land %d on the battlefield = %v", i, g.Battlefield.Contains(id))
		}
	}
	for i, id := range theirs {
		if g.Battlefield.Contains(id) != (i < 3) {
			t.Errorf("their land %d on the battlefield = %v", i, g.Battlefield.Contains(id))
		}
	}
	for _, id := range few {
		if !g.Battlefield.Contains(id) {
			t.Error("a player with three or fewer lands sacrificed one")
		}
	}
}
