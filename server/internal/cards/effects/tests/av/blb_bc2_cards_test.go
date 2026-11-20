package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// blb_bc2_cards_test.go — Bloomburrow cards starting with B and C, second
// batch (#4467). Each test plays the real catalog entry and asserts what
// its text says.

const (
	blbBC2BanditsOracle    = "bcbe879a-4609-4493-bd7d-47c26d27e83c"
	blbBC2BaylenOracle     = "73dc28a7-37ac-4b16-aa00-967c9c44d979"
	blbBC2BlacksmithOracle = "9fdad42a-9db5-4d91-9e80-9d5ce19038b8"
	blbBC2BuilderOracle    = "0b19082c-8f80-448b-9e1b-0f324a159409"
	blbBC2ByrkeOracle      = "e9c6e4e6-0c3f-4048-93e6-161a1221ea36"
	blbBC2CacheGrabOracle  = "06ccec61-27d2-4079-80b5-29ebd3a980ee"
	blbBC2ClementOracle    = "57ce5ebd-cfac-4bb9-b010-04d0c285be27"
	blbBC2HeistOracle      = "7a0bc265-6b40-48d6-a588-c9624df7d1bc"
)

func TestBlbBC2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		blbBC2BanditsOracle:    "Bandit's Talent",
		blbBC2BaylenOracle:     "Baylen, the Haymaker",
		blbBC2BlacksmithOracle: "Blacksmith's Talent",
		blbBC2BuilderOracle:    "Builder's Talent",
		blbBC2ByrkeOracle:      "Byrke, Long Ear of the Law",
		blbBC2CacheGrabOracle:  "Cache Grab",
		blbBC2ClementOracle:    "Clement, the Worrywort",
		blbBC2HeistOracle:      "Cruelclaw's Heist",
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

// --- Bandit's Talent ----------------------------------------------------------

func TestBlbBC2BanditsTalentEachOpponentDiscardsTwoOrANonlandCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := effectstest.BlbB1Seats(g)
	third := g.Seats[(g.Turn.ActiveSeat+3)%len(g.Seats)]
	hand := effectstest.RevealHand(opp, effectstest.RhForest(), effectstest.RhBolt(), effectstest.RhForest())
	lone := effectstest.RevealHand(other, effectstest.RhForest())
	effectstest.RevealHand(third)
	effectstest.CastCatalogSpell(t, g, "Bandit's Talent", "Enchantment — Class", blbBC2BanditsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	prompt := effectstest.DiscardChoiceFor(g, opp.ID)
	if prompt == nil {
		t.Fatal("the opponent with three cards was not asked to discard")
	}
	if err := g.ResolveChooseCards(prompt.ID, opp.ID, []uuid.UUID{hand[0]}); err == nil {
		t.Error("one land card was accepted; it must be two cards or one nonland card")
	}
	if err := g.ResolveChooseCards(prompt.ID, opp.ID, []uuid.UUID{hand[1]}); err != nil {
		t.Fatalf("one nonland card was refused: %v", err)
	}
	if opp.Hand.Size() != 2 || opp.Hand.Contains(hand[1]) {
		t.Errorf("the opponent's hand is %d cards, want the two Forests", opp.Hand.Size())
	}

	prompt = effectstest.DiscardChoiceFor(g, other.ID)
	if prompt == nil {
		t.Fatal("the opponent with one card was not asked to discard it")
	}
	if err := g.ResolveChooseCards(prompt.ID, other.ID, lone); err != nil {
		t.Fatalf("discard the last card: %v", err)
	}
	if other.Hand.Size() != 0 {
		t.Errorf("the opponent with one card still holds %d", other.Hand.Size())
	}
	if effectstest.DiscardChoiceFor(g, third.ID) != nil || effectstest.DiscardChoiceFor(g, me.ID) != nil {
		t.Error("an empty hand, or the controller, was asked to discard")
	}
}

