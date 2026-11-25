package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 275 (Eldritch Moon, delirium): Extricator of Sin, Dusk
// Feaster, Geist of the Lonely Vigil, Strange Augmentation, Mournwillow,
// Crop Sigil, Scour the Laboratory, Whispers of Emrakul.

const (
	oldB275Extricator = "08b3328c-1d96-4a05-ae8b-f1b654084faa"
	oldB275Feaster    = "3d56d82b-03a6-41cf-9833-c4efd8be6ea7"
	oldB275Geist      = "3e80481e-06ed-425a-8991-0557e49aba6a"
	oldB275Augment    = "5a499546-eab4-40e7-8f7b-4d318aa25810"
	oldB275Mournwill  = "742f6478-7c9e-43a4-81c1-09383fe65dad"
	oldB275Sigil      = "76ee013d-5293-4432-b425-649a2e82bd74"
	oldB275Scour      = "983c5c59-806c-4430-adec-c91a099c8d41"
	oldB275Whispers   = "cf8bc2e3-2f22-45d3-9810-bc039202f41c"
)

func TestOldB275CardsAreFullWithNoCaveats(t *testing.T) {
	for _, oracle := range []string{oldB275Extricator, oldB275Extricator + "#1", oldB275Feaster, oldB275Geist, oldB275Augment,
		oldB275Mournwill, oldB275Sigil, oldB275Scour, oldB275Whispers} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %v with caveats %v, want full with none", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// oldB275Delirium fills the player's graveyard with a creature, land,
// instant and sorcery card (four types).
func oldB275Delirium(g *game.Game, p *game.Player) {
	effectstest.GraveyardCardOf(g, p, "Old Bear", "Creature — Bear", "{1}{G}", "")
	effectstest.GraveyardCardOf(g, p, "Old Forest", "Basic Land — Forest", "", "")
	effectstest.GraveyardCardOf(g, p, "Old Bolt", "Instant", "{R}", "")
	effectstest.GraveyardCardOf(g, p, "Old Ritual", "Sorcery", "{1}{B}", "")
	g.BumpLayerVersionForTest() // the cards were pushed without a zone-move event
}

func oldB275Price(t *testing.T, g *game.Game, owner uuid.UUID, name, typeLine, cost, oracle string) int {
	t.Helper()
	card := game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle, Owner: owner, Controller: owner}
	base, err := game.ParseCost(cost)
	if err != nil {
		t.Fatal(err)
	}
	out, err := g.ApplyCostModifiers(base, game.CostQuery{Card: card, Controller: owner, FromZone: game.ZoneHand})
	if err != nil {
		t.Fatal(err)
	}
	return out.ManaValue()
}

func TestOldB275DuskFeasterAndScourCostTwoLessWithDelirium(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := oldB275Price(t, g, me.ID, "Scour the Laboratory", "Instant", "{4}{U}{U}", oldB275Scour)
	beforeF := oldB275Price(t, g, me.ID, "Dusk Feaster", "Creature — Vampire", "{5}{B}{B}", oldB275Feaster)
	effectstest.GraveyardCardOf(g, me, "Old Bear", "Creature — Bear", "{1}{G}", "")
	effectstest.GraveyardCardOf(g, me, "Old Forest", "Basic Land — Forest", "", "")
	effectstest.GraveyardCardOf(g, me, "Old Bolt", "Instant", "{R}", "")
	if got := oldB275Price(t, g, me.ID, "Scour the Laboratory", "Instant", "{4}{U}{U}", oldB275Scour); got != before {
		t.Errorf("three card types: Scour costs %d, want %d", got, before)
	}
	effectstest.GraveyardCardOf(g, me, "Old Ritual", "Sorcery", "{1}{B}", "")
	if got := oldB275Price(t, g, me.ID, "Scour the Laboratory", "Instant", "{4}{U}{U}", oldB275Scour); got != before-2 {
		t.Errorf("delirium: Scour costs %d, want %d", got, before-2)
	}
	if got := oldB275Price(t, g, me.ID, "Dusk Feaster", "Creature — Vampire", "{5}{B}{B}", oldB275Feaster); got != beforeF-2 {
		t.Errorf("delirium: Dusk Feaster costs %d, want %d", got, beforeF-2)
	}
}

