package testsst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Ugin, the Spirit Dragon ------------------------------------

func TestUginPlusTwoDealsThreeToAnyTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	them := g.Seats[(seat+1)%len(g.Seats)]
	ugin := effectstest.PushWalkerForTest(g, owner.ID, "Ugin, the Spirit Dragon", effectstest.UginSpiritDragonOracle, 7)
	effectstest.AdvanceToMainOf(t, g, seat)
	life := them.Life

	// "Any target" reaches a player, which is how a colorless Ugin
	// is usually pointed at a four-player table.
	err := g.ActivateCatalogAbility(owner.ID, ugin, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: them.ID}},
	})
	if err != nil {
		t.Fatalf("+2: %v", err)
	}
	if got := effectstest.LoyaltyOf(g, ugin); got != 9 {
		t.Errorf("loyalty after +2 = %d, want 9", got)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := life - them.Life; got != 3 {
		t.Errorf("damage to the targeted player = %d, want 3", got)
	}
}

// TestUginPlusTwoKillsASmallCreature — the other leg of "any
// target": three damage to a 2/2 is lethal, so it dies to the SBA
// rather than sitting there marked.
func TestUginPlusTwoKillsASmallCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	them := g.Seats[(seat+1)%len(g.Seats)]
	ugin := effectstest.PushWalkerForTest(g, owner.ID, "Ugin, the Spirit Dragon", effectstest.UginSpiritDragonOracle, 7)
	victim := effectstest.PushCreatureToBattlefieldForTest(g, them.ID, "Victim")
	effectstest.AdvanceToMainOf(t, g, seat)

	err := g.ActivateCatalogAbility(owner.ID, ugin, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	})
	if err != nil {
		t.Fatalf("+2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(victim) {
		t.Error("a 2/2 survived three damage")
	}
}

// TestUginMinusTenGainsLifeAndDraws pins the life and the draw; the
// "put up to seven permanent cards" pick that follows is
// TestUginMinusTenPutsUpToSevenPermanentsFromHand (#3861).
func TestUginMinusTenGainsLifeAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	ugin := effectstest.PushWalkerForTest(g, owner.ID, "Ugin, the Spirit Dragon", effectstest.UginSpiritDragonOracle, 12)
	effectstest.AdvanceToMainOf(t, g, seat)
	life, hand := owner.Life, owner.Hand.Size()

	if err := g.ActivateCatalogAbility(owner.ID, ugin, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−10: %v", err)
	}
	if got := effectstest.LoyaltyOf(g, ugin); got != 2 {
		t.Errorf("loyalty after −10 = %d, want 2", got)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := owner.Life - life; got != 7 {
		t.Errorf("life gained = %d, want 7", got)
	}
	if got := owner.Hand.Size() - hand; got != 7 {
		t.Errorf("cards drawn = %d, want 7", got)
	}
}

// --- Liliana of the Veil ----------------------------------------

// TestLilianaPlusOneMakesEveryPlayerDiscard — symmetric, the
// controller included, and each seat chooses their own card.
func TestLilianaPlusOneMakesEveryPlayerDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	lili := effectstest.PushWalkerForTest(g, owner.ID, "Liliana of the Veil", effectstest.LilianaOfTheVeilOracle, 3)
	effectstest.AdvanceToMainOf(t, g, seat)

	if err := g.ActivateCatalogAbility(owner.ID, lili, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	for _, p := range g.Seats {
		if effectstest.DiscardOwed(g, p.ID) != 1 {
			t.Errorf("seat %s owes %d discards, want 1", p.Name, effectstest.DiscardOwed(g, p.ID))
		}
	}
}

// TestLilianaMinusTwoIsAnEdict — the victim picks from their own
// creatures. Not a target on the creature, so hexproof is irrelevant.
func TestLilianaMinusTwoIsAnEdict(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	them := g.Seats[(seat+1)%len(g.Seats)]
	lili := effectstest.PushWalkerForTest(g, owner.ID, "Liliana of the Veil", effectstest.LilianaOfTheVeilOracle, 3)
	effectstest.PushCreatureToBattlefieldForTest(g, them.ID, "Theirs")
	effectstest.AdvanceToMainOf(t, g, seat)

	err := g.ActivateCatalogAbility(owner.ID, lili, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: them.ID}},
	})
	if err != nil {
		t.Fatalf("−2: %v", err)
	}
	if got := effectstest.LoyaltyOf(g, lili); got != 1 {
		t.Errorf("loyalty after −2 = %d, want 1", got)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if !hasSacrificeChoiceFor(g, them.ID) {
		t.Error("the targeted player was not asked to sacrifice a creature")
	}
}