func TestBlbBC2BanditsTalentLevelTwoDrainsAnOpponentWithOneCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := effectstest.BlbB1Seats(g)
	class := effectstest.PushClass(g, me.ID, "Bandit's Talent", "Enchantment — Class", blbBC2BanditsOracle)
	effectstest.ToMainForCost(t, g)
	if err := effectstest.LevelUpTo(t, g, me.ID, class, 0, "{B}"); err != nil {
		t.Fatalf("level 2: %v", err)
	}
	if got := effectstest.LevelOf(g, class); got != 2 {
		t.Fatalf("level %d, want 2", got)
	}
	effectstest.D2dSetHand(me, 3)
	effectstest.D2dSetHand(opp, 1)
	effectstest.D2dSetHand(other, 2)
	oppLife, otherLife := opp.Life, other.Life

	effectstest.AdvanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-2 {
		t.Errorf("the opponent with one card is at %d, want %d", opp.Life, oppLife-2)
	}

	effectstest.AdvanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	effectstest.PassPriorityAroundTable(t, g)
	if other.Life != otherLife {
		t.Errorf("the opponent with two cards lost %d life", otherLife-other.Life)
	}
}

func TestBlbBC2BanditsTalentLevelThreeDrawsPerOpponentWithFewCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := effectstest.BlbB1Seats(g)
	third := g.Seats[(g.Turn.ActiveSeat+3)%len(g.Seats)]
	class := effectstest.PushClass(g, me.ID, "Bandit's Talent", "Enchantment — Class", blbBC2BanditsOracle)
	effectstest.ToMainForCost(t, g)
	if err := effectstest.LevelUpTo(t, g, me.ID, class, 0, "{B}"); err != nil {
		t.Fatalf("level 2: %v", err)
	}
	if err := effectstest.LevelUpTo(t, g, me.ID, class, 1, "{3}{B}"); err != nil {
		t.Fatalf("level 3: %v", err)
	}
	if got := effectstest.LevelOf(g, class); got != 3 {
		t.Fatalf("level %d, want 3", got)
	}
	effectstest.D2dSetHand(me, 3)
	seat := g.Turn.ActiveSeat
	for i := 0; i < 400; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
		if g.Turn.Step == game.StepDraw && g.Turn.ActiveSeat == seat {
			break
		}
	}
	effectstest.D2dSetHand(opp, 0)
	effectstest.D2dSetHand(other, 1)
	effectstest.D2dSetHand(third, 5)
	hand := effectstest.HandCount(g, me)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandCount(g, me); got != hand+2 {
		t.Errorf("hand %d → %d, want two additional cards (two opponents hold one or fewer)", hand, got)
	}
}

// --- Baylen, the Haymaker -----------------------------------------------------

func TestBlbBC2BaylenTapsTwoTokensForManaAndRefusesANontoken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	baylen := effectstest.BlbB1Push(g, me.ID, "Baylen, the Haymaker", "Legendary Creature — Rabbit Warrior", blbBC2BaylenOracle, 4, 3)
	bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	tokens := effectstest.BlbBC2Tokens(t, g, me.ID, 2)
	if err := g.ActivateManaAbility(me.ID, baylen, 0, game.ManaAbilityParams{TapIDs: []uuid.UUID{tokens[0], bear}}); err == nil {
		t.Fatal("a nontoken creature paid the token cost")
	}
	effectstest.AssertTapped(t, g, false, tokens[0], bear, baylen)
	effectstest.ActivateManaFor(t, g, me.ID, baylen, 0, game.ManaAbilityParams{TapIDs: tokens})
	effectstest.ResolveManaPickAnyColor(t, g, me.ID)
	effectstest.AssertTapped(t, g, true, tokens...)
	effectstest.AssertTapped(t, g, false, baylen)
	if me.ManaPool.Total() != 1 {
		t.Errorf("mana pool = %v, want one mana", me.ManaPool)
	}
}

