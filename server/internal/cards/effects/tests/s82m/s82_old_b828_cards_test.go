package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b828_cards_test.go — Ice Age Auras, batch 828 (S82 older sets,
// #5636). Every test plays the real catalog entry through the engine.

const (
	oldB828Wings      = "06413d87-d119-4c04-93d5-5ced7ad4a858"
	oldB828Springs    = "123880ad-7b54-4ab0-82eb-c78b12ab89a7"
	oldB828Coop       = "19f7a0f2-aa3f-4253-be44-b08056918a26"
	oldB828Armor      = "1c482b20-b423-48ef-b162-53960d43b4f7"
	oldB828Grasp      = "1ebdc091-16d5-4803-821a-ac8c4696b5bc"
	oldB828Fetish     = "20c7630d-a97f-4f09-9ec5-b7a1739107de"
	oldB828BlackScrb  = "2774d8b3-dc5f-4281-b4c5-cb003fef0ab2"
	oldB828Shield     = "37d9c065-598d-4ee3-abfc-625bd6169ff1"
	oldB828RedScrb    = "380b6a48-c364-4939-be0b-df542c63a5ff"
	oldB828Caribou    = "39229fbc-1296-4ac7-9f79-65c6f0d09991"
	oldB828Snowblind  = "3f8bfa3d-7490-4259-a403-d13585ed2f66"
	oldB828BearType   = "Creature — Bear"
	oldB828SnowForest = "Basic Snow Land — Forest"
)

func TestOldB828EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB828Wings: "Wings of Aesthir", oldB828Springs: "Hot Springs", oldB828Coop: "Cooperation",
		oldB828Armor: "Armor of Faith", oldB828Grasp: "Binding Grasp", oldB828Fetish: "Krovikan Fetish",
		oldB828BlackScrb: "Black Scarab", oldB828Shield: "Spectral Shield", oldB828RedScrb: "Red Scarab",
		oldB828Caribou: "Caribou Range", oldB828Snowblind: "Snowblind",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB828Bear(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.B12Creature(g, owner, "Bear", oldB828BearType, 2, 2)
}

func oldB828Land(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.B12Permanent(g, owner, "Forest", "Basic Land — Forest")
}

func oldB828StrictCast(t *testing.T, g *game.Game, name, oracle string, host uuid.UUID) error {
	t.Helper()
	return effectstest.CastCatalogSpellErrStrict(t, g, name, effectstest.AuraTypeLine, oracle, effectstest.CardTarget(host))
}

// --- Wings of Aesthir --------------------------------------------------------

func TestOldB828WingsOfAesthirPumpsAndGrantsFlyingAndFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := oldB828Bear(g, me.ID)
	effectstest.AuraCast(t, g, "Wings of Aesthir", oldB828Wings, bear)
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("enchanted Bear = %d/%d, want 3/2", p, tough)
	}
	for _, kw := range []string{"flying", "first strike"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, bear, kw) {
			t.Errorf("enchanted Bear lacks %s", kw)
		}
	}
}

func TestOldB828WingsOfAesthirCannotEnchantALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	land := oldB828Land(g, g.Seats[0].ID)
	if err := oldB828StrictCast(t, g, "Wings of Aesthir", oldB828Wings, land); err == nil {
		t.Error("Wings of Aesthir was cast on a land")
	}
}

// --- Hot Springs -------------------------------------------------------------

func TestOldB828HotSpringsGivesTheLandAPreventionShield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := oldB828Land(g, me.ID)
	bear := oldB828Bear(g, me.ID)
	effectstest.AdvanceToMain(t, g)
	effectstest.AuraCast(t, g, "Hot Springs", oldB828Springs, land)
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DealDamageTo(t, g, uuid.Nil, bear, 1)
	if got := effectstest.ApaLive(g, bear).DamageMarked; got != 0 {
		t.Errorf("Bear has %d damage marked, want 0 (1 of 1 prevented)", got)
	}
	effectstest.DealDamageTo(t, g, uuid.Nil, bear, 1)
	if got := effectstest.ApaLive(g, bear).DamageMarked; got != 1 {
		t.Errorf("a second damage: Bear has %d marked, want 1", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Error("the tapped land activated the granted ability again")
	}
}