// --- Teferi, Temporal Pilgrim -----------------------------------

// TestTeferiPilgrimZeroDrawsAndTriggersItsOwnLoyalty is the card in
// one activation: [0] is a real printed cost that spends the turn's
// window, the draw fires "whenever you draw a card", and the trigger
// puts the loyalty counter back on.
func TestTeferiPilgrimZeroDrawsAndTriggersItsOwnLoyalty(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	teferi := effectstest.PushWalkerForTest(g, owner.ID, "Teferi, Temporal Pilgrim", effectstest.TeferiTemporalPilgrimOracle, 4)
	effectstest.AdvanceToMainOf(t, g, seat)
	hand := owner.Hand.Size()

	if err := g.ActivateCatalogAbility(owner.ID, teferi, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("[0]: %v", err)
	}
	// [0] pays nothing, and that is the point of the pointer.
	if got := effectstest.LoyaltyOf(g, teferi); got != 4 {
		t.Errorf("loyalty after [0] = %d, want 4", got)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := owner.Hand.Size() - hand; got != 1 {
		t.Fatalf("cards drawn = %d, want 1", got)
	}
	if got := effectstest.LoyaltyOf(g, teferi); got != 5 {
		t.Errorf("loyalty after the draw trigger = %d, want 5", got)
	}
	// CR 606.3: [0] still burns the turn's activation.
	if err := g.ActivateCatalogAbility(owner.ID, teferi, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("a [0] activation did not spend the once-per-turn window")
	}
}

// TestTeferiPilgrimDrawTriggerIgnoresOpponents — "whenever YOU draw
// a card".
func TestTeferiPilgrimDrawTriggerIgnoresOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	them := g.Seats[(seat+1)%len(g.Seats)]
	teferi := effectstest.PushWalkerForTest(g, owner.ID, "Teferi, Temporal Pilgrim", effectstest.TeferiTemporalPilgrimOracle, 4)
	effectstest.AdvanceToMainOf(t, g, seat)

	if err := g.DrawNForEffect(them.ID, 1); err != nil {
		t.Fatalf("opponent draw: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoyaltyOf(g, teferi); got != 4 {
		t.Errorf("loyalty = %d, want 4 — an opponent's draw is not yours", got)
	}
}

func TestTeferiPilgrimMinusTwoMakesAVigilantSpirit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	teferi := effectstest.PushWalkerForTest(g, owner.ID, "Teferi, Temporal Pilgrim", effectstest.TeferiTemporalPilgrimOracle, 4)
	effectstest.AdvanceToMainOf(t, g, seat)

	if err := g.ActivateCatalogAbility(owner.ID, teferi, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	var spirit *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Spirit" {
			spirit = &g.Battlefield.Cards[i]
		}
	}
	if spirit == nil {
		t.Fatal("no Spirit token was created")
	}
	if spirit.Power != 2 || spirit.Toughness != 2 {
		t.Errorf("Spirit is %d/%d, want 2/2", spirit.Power, spirit.Toughness)
	}
	if !game.HasKeyword(spirit, "vigilance") {
		t.Error("the Spirit token is missing vigilance")
	}
}

// --- Teferi, Hero of Dominaria ----------------------------------