func TestBlbBC2BaylenDrawsAndGrowsByTappingTokens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	baylen := effectstest.BlbB1Push(g, me.ID, "Baylen, the Haymaker", "Legendary Creature — Rabbit Warrior", blbBC2BaylenOracle, 4, 3)
	tokens := effectstest.BlbBC2Tokens(t, g, me.ID, 7)
	effectstest.ToMainForCost(t, g)

	hand := effectstest.HandCount(g, me)
	if err := g.ActivateCatalogAbility(me.ID, baylen, 0, game.ActivateAbilityParams{TapIDs: tokens[:2]}); err == nil {
		t.Fatal("the draw was paid with two tokens")
	}
	if err := g.ActivateCatalogAbility(me.ID, baylen, 0, game.ActivateAbilityParams{TapIDs: tokens[:3]}); err != nil {
		t.Fatalf("draw: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandCount(g, me); got != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, got)
	}

	if err := g.ActivateCatalogAbility(me.ID, baylen, 1, game.ActivateAbilityParams{TapIDs: tokens[3:]}); err != nil {
		t.Fatalf("grow: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, baylen, game.CounterPlusOne); got != 3 {
		t.Errorf("Baylen has %d +1/+1 counters, want 3", got)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, baylen, "trample") {
		t.Error("Baylen should have trample until end of turn")
	}
	effectstest.AssertTapped(t, g, false, baylen)
}

// --- Blacksmith's Talent ------------------------------------------------------

func TestBlbBC2BlacksmithsTalentMakesASwordThatEquips(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "Blacksmith's Talent", "Enchantment — Class", blbBC2BlacksmithOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Sword"); n != 1 {
		t.Fatalf("%d Sword tokens, want 1", n)
	}
	var sword game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Sword" && effects.IsToken(c) {
			sword = c
		}
	}
	if !sword.IsArtifact() || !sword.HasSubtype("Equipment") || len(sword.Colors) != 0 || sword.AttachedTo.ID != uuid.Nil {
		t.Errorf("the Sword is %q, colors %v, attached %v; want an unattached colorless Equipment", sword.TypeLine, sword.Colors, sword.AttachedTo)
	}
	bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, sword.InstanceID, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("equip: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the equipped bear is %d/%d, want 3/3", p, tough)
	}
}

func TestBlbBC2BlacksmithsTalentLevelTwoAttachesAtBeginningOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	class := effectstest.PushClass(g, me.ID, "Blacksmith's Talent", "Enchantment — Class", blbBC2BlacksmithOracle)
	blade := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Test Blade", TypeLine: "Artifact — Equipment"})
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.ToMainForCost(t, g)
	if err := effectstest.LevelUpTo(t, g, me.ID, class, 0, "{2}{R}"); err != nil {
		t.Fatalf("level 2: %v", err)
	}
	effectstest.AdvanceToStepInTurn(t, g, game.StepBeginCombat)
	first := effectstest.LatestPickTarget(g, me.ID)
	if first == nil {
		t.Fatal("no prompt for the Equipment at the beginning of combat")
	}
	if err := g.ResolvePickTarget(first.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: blade}); err != nil {
		t.Fatalf("pick the Equipment: %v", err)
	}
	second := effectstest.LatestPickTarget(g, me.ID)
	if second == nil {
		t.Fatal("no prompt for the creature")
	}
	if err := g.ResolvePickTargets(second.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: bear, Slot: 1}}); err != nil {
		t.Fatalf("pick the creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if b := effectstest.FindBattlefieldCardByID(g, blade); b == nil || !b.IsAttachedTo(bear) {
		t.Fatalf("the Equipment should be attached to the bear: %+v", b)
	}
}