func TestOldB828HotSpringsCannotEnchantAnOpponentsLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	land := oldB828Land(g, g.Seats[1].ID)
	if err := oldB828StrictCast(t, g, "Hot Springs", oldB828Springs, land); err == nil {
		t.Error("Hot Springs was cast on an opponent's land")
	}
}

// --- Cooperation -------------------------------------------------------------

func TestOldB828CooperationGivesBanding(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := oldB828Bear(g, g.Seats[0].ID)
	if effectstest.HasKeywordOnBattlefield(t, g, bear, game.KeywordBanding) {
		t.Fatal("the Bear had banding before the Aura")
	}
	effectstest.AuraCast(t, g, "Cooperation", oldB828Coop, bear)
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, game.KeywordBanding) {
		t.Error("the enchanted Bear does not have banding")
	}
}

func TestOldB828CooperationIsRefusedWithoutItsMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := oldB828Bear(g, me.ID)
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Cooperation", TypeLine: effectstest.AuraTypeLine, OracleID: oldB828Coop, ManaCost: "{2}{W}", Colors: []string{"W"}, Owner: me.ID, Controller: me.ID})
	effectstest.FloatForTest(g, me, "WW")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardTarget(bear), Strict: true}); err == nil {
		t.Error("Cooperation was cast with two mana")
	}
}

// --- Armor of Faith ----------------------------------------------------------

func TestOldB828ArmorOfFaithPumpsAndAddsToughnessForW(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := oldB828Bear(g, me.ID)
	effectstest.AdvanceToMain(t, g)
	aura := effectstest.AuraCast(t, g, "Armor of Faith", oldB828Armor, bear)
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 3 {
		t.Fatalf("enchanted Bear = %d/%d, want 3/3", p, tough)
	}
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the ability was activated without {W}")
	}
	effectstest.DcMana(t, g, me.ID, "{W}")
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 4 {
		t.Errorf("after {W} the Bear = %d/%d, want 3/4", p, tough)
	}
}

// --- Binding Grasp -----------------------------------------------------------

func TestOldB828BindingGraspStealsAndIsSacrificedUnlessPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := oldB828Bear(g, opp.ID)
	aura := effectstest.AuraCast(t, g, "Binding Grasp", oldB828Grasp, theirs)
	if got := effectstest.ControllerOf(t, g, theirs); got != me.ID {
		t.Fatalf("controller %s, want the Aura's controller", got)
	}
	if p, tough := effectstest.PtOf(t, g, theirs); p != 2 || tough != 3 {
		t.Errorf("enchanted Bear = %d/%d, want 2/3", p, tough)
	}
	effectstest.WalkToUpkeepOfSeat0(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	if g.Battlefield.Contains(aura) || !me.Graveyard.Contains(aura) {
		t.Error("declining {1}{U} did not sacrifice Binding Grasp")
	}
	if got := effectstest.ControllerOf(t, g, theirs); got != opp.ID {
		t.Errorf("the Bear is controlled by %s after the Aura left, want its owner", got)
	}
}

func TestOldB828BindingGraspStaysWhenPaidFor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := oldB828Bear(g, opp.ID)
	aura := effectstest.AuraCast(t, g, "Binding Grasp", oldB828Grasp, theirs)
	effectstest.WalkToUpkeepOfSeat0(t, g)
	effectstest.FloatForTest(g, me, "CU")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	if !g.Battlefield.Contains(aura) {
		t.Error("paying {1}{U} did not keep Binding Grasp")
	}
	if got := effectstest.ControllerOf(t, g, theirs); got != me.ID {
		t.Error("control did not stay with the Aura's controller")
	}
}

// --- Krovikan Fetish ---------------------------------------------------------

func TestOldB828KrovikanFetishPumpsAndDrawsAtTheNextUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := oldB828Bear(g, me.ID)
	effectstest.AuraCast(t, g, "Krovikan Fetish", oldB828Fetish, bear)
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("enchanted Bear = %d/%d, want 3/3", p, tough)
	}
	before := me.Hand.Size()
	effectstest.AdvanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if me.Hand.Size() != before {
		t.Fatalf("drew before the upkeep trigger resolved: hand %d -> %d", before, me.Hand.Size())
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before+1 {
		t.Errorf("hand %d -> %d at the next upkeep, want +1", before, me.Hand.Size())
	}
}

func TestOldB828KrovikanFetishCannotEnchantALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	land := oldB828Land(g, g.Seats[0].ID)
	if err := oldB828StrictCast(t, g, "Krovikan Fetish", oldB828Fetish, land); err == nil {
		t.Error("Krovikan Fetish was cast on a land")
	}
}

// --- Black Scarab and Red Scarab ---------------------------------------------

func oldB828ScarabChecks(t *testing.T, name, oracle, color string) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	attacker := oldB828Bear(g, me.ID)
	effectstest.AuraCast(t, g, name, oracle, attacker)
	if p, tough := effectstest.PtOf(t, g, attacker); p != 2 || tough != 2 {
		t.Fatalf("with no %s permanent on the other side the Bear = %d/%d, want 2/2", color, p, tough)
	}
	// Your own permanent of that colour does not count.
	effectstest.SeedColoredCreature(g, me.ID, "My "+color, []string{color}, 0)
	if p, tough := effectstest.PtOf(t, g, attacker); p != 2 || tough != 2 {
		t.Fatalf("with only your own %s permanent the Bear = %d/%d, want 2/2", color, p, tough)
	}
	other := effectstest.SeedColoredCreature(g, them.ID, "Other", []string{"G"}, 0)
	if p, tough := effectstest.PtOf(t, g, attacker); p != 2 || tough != 2 {
		t.Fatalf("with an off-colour opposing permanent the Bear = %d/%d, want 2/2", p, tough)
	}
	blocker := effectstest.SeedColoredCreature(g, them.ID, "Their "+color, []string{color}, 0)
	if p, tough := effectstest.PtOf(t, g, attacker); p != 4 || tough != 4 {
		t.Fatalf("with an opposing %s permanent the Bear = %d/%d, want 4/4", color, p, tough)
	}
	effectstest.AttackIntoBlocks(t, g, them.ID, attacker)
	if effectstest.BlockRefusal(t, g, blocker, attacker) == "" {
		t.Errorf("a %s creature blocked the enchanted creature", color)
	}
	if got := effectstest.BlockRefusal(t, g, other, attacker); got != "" {
		t.Errorf("a green creature was refused: %q", got)
	}
}

func TestOldB828BlackScarab(t *testing.T) {
	oldB828ScarabChecks(t, "Black Scarab", oldB828BlackScrb, "B")
}

func TestOldB828RedScarab(t *testing.T) {
	oldB828ScarabChecks(t, "Red Scarab", oldB828RedScrb, "R")
}

func TestOldB828ScarabsCannotEnchantALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	land := oldB828Land(g, g.Seats[0].ID)
	if err := oldB828StrictCast(t, g, "Black Scarab", oldB828BlackScrb, land); err == nil {
		t.Error("Black Scarab was cast on a land")
	}
}

// --- Spectral Shield ---------------------------------------------------------

func TestOldB828SpectralShieldPumpsAndStopsSpellTargeting(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := oldB828Bear(g, me.ID)
	effectstest.AuraCast(t, g, "Spectral Shield", oldB828Shield, bear)
	if p, tough := effectstest.PtOf(t, g, bear); p != 2 || tough != 4 {
		t.Errorf("enchanted Bear = %d/%d, want 2/4", p, tough)
	}
	if err := oldB828StrictCast(t, g, "Wings of Aesthir", oldB828Wings, bear); err == nil {
		t.Error("a spell targeted the shielded creature")
	}
	other := oldB828Bear(g, me.ID)
	if err := oldB828StrictCast(t, g, "Wings of Aesthir", oldB828Wings, other); err != nil {
		t.Errorf("an unshielded creature was refused: %v", err)
	}
}

// --- Caribou Range -----------------------------------------------------------

func oldB828Caribous(g *game.Game, owner uuid.UUID) int {
	return effectstest.CountBattlefieldNamed(g, owner, "Caribou")
}