// TestTeferiHeroMinusThreeTucksThirdFromTheTop is the printed
// position, which is the whole difference between this removal and
// "put it on top".
func TestTeferiHeroMinusThreeTucksThirdFromTheTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	them := g.Seats[(seat+1)%len(g.Seats)]
	teferi := effectstest.PushWalkerForTest(g, owner.ID, "Teferi, Hero of Dominaria", effectstest.TeferiHeroOfDominariaOracle, 4)
	victim := effectstest.PushCreatureToBattlefieldForTest(g, them.ID, "Victim")
	before := them.Library.Size()
	effectstest.AdvanceToMainOf(t, g, seat)

	err := g.ActivateCatalogAbility(owner.ID, teferi, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	})
	if err != nil {
		t.Fatalf("−3: %v", err)
	}
	if got := effectstest.LoyaltyOf(g, teferi); got != 1 {
		t.Errorf("loyalty after −3 = %d, want 1", got)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if them.Library.Size() != before+1 {
		t.Fatalf("library = %d cards, want %d", them.Library.Size(), before+1)
	}
	// Zone order: the top is the LAST element.
	top := them.Library.Size() - 1
	if got := them.Library.Cards[top].InstanceID; got == victim {
		t.Error("the permanent went to the top of the library, not third from it")
	}
	if got := them.Library.Cards[top-2].InstanceID; got != victim {
		t.Errorf("third-from-top is %s, want the tucked permanent",
			them.Library.Cards[top-2].Name)
	}
}

// TestTeferiHeroTucksACommanderThroughTheCommandZonePrompt is the
// CR 903.9b interaction #529 / #539 opened up: a library is a 903.9b
// destination, so a tucked commander's OWNER is asked whether to
// send it to the command zone instead. The prompt is the assertion —
// nothing has moved while it is open.
func TestTeferiHeroTucksACommanderThroughTheCommandZonePrompt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	them := g.Seats[(seat+1)%len(g.Seats)]
	teferi := effectstest.PushWalkerForTest(g, owner.ID, "Teferi, Hero of Dominaria", effectstest.TeferiHeroOfDominariaOracle, 4)

	cmdr := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID:  cmdr,
		Name:        "Their Commander",
		TypeLine:    "Legendary Creature — Test",
		Power:       3,
		Toughness:   3,
		Owner:       them.ID,
		Controller:  them.ID,
		IsCommander: true,
	})
	effectstest.AdvanceToMainOf(t, g, seat)

	err := g.ActivateCatalogAbility(owner.ID, teferi, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: cmdr}},
	})
	if err != nil {
		t.Fatalf("−3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if !hasOptionalReplacementFor(g, them.ID) {
		t.Fatal("a tucked commander did not offer its owner the command zone")
	}
	// The card has not moved while the choice is open.
	if !g.Battlefield.Contains(cmdr) {
		t.Error("the commander left the battlefield before its owner answered")
	}
}

// TestTeferiHeroPlusOneDrawsAndUntapsAtTheNextEndStep — the untap is
// a delayed trigger whose lands are chosen by the controller at
// resolution (#1337: UntapUpToLands, not an auto-pick).
func TestTeferiHeroPlusOneDrawsAndUntapsAtTheNextEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	teferi := effectstest.PushWalkerForTest(g, owner.ID, "Teferi, Hero of Dominaria", effectstest.TeferiHeroOfDominariaOracle, 4)

	var lands []uuid.UUID
	for i := 0; i < 3; i++ {
		id := uuid.New()
		g.Battlefield.PushTop(game.Card{
			InstanceID: id,
			Name:       "Island",
			TypeLine:   "Basic Land — Island",
			Owner:      owner.ID,
			Controller: owner.ID,
			Tapped:     true,
		})
		lands = append(lands, id)
	}
	effectstest.AdvanceToMainOf(t, g, seat)
	hand := owner.Hand.Size()

	if err := g.ActivateCatalogAbility(owner.ID, teferi, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := owner.Hand.Size() - hand; got != 1 {
		t.Fatalf("cards drawn = %d, want 1", got)
	}
	if untappedCount(g, lands) != 0 {
		t.Fatal("the untap happened immediately rather than at the end step")
	}

	effectstest.AdvanceToStepOnThisTurn(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, owner.ID, lands[:2]...) // "untap up to two lands", 3 tapped offered

	if got := untappedCount(g, lands); got != 2 {
		t.Errorf("untapped lands at the end step = %d, want exactly 2 of 3", got)
	}
}

// --- Karn Liberated ---------------------------------------------