func TestBlbBC2BlacksmithsTalentLevelThreeGivesEquippedCreaturesDoubleStrikeAndHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	class := effectstest.PushClass(g, me.ID, "Blacksmith's Talent", "Enchantment — Class", blbBC2BlacksmithOracle)
	blade := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Test Blade", TypeLine: "Artifact — Equipment"})
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	if err := g.AttachForEffect(blade, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
		t.Fatalf("attach: %v", err)
	}
	effectstest.ToMainForCost(t, g)
	if err := effectstest.LevelUpTo(t, g, me.ID, class, 0, "{2}{R}"); err != nil {
		t.Fatalf("level 2: %v", err)
	}
	if effectstest.HasKeywordOnBattlefield(t, g, bear, "double strike") {
		t.Fatal("double strike before level 3")
	}
	if err := effectstest.LevelUpTo(t, g, me.ID, class, 1, "{3}{R}"); err != nil {
		t.Fatalf("level 3: %v", err)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "double strike") || !effectstest.HasKeywordOnBattlefield(t, g, bear, "haste") {
		t.Error("the equipped bear should have double strike and haste during your turn")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, plain, "double strike") || effectstest.HasKeywordOnBattlefield(t, g, plain, "haste") {
		t.Error("an unequipped creature gained the keywords")
	}
}

// --- Builder's Talent ---------------------------------------------------------

func TestBlbBC2BuildersTalentMakesAWall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "Builder's Talent", "Enchantment — Class", blbBC2BuilderOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	wall := effectstest.FindBattlefieldCardByID(g, effectstest.FindBattlefieldByName(g, "Wall"))
	if wall == nil || !effects.IsToken(*wall) {
		t.Fatal("no Wall token")
	}
	if p, tough := effectstest.PtOf(t, g, wall.InstanceID); p != 0 || tough != 4 {
		t.Errorf("the Wall is %d/%d, want 0/4", p, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, wall.InstanceID, "defender") {
		t.Error("the Wall should have defender")
	}
	if wall.Controller != me.ID {
		t.Error("the Wall is not yours")
	}
}

func TestBlbBC2BuildersTalentLevelTwoCountersOnANoncreaturePermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	class := effectstest.PushClass(g, me.ID, "Builder's Talent", "Enchantment — Class", blbBC2BuilderOracle)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.ToMainForCost(t, g)
	if err := effectstest.LevelUpTo(t, g, me.ID, class, 0, "{W}"); err != nil {
		t.Fatalf("level 2: %v", err)
	}
	effectstest.CastHoldCreature(t, g, "Other Bear", "Creature — Bear", "", 2, 2)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a creature entering triggered level 2")
	}
	effectstest.CastCatalogSpell(t, g, "Trinket", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the bear has %d +1/+1 counters, want 1", got)
	}
}

func TestBlbBC2BuildersTalentLevelThreeReturnsAPermanentCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	class := effectstest.PushClass(g, me.ID, "Builder's Talent", "Enchantment — Class", blbBC2BuilderOracle)
	relic := effectstest.PushGraveyardTyped(me, "Relic", "Artifact")
	effectstest.ToMainForCost(t, g)
	if err := effectstest.LevelUpTo(t, g, me.ID, class, 0, "{W}"); err != nil {
		t.Fatalf("level 2: %v", err)
	}
	if err := effectstest.LevelUpTo(t, g, me.ID, class, 1, "{4}{W}"); err != nil {
		t.Fatalf("level 3: %v", err)
	}
	effectstest.AnswerPickTarget(t, g, relic)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(relic) {
		t.Fatal("the artifact card did not return to the battlefield")
	}
	if c := effectstest.FindBattlefieldCardByID(g, relic); c == nil || c.Controller != me.ID {
		t.Error("the returned card is not under your control")
	}
}

// --- Byrke, Long Ear of the Law -----------------------------------------------

func TestBlbBC2ByrkePutsACounterOnEachOfTwoTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	mine := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.BlbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	effectstest.CastHoldCreature(t, g, "Byrke, Long Ear of the Law", "Legendary Creature — Rabbit Soldier", blbBC2ByrkeOracle, 4, 4)
	prompt := effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no target prompt for the entry")
	}
	if err := g.ResolvePickTargets(prompt.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}}); err != nil {
		t.Fatalf("pick two creatures: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, theirs} {
		if got := effectstest.CountersOn(g, id, game.CounterPlusOne); got != 1 {
			t.Errorf("creature %s has %d +1/+1 counters, want 1", id, got)
		}
	}
}