func TestOldB275ScourDrawsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Scour the Laboratory", "Instant", oldB275Scour, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("hand %d → %d, want three cards drawn", hand, got)
	}
}

func TestOldB275GeistAttacksOnlyWithDelirium(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	geist := effectstest.B12Push(g, me.ID, "Geist of the Lonely Vigil", "Creature — Spirit Cleric", oldB275Geist, 2, 3)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(geist, opp.ID); err == nil {
		t.Fatal("a defender attacked without delirium")
	}
	oldB275Delirium(g, me)
	if err := g.DeclareAttacker(geist, opp.ID); err != nil {
		t.Fatalf("with delirium the Geist could not attack: %v", err)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, geist, "defender") {
		t.Error("the waiver removed defender")
	}
}

func TestOldB275StrangeAugmentationGivesTwoMoreWithDelirium(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Strange Augmentation", "Enchantment — Aura", oldB275Augment,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 3 || tough != 3 {
		t.Fatalf("without delirium the Bear is %d/%d, want 3/3", p, tough)
	}
	oldB275Delirium(g, me)
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("with delirium the Bear is %d/%d, want 5/5", p, tough)
	}
}

func TestOldB275WhispersDiscardsTwoAtRandomOnlyWithDelirium(t *testing.T) {
	for _, delirium := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		opp.Hand.Cards = nil
		for _, n := range []string{"A", "B", "C", "D"} {
			effectstest.PushCatalogHandCard(opp, n, "Instant", "")
		}
		if delirium {
			oldB275Delirium(g, me)
		}
		effectstest.AdvanceToMain(t, g)
		effectstest.CastCatalogSpell(t, g, "Whispers of Emrakul", "Sorcery", oldB275Whispers,
			[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
		effectstest.PassPriorityAroundTable(t, g)
		want := 3
		if delirium {
			want = 2
		}
		if got := opp.Hand.Size(); got != want {
			t.Errorf("delirium %v: opponent has %d cards, want %d", delirium, got, want)
		}
	}
}

func TestOldB275MournwillowStopsSmallCreaturesBlockingOnlyWithDelirium(t *testing.T) {
	for _, delirium := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		effectstest.ToMainForCost(t, g)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
		big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Bear", 3, 3)
		mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
		if delirium {
			oldB275Delirium(g, me)
		}
		effectstest.CastAndResolveCreature(t, g, "Mournwillow", "Creature — Plant Skeleton", oldB275Mournwill)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.AuraRestrictions(t, g, small)&game.CantBlock != 0; got != delirium {
			t.Errorf("delirium %v: power-2 creature can't block = %v", delirium, got)
		}
		if effectstest.AuraRestrictions(t, g, big)&game.CantBlock != 0 {
			t.Errorf("delirium %v: the power-3 creature was stopped from blocking", delirium)
		}
		if got := effectstest.AuraRestrictions(t, g, mine)&game.CantBlock != 0; got != delirium {
			t.Errorf("delirium %v: your own power-1 creature can't block = %v (any player's creatures)", delirium, got)
		}
	}
}

func TestOldB275CropSigilNeedsDeliriumToReturnACreatureAndALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sigil := effectstest.PushCatalogPermanent(g, me.ID, "Crop Sigil", "Enchantment", oldB275Sigil, false)
	bear := effectstest.GraveyardCardOf(g, me, "Old Bear", "Creature — Bear", "{1}{G}", "")
	forest := effectstest.GraveyardCardOf(g, me, "Old Forest", "Basic Land — Forest", "", "")
	effectstest.GraveyardCardOf(g, me, "Old Bolt", "Instant", "{R}", "")
	effectstest.FillPoolColored(me, "G", 3)
	refs := []game.TargetRef{{Kind: game.TargetCard, ID: bear}, {Kind: game.TargetCard, ID: forest}}
	if err := g.ActivateCatalogAbility(me.ID, sigil, 0, game.ActivateAbilityParams{Strict: true, Targets: refs}); err == nil {
		t.Fatal("activated with three card types in the graveyard")
	}
	effectstest.GraveyardCardOf(g, me, "Old Ritual", "Sorcery", "{1}{B}", "")
	if err := g.ActivateCatalogAbility(me.ID, sigil, 0, game.ActivateAbilityParams{Strict: true, Targets: refs}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) || !me.Hand.Contains(forest) {
		t.Error("the creature and land cards did not return to hand")
	}
	if g.Battlefield.Contains(sigil) {
		t.Error("the Sigil was not sacrificed")
	}
}