// TestKarnPlusFourMakesTheTargetExileFromHand — the target picks,
// from their own hand, and the card goes to exile, not the graveyard
// (#3861). Nobody is asked to discard.
func TestKarnPlusFourMakesTheTargetExileFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	them := g.Seats[(seat+1)%len(g.Seats)]
	karn := effectstest.PushWalkerForTest(g, owner.ID, "Karn Liberated", effectstest.KarnLiberatedOracle, 6)
	effectstest.AdvanceToMainOf(t, g, seat)
	keep := effectstest.PushHandCardWithManaCost(them, "Their Keeper", "Creature — Bear", "{1}{G}")
	lose := effectstest.PushHandCardWithManaCost(them, "Their Loss", "Instant", "{U}")

	err := g.ActivateCatalogAbility(owner.ID, karn, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: them.ID}},
	})
	if err != nil {
		t.Fatalf("+4: %v", err)
	}
	if got := effectstest.LoyaltyOf(g, karn); got != 10 {
		t.Errorf("loyalty after +4 = %d, want 10", got)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.ChooseCardsChoiceFor(g, owner.ID) != nil {
		t.Error("Karn's controller was asked to pick from the target's hand")
	}
	pick := effectstest.ChooseCardsChoiceFor(g, them.ID)
	if pick == nil {
		t.Fatal("the target was not asked which card to exile")
	}
	if pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Errorf("pick bounds = [%d, %d], want exactly one", pick.ChooseMin, pick.ChooseMax)
	}
	if !effectstest.HasID(pick.ChooseCards, keep) || !effectstest.HasID(pick.ChooseCards, lose) {
		t.Errorf("the pick does not offer the whole hand: %v", pick.ChooseCards)
	}
	if err := g.ResolveChooseCards(pick.ID, them.ID, nil); err == nil {
		t.Error("declining was accepted; exiling a card is not optional")
	}
	if err := g.ResolveChooseCards(pick.ID, them.ID, []uuid.UUID{lose}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if z := g.FindCardZoneForEffect(lose); z == nil || z.Kind != game.ZoneExile {
		t.Errorf("the chosen card is in %v, want exile", z)
	}
	if them.Graveyard.Contains(lose) {
		t.Error("the chosen card went to the graveyard")
	}
	if !them.Hand.Contains(keep) {
		t.Error("the card not chosen left the hand")
	}
}

// TestKarnPlusFourOnAnEmptyHandAsksNothing — nothing to exile, no
// prompt to wedge the table on.
func TestKarnPlusFourOnAnEmptyHandAsksNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	them := g.Seats[(seat+1)%len(g.Seats)]
	karn := effectstest.PushWalkerForTest(g, owner.ID, "Karn Liberated", effectstest.KarnLiberatedOracle, 6)
	effectstest.AdvanceToMainOf(t, g, seat)
	them.Hand.Cards = nil

	if err := g.ActivateCatalogAbility(owner.ID, karn, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: them.ID}},
	}); err != nil {
		t.Fatalf("+4: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if pick := effectstest.ChooseCardsChoiceFor(g, them.ID); pick != nil {
		t.Errorf("a player with no hand was asked: %+v", pick)
	}
}

func TestKarnMinusThreeExilesAnyPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	them := g.Seats[(seat+1)%len(g.Seats)]
	karn := effectstest.PushWalkerForTest(g, owner.ID, "Karn Liberated", effectstest.KarnLiberatedOracle, 6)
	victim := effectstest.PushCreatureToBattlefieldForTest(g, them.ID, "Victim")
	effectstest.AdvanceToMainOf(t, g, seat)

	err := g.ActivateCatalogAbility(owner.ID, karn, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	})
	if err != nil {
		t.Fatalf("−3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(victim) {
		t.Error("the permanent was not exiled")
	}
	if !g.Exile.Contains(victim) {
		t.Error("the permanent is not in exile")
	}
}

// --- Saheeli Rai ------------------------------------------------

