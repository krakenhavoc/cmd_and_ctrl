package testss82m

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b753_cards_test.go — Tempest, batch 753 (S82 older sets, #5628):
// activated abilities. Every test plays the real catalog entry through the
// engine.

const (
	oldB753Splicer = "4433a5fd-2daa-440a-b9cc-4ad028358cbc"
	oldB753Lurker  = "4b35b928-1088-457d-bd0d-3b14dc4ab230"
	oldB753Earth   = "50aa7aff-1f01-4224-9a83-01f74d703ec2"
	oldB753Knight  = "51262155-f1be-45a8-a21f-94e920a93e55"
	oldB753Nurture = "537a7af1-2b7e-4828-b645-7e34e86058d5"
	oldB753Harpy   = "58bbd316-e22c-4c23-8024-364122d1e9a2"
	oldB753Enrage  = "5ec9f317-887a-4e08-a165-e7cf15eb071d"
	oldB753Echo    = "75960649-93a4-4d56-b40f-f9fcbb4341d4"
	oldB753Crab    = "780f029f-e2e1-431b-9a8b-a1af51d7da77"
	oldB753Armodon = "7fc3c2a5-2754-4237-8a37-cb3166c20c3c"
)

func TestOldB753EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		oldB753Splicer, oldB753Lurker, oldB753Earth, oldB753Knight, oldB753Nurture,
		oldB753Harpy, oldB753Enrage, oldB753Echo, oldB753Crab, oldB753Armodon,
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

func oldB753Activate(g *game.Game, me *game.Player, id uuid.UUID, idx int, params game.ActivateAbilityParams) error {
	params.Strict = true
	return g.ActivateCatalogAbility(me.ID, id, idx, params)
}

func oldB753Float(g *game.Game, me *game.Player, mana string) {
	effectstest.FloatForTest(g, me, strings.TrimSpace(mana))
}

func oldB753Target(ids ...uuid.UUID) game.ActivateAbilityParams {
	var p game.ActivateAbilityParams
	for _, id := range ids {
		p.Targets = append(p.Targets, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	return p
}

func TestOldB753HarpyRegenerates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	harpy := effectstest.B12Push(g, me.ID, "Screeching Harpy", "Creature — Harpy Beast", oldB753Harpy, 2, 2)
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, harpy), "flying") {
		t.Error("the Harpy does not fly")
	}
	oldB753Float(g, me, "CB")
	if err := oldB753Activate(g, me, harpy, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.UsgB3Destroy(t, g, harpy)
	if !g.Battlefield.Contains(harpy) {
		t.Error("the regenerated Harpy was destroyed")
	}
}

func TestOldB753CrabGainsShroud(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	crab := effectstest.B12Push(g, me.ID, "Giant Crab", "Creature — Crab", oldB753Crab, 3, 3)
	if effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, crab), "shroud") {
		t.Fatal("the Crab starts with shroud")
	}
	oldB753Float(g, me, "U")
	if err := oldB753Activate(g, me, crab, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, crab), "shroud") {
		t.Error("the Crab did not gain shroud")
	}
	if err := oldB753Activate(g, me, crab, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("activated with no mana")
	}
}

func TestOldB753LurkerSacrificesASwampForFear(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	lurker := effectstest.B12Push(g, me.ID, "Marsh Lurker", "Creature — Beast", oldB753Lurker, 3, 2)
	forest := effectstest.SeedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
	if err := oldB753Activate(g, me, lurker, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{forest}}); err == nil {
		t.Fatal("the Lurker sacrificed a Forest")
	}
	swamp := effectstest.SeedLand(g, me.ID, "Swamp", "Basic Land — Swamp", "")
	if err := oldB753Activate(g, me, lurker, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{swamp}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(swamp) {
		t.Error("the Swamp was not sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, lurker), "fear") {
		t.Error("the Lurker did not gain fear")
	}
}

func TestOldB753EarthcraftTapsACreatureToUntapABasicLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	craft := effectstest.B12Push(g, me.ID, "Earthcraft", "Enchantment", oldB753Earth, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	basic := effectstest.SeedLand(g, opp.ID, "Island", "Basic Land — Island", "")
	nonbasic := effectstest.SeedLand(g, me.ID, "Cabal Coffers", "Land", "")
	effectstest.RtRetap(g, basic)
	effectstest.RtRetap(g, nonbasic)
	if err := oldB753Activate(g, me, craft, 0, oldB753Target(nonbasic)); err == nil {
		t.Fatal("Earthcraft untapped a nonbasic land")
	}
	p := oldB753Target(basic)
	p.TapIDs = []uuid.UUID{bear}
	if err := oldB753Activate(g, me, craft, 0, p); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !effectstest.B13Tapped(t, g, bear) {
		t.Error("the creature was not tapped to pay")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B13Tapped(t, g, basic) {
		t.Error("the basic land is still tapped")
	}
}

func TestOldB753KnightGainsProtectionFromAChosenColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	knight := effectstest.B12Push(g, me.ID, "Knight of Dawn", "Creature — Human Knight", oldB753Knight, 2, 2)
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, knight), "first strike") {
		t.Error("the Knight lacks first strike")
	}
	oldB753Float(g, me, "WW")
	if err := oldB753Activate(g, me, knight, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "R")
	if !effectstest.EffectiveAbilitiesContain(t, g, knight, "protection from red") {
		t.Errorf("the Knight lacks protection from red: %v", effectstest.EffectiveAbilities(t, g, knight))
	}
}

func TestOldB753ArmodonTargetMustBlockIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	armodon := effectstest.B12Push(g, me.ID, "Trumpeting Armodon", "Creature — Elephant", oldB753Armodon, 3, 3)
	foe := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	oldB753Float(g, me, "CG")
	if err := oldB753Activate(g, me, armodon, 0, oldB753Target(foe)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	recs := effectstest.BlocksAttackerRecords(g)
	if len(recs) != 1 || recs[0].Objects[0].ID != armodon {
		t.Fatalf("records = %+v, want one naming the Armodon", recs)
	}
}

func oldB753LicidTable(t *testing.T, oracle, name string) (g *game.Game, me *game.Player, licid, bear uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	licid = effectstest.B12Push(g, me.ID, name, "Creature — Licid", oracle, 1, 1)
	bear = effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	return
}

func TestOldB753NurturingLicidRegeneratesTheCreatureItEnchants(t *testing.T) {
	g, me, licid, bear := oldB753LicidTable(t, oldB753Nurture, "Nurturing Licid")
	oldB753Float(g, me, "G")
	if err := oldB753Activate(g, me, licid, 0, oldB753Target(bear)); err != nil {
		t.Fatalf("become an Aura: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	l := effectstest.B12Card(t, g, licid)
	if !l.IsAura() || l.IsCreature() || l.AttachedTo.ID != bear {
		t.Fatalf("not an Aura on the bear: aura %v creature %v attached %+v", l.IsAura(), l.IsCreature(), l.AttachedTo)
	}
	oldB753Float(g, me, "G")
	if err := oldB753Activate(g, me, licid, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("regenerate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.UsgB3Destroy(t, g, bear)
	if !g.Battlefield.Contains(bear) {
		t.Error("the enchanted bear was destroyed")
	}
}

func TestOldB753EnragingLicidGivesHasteToTheCreatureItEnchants(t *testing.T) {
	g, me, licid, bear := oldB753LicidTable(t, oldB753Enrage, "Enraging Licid")
	if effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, bear), "haste") {
		t.Fatal("the bear starts with haste")
	}
	oldB753Float(g, me, "R")
	if err := oldB753Activate(g, me, licid, 0, oldB753Target(bear)); err != nil {
		t.Fatalf("become an Aura: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, bear), "haste") {
		t.Error("the enchanted bear has no haste")
	}
	oldB753Float(g, me, "R")
	if err := g.PerformSpecialAction(me.ID, licid, game.SpecialActionEndEffect, game.SpecialActionParams{}); err != nil {
		t.Fatalf("end_effect: %v", err)
	}
	if effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, bear), "haste") {
		t.Error("the bear kept haste after the effect ended")
	}
	if l := effectstest.B12Card(t, g, licid); !l.IsCreature() || l.IsAura() {
		t.Error("the licid is not a creature again")
	}
}

func TestOldB753EchoChamberCopiesTheOpponentsChoiceForATurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	chamber := effectstest.B12Push(g, me.ID, "Echo Chamber", "Artifact", oldB753Echo, 0, 0)
	big := effectstest.B12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 5, 5)
	small := effectstest.B12Creature(g, opp.ID, "Small Bear", "Creature — Bear", 1, 1)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 3, 3)
	oldB753Float(g, me, "CCCC")
	if err := oldB753Activate(g, me, chamber, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	p := effectstest.OtPickFor(t, g, opp.ID, me.ID)
	if effectstest.HasID(p.PickTargetCards, mine) || !effectstest.HasID(p.PickTargetCards, big) || !effectstest.HasID(p.PickTargetCards, small) {
		t.Fatalf("offered %v, want the opponent's creatures only", p.PickTargetCards)
	}
	effectstest.OtAnswer(t, g, opp.ID, effectstest.OtCard(small))
	effectstest.PassPriorityAroundTable(t, g)
	toks := effectstest.Msc16TokensNamed(g, "Small Bear")
	if len(toks) != 1 {
		t.Fatalf("%d copies of the creature the opponent chose, want 1", len(toks))
	}
	token := &toks[0]
	if token.Controller != me.ID {
		t.Errorf("the token is controlled by %s, want the Chamber's controller", token.Controller)
	}
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, token.InstanceID), "haste") {
		t.Error("the token has no haste")
	}
	effectstest.AdvanceToStepInTurn(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(token.InstanceID) {
		t.Error("the token was not exiled at the end step")
	}
	if err := oldB753Activate(g, me, chamber, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("the Chamber activated again with no mana and tapped")
	}
}

func TestOldB753SplicerMovesAnAbilityBetweenCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	splicer := effectstest.B12Push(g, me.ID, "Phyrexian Splicer", "Artifact", oldB753Splicer, 0, 0)
	flier := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Flier", TypeLine: "Creature — Bird", Power: 2, Toughness: 2,
		Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID,
	})
	ground := effectstest.B12Creature(g, me.ID, "Ground Bear", "Creature — Bear", 2, 2)
	rows := map[string]int{"flying": 0, "first strike": 1, "trample": 2, "shadow": 3}
	oldB753Float(g, me, "CC")
	if err := oldB753Activate(g, me, splicer, rows["trample"], oldB753Target(flier, ground)); err == nil {
		t.Fatal("the trample row targeted a creature with no trample")
	}
	if err := oldB753Activate(g, me, splicer, rows["flying"], oldB753Target(flier, flier)); err == nil {
		t.Fatal("the same creature was both targets")
	}
	if err := oldB753Activate(g, me, splicer, rows["flying"], oldB753Target(flier, ground)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, flier), "flying") {
		t.Error("the flier kept flying")
	}
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, ground), "flying") {
		t.Error("the other creature did not gain flying")
	}
}