func TestOldB275CropSigilMayMillAtUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[1]
	effectstest.PushCatalogPermanent(g, owner.ID, "Crop Sigil", "Enchantment", oldB275Sigil, false)
	grave := owner.Graveyard.Size()
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AnswerLatestTriggerPrompt(t, g, owner.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := owner.Graveyard.Size(); got != grave+1 {
		t.Errorf("graveyard %d after the upkeep, want %d", got, grave+1)
	}
}

func TestOldB275ExtricatorMaySacrificeAnotherPermanentForAHorror(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	relic := effectstest.B12Permanent(g, me.ID, "Relic", "Artifact")
	effectstest.CastHoldCreature(t, g, "Extricator of Sin", "Creature — Human Cleric", oldB275Extricator, 0, 3)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if prompt == nil || prompt.ChooseMin != 0 || !effectstest.HasID(prompt.ChooseCards, relic) {
		t.Fatalf("want a may-sacrifice over another permanent: %+v", prompt)
	}
	effectstest.AnswerOwnPermanents(t, g, me.ID, relic)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) {
		t.Error("the artifact was not sacrificed")
	}
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Eldrazi Horror"); n != 1 {
		t.Errorf("%d Eldrazi Horror tokens, want 1", n)
	}
}

func TestOldB275ExtricatorMakesNothingIfYouDeclineTheSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	relic := effectstest.B12Permanent(g, me.ID, "Relic", "Artifact")
	effectstest.CastHoldCreature(t, g, "Extricator of Sin", "Creature — Human Cleric", oldB275Extricator, 0, 3)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(relic) {
		t.Error("declined, but the artifact was sacrificed")
	}
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Eldrazi Horror"); n != 0 {
		t.Errorf("%d Eldrazi Horror tokens, want 0", n)
	}
}

func TestOldB275ExtricatorTransformsAtUpkeepOnlyWithDelirium(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	card := effectstest.VowB15Put(t, g, effectstest.VowB15Row(oldB275Extricator, "Extricator of Sin", "Creature — Human Cleric", "{2}{W}", "0", "3",
		"Extricator of Flesh", "Creature — Eldrazi Horror", "3", "5", []string{"W"}), me)
	effectstest.BeginUpkeep(t, g, me.ID)
	if f := effectstest.IwFace(g, card); f != 0 {
		t.Fatalf("without delirium the Extricator is on face %d", f)
	}
	oldB275Delirium(g, me)
	effectstest.BeginUpkeep(t, g, me.ID)
	if f := effectstest.IwFace(g, card); f != 1 {
		t.Fatalf("with delirium the Extricator is on face %d, want 1", f)
	}
}

func TestOldB275ExtricatorOfFleshSacrificesANonEldraziForAHorror(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	card := effectstest.VowB15Put(t, g, effectstest.VowB15Row(oldB275Extricator, "Extricator of Sin", "Creature — Human Cleric", "{2}{W}", "0", "3",
		"Extricator of Flesh", "Creature — Eldrazi Horror", "3", "5", []string{"W"}), me)
	oldB275Delirium(g, me)
	effectstest.BeginUpkeep(t, g, me.ID)
	if f := effectstest.IwFace(g, card); f != 1 {
		t.Fatalf("not transformed: face %d", f)
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPool(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, card, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{card}}); err == nil {
		t.Fatal("sacrificed the Eldrazi itself")
	}
	if err := g.ActivateCatalogAbility(me.ID, card, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear was not sacrificed")
	}
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Eldrazi Horror"); n != 1 {
		t.Errorf("%d Eldrazi Horror tokens, want 1", n)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, card, "vigilance") {
		t.Error("an Eldrazi you control lacks vigilance")
	}
}