func TestSaheeliPlusOneScriesThenPingsEachOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	saheeli := effectstest.PushWalkerForTest(g, owner.ID, "Saheeli Rai", effectstest.SaheeliRaiOracle, 3)
	effectstest.AdvanceToMainOf(t, g, seat)

	life := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		life[p.ID] = p.Life
	}
	if err := g.ActivateCatalogAbility(owner.ID, saheeli, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	// The scry queues a prompt; the damage is in its continuation,
	// so it lands once the scry is answered.
	effectstest.ResolveAnyScryFor(t, g, owner.ID)

	for _, p := range g.Seats {
		want := life[p.ID]
		if p.ID != owner.ID {
			want--
		}
		if p.Life != want {
			t.Errorf("seat %s life = %d, want %d", p.Name, p.Life, want)
		}
	}
}

// TestSaheeliMinusTwoCopiesWithHasteAsAnArtifact — the copy is the
// original plus two printed changes, and it leaves at the end step.
func TestSaheeliMinusTwoCopiesWithHasteAsAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	saheeli := effectstest.PushWalkerForTest(g, owner.ID, "Saheeli Rai", effectstest.SaheeliRaiOracle, 3)
	original := effectstest.PushCreatureToBattlefieldForTest(g, owner.ID, "Bear")
	effectstest.AdvanceToMainOf(t, g, seat)

	err := g.ActivateCatalogAbility(owner.ID, saheeli, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: original}},
	})
	if err != nil {
		t.Fatalf("−2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	var dup *game.Card
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.InstanceID != original && c.Name == "Bear" {
			dup = c
		}
	}
	if dup == nil {
		t.Fatal("no token copy was created")
	}
	if !dup.IsArtifact() {
		t.Errorf("the copy's type line %q is not an artifact", dup.TypeLine)
	}
	if !dup.IsCreature() {
		t.Errorf("the copy's type line %q stopped being a creature", dup.TypeLine)
	}
	if !game.HasKeyword(dup, "haste") {
		t.Error("the copy did not gain haste")
	}
	token := dup.InstanceID

	effectstest.AdvanceToStepOnThisTurn(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(token) {
		t.Error("the copy survived the end step")
	}
}

// TestSaheeliArtifactInAdditionKeepsTheOtherTypes is the type-line
// surgery on its own, including the "already an artifact" case.
func TestSaheeliArtifactInAdditionKeepsTheOtherTypes(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"Token Creature — Bear", "Token Artifact Creature — Bear"},
		{"Token Legendary Creature — Elf Druid", "Token Legendary Artifact Creature — Elf Druid"},
		{"Token Artifact — Equipment", "Token Artifact — Equipment"},
		{"Token Enchantment", "Token Artifact Enchantment"},
	} {
		if got := effects.SaheeliArtifactInAddition(tc.in); got != tc.want {
			t.Errorf("saheeliArtifactInAddition(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// The −7's "with different names" is the shared search rule (#4253 slice 5).
func TestSaheeliDifferentNamesRejectsARepeat(t *testing.T) {
	sol := game.Card{Name: "Sol Ring"}
	mox := game.Card{Name: "Mox Diamond"}
	if !effects.DifferentNames([]game.Card{sol, mox}) {
		t.Error("two different names were rejected")
	}
	if effects.DifferentNames([]game.Card{sol, sol}) {
		t.Error("two copies of one name were accepted")
	}
	if !effects.DifferentNames(nil) {
		t.Error("failing to find was rejected")
	}
}

// --- shared assertions ------------------------------------------

func untappedCount(g *game.Game, ids []uuid.UUID) int {
	want := map[uuid.UUID]bool{}
	for _, id := range ids {
		want[id] = true
	}
	n := 0
	for _, c := range g.Battlefield.Cards {
		if want[c.InstanceID] && !c.Tapped {
			n++
		}
	}
	return n
}

func hasSacrificeChoiceFor(g *game.Game, player uuid.UUID) bool {
	for _, c := range g.PendingChoices {
		if effectstest.IsSacrificePrompt(c) && c.Chooser == player {
			return true
		}
	}
	return false
}

func hasOptionalReplacementFor(g *game.Game, player uuid.UUID) bool {
	for _, c := range g.PendingChoices {
		if c.Kind == game.PendingChoiceOptionalReplacement && c.Chooser == player {
			return true
		}
	}
	return false
}