func TestBlbBC2ByrkeDoublesTheCountersOnACounteredAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.BlbB1Push(g, me.ID, "Byrke, Long Ear of the Law", "Legendary Creature — Rabbit Soldier", blbBC2ByrkeOracle, 4, 4)
	grown := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Grown Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterPlusOne: 2},
	})
	plain := effectstest.BlbB1Push(g, me.ID, "Plain Bear", "Creature — Bear", "", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, grown, plain)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, grown, game.CounterPlusOne); got != 4 {
		t.Errorf("the countered attacker has %d +1/+1 counters, want 4", got)
	}
	if got := effectstest.CountersOn(g, plain, game.CounterPlusOne); got != 0 {
		t.Errorf("the attacker with no counter got %d", got)
	}
}

// --- Cache Grab ---------------------------------------------------------------

// blbBC2CastCacheGrab seeds the library and casts Cache Grab.
func blbBC2CastCacheGrab(t *testing.T, g *game.Game, me *game.Player, lib ...game.Card) []uuid.UUID {
	t.Helper()
	effectstest.ToMainForCost(t, g)
	ids := effectstest.Mm2Library(me, lib...)
	id := effectstest.HandCardFull(me, "Cache Grab", "Instant", "", blbBC2CacheGrabOracle, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return ids
}

func TestBlbBC2CacheGrabReturningASquirrelMakesAFood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := blbBC2CastCacheGrab(t, g, me, effectstest.Mm2Spell("A"),
		game.Card{Name: "Nut Hoarder", TypeLine: "Creature — Squirrel"}, effectstest.Mm2Spell("B"), effectstest.Mm2Spell("C"), effectstest.Mm2Spell("Deep"))
	effectstest.AnswerChooseCards(t, g, me.ID, ids[1])
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[1]) {
		t.Fatal("the Squirrel card should be in hand")
	}
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Food"); n != 1 {
		t.Errorf("%d Food tokens, want 1", n)
	}
}

func TestBlbBC2CacheGrabWithoutASquirrelMakesNoFood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := blbBC2CastCacheGrab(t, g, me, effectstest.Mm2Spell("A"),
		game.Card{Name: "Relic", TypeLine: "Artifact"}, effectstest.Mm2Spell("B"), effectstest.Mm2Spell("C"), effectstest.Mm2Spell("Deep"))
	effectstest.AnswerChooseCards(t, g, me.ID, ids[1])
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[1]) {
		t.Fatal("the permanent card should be in hand")
	}
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Food"); n != 0 {
		t.Errorf("%d Food tokens with no Squirrel, want 0", n)
	}
}

func TestBlbBC2CacheGrabControllingASquirrelMakesAFood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.BlbB1Push(g, me.ID, "Squirrel", "Creature — Squirrel", "", 1, 1)
	ids := blbBC2CastCacheGrab(t, g, me, effectstest.Mm2Spell("A"),
		game.Card{Name: "Relic", TypeLine: "Artifact"}, effectstest.Mm2Spell("B"), effectstest.Mm2Spell("C"), effectstest.Mm2Spell("Deep"))
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(ids[1]) {
		t.Fatal("a declined pick put a card in hand")
	}
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Food"); n != 1 {
		t.Errorf("%d Food tokens while you control a Squirrel, want 1", n)
	}
	if me.Library.Size() != 1 {
		t.Errorf("library %d, want exactly four milled", me.Library.Size())
	}
}

// --- Clement, the Worrywort ---------------------------------------------------

func TestBlbBC2ClementBouncesACheaperCreatureWhenOneEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.BlbB1Push(g, me.ID, "Clement, the Worrywort", "Legendary Creature — Frog Druid", blbBC2ClementOracle, 3, 3)
	cheap := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Cub", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	big := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Giant", TypeLine: "Creature — Giant", ManaCost: "{4}{G}",
		Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	if _, err := effectstest.BlbHCastWithMana(t, g, "Ogre", "Creature — Ogre", "", "{2}{G}", 3, 3, "G", "G", "G"); err != nil {
		t.Fatalf("cast the Ogre: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no target prompt when a creature entered")
	}
	if err := g.ResolvePickTarget(prompt.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: big}); err == nil {
		t.Fatal("a creature with greater mana value was a legal target")
	}
	if err := g.ResolvePickTarget(prompt.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: cheap}); err != nil {
		t.Fatalf("pick the cheaper creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(cheap) {
		t.Error("the cheaper creature did not return to its owner's hand")
	}
	if !g.Battlefield.Contains(big) {
		t.Error("the bigger creature left the battlefield")
	}
}