func TestOldB828CaribouRangeMakesCaribousAndSacrificesThemForLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := oldB828Land(g, me.ID)
	effectstest.AdvanceToMain(t, g)
	aura := effectstest.AuraCast(t, g, "Caribou Range", oldB828Caribou, land)
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("a Caribou was sacrificed with none in play")
	}
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the granted ability was activated without {W}{W}")
	}
	effectstest.DcMana(t, g, me.ID, "{W}{W}")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB828Caribous(g, me.ID); got != 1 {
		t.Fatalf("%d Caribou tokens, want 1", got)
	}
	var caribou game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Caribou" {
			caribou = c
		}
	}
	if caribou.Power != 0 || caribou.Toughness != 1 || len(caribou.Colors) != 1 || caribou.Colors[0] != "W" {
		t.Errorf("Caribou is %d/%d %v, want 0/1 white", caribou.Power, caribou.Toughness, caribou.Colors)
	}
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{caribou.InstanceID}}); err != nil {
		t.Fatalf("sacrifice the Caribou: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d -> %d, want +1", life, me.Life)
	}
	if got := oldB828Caribous(g, me.ID); got != 0 {
		t.Errorf("%d Caribou tokens after the sacrifice, want 0", got)
	}
}

func TestOldB828CaribouRangeSacrificesOnlyCaribouTokens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := oldB828Land(g, me.ID)
	effectstest.AdvanceToMain(t, g)
	aura := effectstest.AuraCast(t, g, "Caribou Range", oldB828Caribou, land)
	bear := oldB828Bear(g, me.ID)
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Error("a Bear was sacrificed as a Caribou token")
	}
}

// --- Snowblind ---------------------------------------------------------------

func TestOldB828SnowblindCountsItsControllersSnowLandsWhenNotAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := effectstest.B12Creature(g, opp.ID, "Big", oldB828BearType, 5, 5)
	effectstest.B12Permanent(g, opp.ID, "Snow-Covered Forest", oldB828SnowForest)
	effectstest.B12Permanent(g, opp.ID, "Snow-Covered Forest", oldB828SnowForest)
	effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	effectstest.B12Permanent(g, me.ID, "Snow-Covered Forest", oldB828SnowForest)
	effectstest.AuraCast(t, g, "Snowblind", oldB828Snowblind, big)
	if p, tough := effectstest.PtOf(t, g, big); p != 3 || tough != 3 {
		t.Errorf("enchanted creature = %d/%d, want 3/3 (two snow lands its controller controls)", p, tough)
	}
}

func TestOldB828SnowblindToughnessLossLeavesOneToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := effectstest.B12Creature(g, opp.ID, "Small", oldB828BearType, 2, 2)
	for i := 0; i < 4; i++ {
		effectstest.B12Permanent(g, opp.ID, "Snow-Covered Forest", oldB828SnowForest)
	}
	effectstest.B12Permanent(g, me.ID, "Snow-Covered Forest", oldB828SnowForest)
	effectstest.AuraCast(t, g, "Snowblind", oldB828Snowblind, small)
	if p, tough := effectstest.PtOf(t, g, small); p != -2 || tough != 1 {
		t.Errorf("enchanted creature = %d/%d, want -2/1 (X is 4, Y is 1)", p, tough)
	}
}

func TestOldB828SnowblindCountsTheDefendingPlayersSnowLandsWhileAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	attacker := effectstest.B12Creature(g, me.ID, "Big", oldB828BearType, 5, 5)
	effectstest.B12Permanent(g, me.ID, "Snow-Covered Forest", oldB828SnowForest)
	effectstest.B12Permanent(g, opp.ID, "Snow-Covered Forest", oldB828SnowForest)
	effectstest.B12Permanent(g, opp.ID, "Snow-Covered Forest", oldB828SnowForest)
	effectstest.AuraCast(t, g, "Snowblind", oldB828Snowblind, attacker)
	if p, tough := effectstest.PtOf(t, g, attacker); p != 4 || tough != 4 {
		t.Fatalf("not attacking the creature = %d/%d, want 4/4 (one snow land its controller controls)", p, tough)
	}
	effectstest.Pr7bAttack(t, g, opp.ID, attacker)
	if p, tough := effectstest.PtOf(t, g, attacker); p != 3 || tough != 3 {
		t.Errorf("attacking the creature = %d/%d, want 3/3 (two snow lands the defending player controls)", p, tough)
	}
}

func TestOldB828SnowblindCannotEnchantALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	land := oldB828Land(g, g.Seats[0].ID)
	if err := oldB828StrictCast(t, g, "Snowblind", oldB828Snowblind, land); err == nil {
		t.Error("Snowblind was cast on a land")
	}
}