func TestBlbBC2ClementGivesFrogsRestrictedMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	clement := effectstest.BlbB1Push(g, me.ID, "Clement, the Worrywort", "Legendary Creature — Frog Druid", blbBC2ClementOracle, 3, 3)
	frog := effectstest.B12Creature(g, me.ID, "Frog", "Creature — Frog", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateManaAbility(me.ID, bear, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("a non-Frog was given the mana ability")
	}
	for _, id := range []uuid.UUID{clement, frog} {
		effectstest.ActivateManaFor(t, g, me.ID, id, 0, game.ManaAbilityParams{})
		if c := effectstest.LatestChoiceOfKind(g, game.PendingChoiceMana); c != nil {
			if err := g.ResolveManaChoice(c.ID, me.ID, "U"); err != nil {
				t.Fatalf("ResolveManaChoice: %v", err)
			}
		}
	}
	if me.ManaPool.Total() != 2 {
		t.Fatalf("mana pool = %v, want two mana", me.ManaPool)
	}
	for _, m := range me.ManaPool.Units() {
		if m.Color != "U" || len(m.Restrictions) == 0 {
			t.Errorf("mana %+v, want blue mana that only casts a creature spell", m)
		}
	}
}

// --- Cruelclaw's Heist --------------------------------------------------------

// blbBC2Heist casts Cruelclaw's Heist at the next seat, promising the gift
// to it when `promise`, and exiles the Big Bear from its hand.
func blbBC2Heist(t *testing.T, g *game.Game, promise bool) (me, opp *game.Player, bear uuid.UUID) {
	t.Helper()
	me, opp, _ = effectstest.BlbB1Seats(g)
	ids := effectstest.RevealHand(opp, effectstest.RhForest(), effectstest.RhBolt(),
		game.Card{Name: "Big Bear", TypeLine: "Creature — Bear", ManaCost: "{3}{R}{R}", Colors: []string{"R"}, Power: 3, Toughness: 3})
	bear = ids[2]
	to := uuid.Nil
	if promise {
		to = opp.ID
	}
	effectstest.MustCastWithGift(t, g, "Cruelclaw's Heist", "Sorcery", blbBC2HeistOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, to)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.OpenVariant(t, g)
	if pick.PickDestination != game.PickExile {
		t.Fatalf("the pick goes to %q, want exile", pick.PickDestination)
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("resolve the pick: %v", err)
	}
	if !g.Exile.Contains(bear) {
		t.Fatal("the chosen card was not exiled")
	}
	// Let the resolution settle, then cast in a main phase of your turn.
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(g.StackMeta); n != 0 {
		t.Fatalf("%d items still on the stack: %+v", n, g.StackMeta)
	}
	g.Turn.Step = game.StepPrecombatMain
	effectstest.FloatForTest(g, me, "UUUUU")
	return me, opp, bear
}

func TestBlbBC2CruelclawsHeistPromisedLetsYouCastTheCardWithAnyMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, bear := blbBC2Heist(t, g, true)
	if opp.Hand.Size() != 3 {
		t.Errorf("the gifted opponent holds %d cards, want 3 (drew one, lost one)", opp.Hand.Size())
	}
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast the exiled card with blue mana: %v", err)
	}
}

func TestBlbBC2CruelclawsHeistUnpromisedOnlyExiles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, bear := blbBC2Heist(t, g, false)
	if opp.Hand.Size() != 2 {
		t.Errorf("the opponent holds %d cards, want 2", opp.Hand.Size())
	}
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Fatal("cast the exiled card with no gift promised")
	}
}
