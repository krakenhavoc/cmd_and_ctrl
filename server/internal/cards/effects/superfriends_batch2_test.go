package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// superfriends_batch2_test.go — S27's remaining six planeswalkers.
// The helpers (pushWalkerForTest, loyaltyOf, advanceToMainOf,
// countNamed) are superfriends_test.go's; this file is separate only
// so two concurrent card batches do not collide on one test file.

const (
	uginSpiritDragonOracle      = "eecb3047-a563-441a-9175-200421981ac3"
	lilianaOfTheVeilOracle      = "0ba134d8-ee7d-48ec-8dc6-57942b8e9261"
	teferiTemporalPilgrimOracle = "5f0fa785-37ab-46a8-9ec0-6f8b29576d93"
	teferiHeroOfDominariaOracle = "f2f165b6-ef0a-42ad-9352-ba68be8248b0"
	karnLiberatedOracle         = "0ca233f4-1b7f-4807-ab6e-2b1f5439b3db"
	saheeliRaiOracle            = "5f3fe679-aff1-41b3-8d75-c78c2c0636f0"
)

// --- Ugin, the Spirit Dragon ------------------------------------

func TestUginPlusTwoDealsThreeToAnyTarget(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	them := g.Seats[(seat+1)%len(g.Seats)]
	ugin := pushWalkerForTest(g, owner.ID, "Ugin, the Spirit Dragon", uginSpiritDragonOracle, 7)
	advanceToMainOf(t, g, seat)
	life := them.Life

	// "Any target" reaches a player, which is how a colorless Ugin
	// is usually pointed at a four-player table.
	err := g.ActivateCatalogAbility(owner.ID, ugin, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: them.ID}},
	})
	if err != nil {
		t.Fatalf("+2: %v", err)
	}
	if got := loyaltyOf(g, ugin); got != 9 {
		t.Errorf("loyalty after +2 = %d, want 9", got)
	}
	passPriorityAroundTable(t, g)

	if got := life - them.Life; got != 3 {
		t.Errorf("damage to the targeted player = %d, want 3", got)
	}
}

// TestUginPlusTwoKillsASmallCreature — the other leg of "any
// target": three damage to a 2/2 is lethal, so it dies to the SBA
// rather than sitting there marked.
func TestUginPlusTwoKillsASmallCreature(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	them := g.Seats[(seat+1)%len(g.Seats)]
	ugin := pushWalkerForTest(g, owner.ID, "Ugin, the Spirit Dragon", uginSpiritDragonOracle, 7)
	victim := pushCreatureToBattlefieldForTest(g, them.ID, "Victim")
	advanceToMainOf(t, g, seat)

	err := g.ActivateCatalogAbility(owner.ID, ugin, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	})
	if err != nil {
		t.Fatalf("+2: %v", err)
	}
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(victim) {
		t.Error("a 2/2 survived three damage")
	}
}

// TestUginMinusTenGainsLifeAndDraws pins the half of the ultimate
// that ships. The "put up to seven permanent cards from your hand
// onto the battlefield" clause is declared as a caveat and is
// deliberately absent — there is no hand picker to ask with.
func TestUginMinusTenGainsLifeAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	ugin := pushWalkerForTest(g, owner.ID, "Ugin, the Spirit Dragon", uginSpiritDragonOracle, 12)
	advanceToMainOf(t, g, seat)
	life, hand := owner.Life, owner.Hand.Size()

	if err := g.ActivateCatalogAbility(owner.ID, ugin, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−10: %v", err)
	}
	if got := loyaltyOf(g, ugin); got != 2 {
		t.Errorf("loyalty after −10 = %d, want 2", got)
	}
	passPriorityAroundTable(t, g)

	if got := owner.Life - life; got != 7 {
		t.Errorf("life gained = %d, want 7", got)
	}
	if got := owner.Hand.Size() - hand; got != 7 {
		t.Errorf("cards drawn = %d, want 7", got)
	}
}

// TestUginHasNoVariableLoyaltyAbility is the finding, pinned. The −X
// is not registered because a loyalty cost is a fixed number: #550's
// {X} lives in an ability's MANA component and there is no mana
// component on a loyalty ability. If a variable loyalty cost ever
// lands, this test is the one that should start failing.
func TestUginHasNoVariableLoyaltyAbility(t *testing.T) {
	spec, ok := Lookup(uginSpiritDragonOracle)
	if !ok {
		t.Fatal("Ugin is not registered")
	}
	if len(spec.Activated) != 2 {
		t.Fatalf("Ugin declares %d abilities, want 2 (+2 and −10)", len(spec.Activated))
	}
	for i, ab := range spec.Activated {
		if ab.Cost.DemandsX() {
			t.Errorf("ability %d announces an X; loyalty costs are fixed", i)
		}
		if ab.Cost.Loyalty == nil {
			t.Errorf("ability %d is not a loyalty ability", i)
		}
	}
}

// --- Liliana of the Veil ----------------------------------------

// TestLilianaPlusOneMakesEveryPlayerDiscard — symmetric, the
// controller included, and each seat chooses their own card.
func TestLilianaPlusOneMakesEveryPlayerDiscard(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	lili := pushWalkerForTest(g, owner.ID, "Liliana of the Veil", lilianaOfTheVeilOracle, 3)
	advanceToMainOf(t, g, seat)

	if err := g.ActivateCatalogAbility(owner.ID, lili, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	passPriorityAroundTable(t, g)

	for _, p := range g.Seats {
		if g.DiscardPending[p.ID] != 1 {
			t.Errorf("seat %s owes %d discards, want 1", p.Name, g.DiscardPending[p.ID])
		}
	}
}

// TestLilianaMinusTwoIsAnEdict — the victim picks from their own
// creatures. Not a target on the creature, so hexproof is irrelevant.
func TestLilianaMinusTwoIsAnEdict(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	them := g.Seats[(seat+1)%len(g.Seats)]
	lili := pushWalkerForTest(g, owner.ID, "Liliana of the Veil", lilianaOfTheVeilOracle, 3)
	pushCreatureToBattlefieldForTest(g, them.ID, "Theirs")
	advanceToMainOf(t, g, seat)

	err := g.ActivateCatalogAbility(owner.ID, lili, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: them.ID}},
	})
	if err != nil {
		t.Fatalf("−2: %v", err)
	}
	if got := loyaltyOf(g, lili); got != 1 {
		t.Errorf("loyalty after −2 = %d, want 1", got)
	}
	passPriorityAroundTable(t, g)

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
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	teferi := pushWalkerForTest(g, owner.ID, "Teferi, Temporal Pilgrim", teferiTemporalPilgrimOracle, 4)
	advanceToMainOf(t, g, seat)
	hand := owner.Hand.Size()

	if err := g.ActivateCatalogAbility(owner.ID, teferi, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("[0]: %v", err)
	}
	// [0] pays nothing, and that is the point of the pointer.
	if got := loyaltyOf(g, teferi); got != 4 {
		t.Errorf("loyalty after [0] = %d, want 4", got)
	}
	passPriorityAroundTable(t, g)

	if got := owner.Hand.Size() - hand; got != 1 {
		t.Fatalf("cards drawn = %d, want 1", got)
	}
	if got := loyaltyOf(g, teferi); got != 5 {
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
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	them := g.Seats[(seat+1)%len(g.Seats)]
	teferi := pushWalkerForTest(g, owner.ID, "Teferi, Temporal Pilgrim", teferiTemporalPilgrimOracle, 4)
	advanceToMainOf(t, g, seat)

	if err := g.DrawNForEffect(them.ID, 1); err != nil {
		t.Fatalf("opponent draw: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := loyaltyOf(g, teferi); got != 4 {
		t.Errorf("loyalty = %d, want 4 — an opponent's draw is not yours", got)
	}
}

func TestTeferiPilgrimMinusTwoMakesAVigilantSpirit(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	teferi := pushWalkerForTest(g, owner.ID, "Teferi, Temporal Pilgrim", teferiTemporalPilgrimOracle, 4)
	advanceToMainOf(t, g, seat)

	if err := g.ActivateCatalogAbility(owner.ID, teferi, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−2: %v", err)
	}
	passPriorityAroundTable(t, g)

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
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	them := g.Seats[(seat+1)%len(g.Seats)]
	teferi := pushWalkerForTest(g, owner.ID, "Teferi, Hero of Dominaria", teferiHeroOfDominariaOracle, 4)
	victim := pushCreatureToBattlefieldForTest(g, them.ID, "Victim")
	before := them.Library.Size()
	advanceToMainOf(t, g, seat)

	err := g.ActivateCatalogAbility(owner.ID, teferi, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	})
	if err != nil {
		t.Fatalf("−3: %v", err)
	}
	if got := loyaltyOf(g, teferi); got != 1 {
		t.Errorf("loyalty after −3 = %d, want 1", got)
	}
	passPriorityAroundTable(t, g)

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
// CR 903.9 interaction #529 / #539 opened up: a library is a 903.9
// destination, so a tucked commander's OWNER is asked whether to
// send it to the command zone instead. The prompt is the assertion —
// nothing has moved while it is open.
func TestTeferiHeroTucksACommanderThroughTheCommandZonePrompt(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	them := g.Seats[(seat+1)%len(g.Seats)]
	teferi := pushWalkerForTest(g, owner.ID, "Teferi, Hero of Dominaria", teferiHeroOfDominariaOracle, 4)

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
	advanceToMainOf(t, g, seat)

	err := g.ActivateCatalogAbility(owner.ID, teferi, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: cmdr}},
	})
	if err != nil {
		t.Fatalf("−3: %v", err)
	}
	passPriorityAroundTable(t, g)

	if !hasOptionalReplacementFor(g, them.ID) {
		t.Fatal("a tucked commander did not offer its owner the command zone")
	}
	// The card has not moved while the choice is open.
	if !g.Battlefield.Contains(cmdr) {
		t.Error("the commander left the battlefield before its owner answered")
	}
}

// TestTeferiHeroPlusOneDrawsAndUntapsAtTheNextEndStep — the untap is
// a delayed trigger, and the lands it picks are chosen for the
// player (declared caveat).
func TestTeferiHeroPlusOneDrawsAndUntapsAtTheNextEndStep(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	teferi := pushWalkerForTest(g, owner.ID, "Teferi, Hero of Dominaria", teferiHeroOfDominariaOracle, 4)

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
	advanceToMainOf(t, g, seat)
	hand := owner.Hand.Size()

	if err := g.ActivateCatalogAbility(owner.ID, teferi, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := owner.Hand.Size() - hand; got != 1 {
		t.Fatalf("cards drawn = %d, want 1", got)
	}
	if untappedCount(g, lands) != 0 {
		t.Fatal("the untap happened immediately rather than at the end step")
	}

	advanceToStepOnThisTurn(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)

	if got := untappedCount(g, lands); got != 2 {
		t.Errorf("untapped lands at the end step = %d, want exactly 2 of 3", got)
	}
}

// --- Karn Liberated ---------------------------------------------

// TestKarnPlusFourMakesTheTargetDiscard — weaker than printed on
// purpose: the card lands in the graveyard rather than exile, and
// the choice is still the victim's.
func TestKarnPlusFourMakesTheTargetDiscard(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	them := g.Seats[(seat+1)%len(g.Seats)]
	karn := pushWalkerForTest(g, owner.ID, "Karn Liberated", karnLiberatedOracle, 6)
	advanceToMainOf(t, g, seat)

	err := g.ActivateCatalogAbility(owner.ID, karn, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: them.ID}},
	})
	if err != nil {
		t.Fatalf("+4: %v", err)
	}
	if got := loyaltyOf(g, karn); got != 10 {
		t.Errorf("loyalty after +4 = %d, want 10", got)
	}
	passPriorityAroundTable(t, g)

	if g.DiscardPending[them.ID] != 1 {
		t.Errorf("the target owes %d discards, want 1", g.DiscardPending[them.ID])
	}
	if g.DiscardPending[owner.ID] != 0 {
		t.Error("Karn's controller was asked to discard")
	}
}

func TestKarnMinusThreeExilesAnyPermanent(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	them := g.Seats[(seat+1)%len(g.Seats)]
	karn := pushWalkerForTest(g, owner.ID, "Karn Liberated", karnLiberatedOracle, 6)
	victim := pushCreatureToBattlefieldForTest(g, them.ID, "Victim")
	advanceToMainOf(t, g, seat)

	err := g.ActivateCatalogAbility(owner.ID, karn, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	})
	if err != nil {
		t.Fatalf("−3: %v", err)
	}
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(victim) {
		t.Error("the permanent was not exiled")
	}
	if !g.Exile.Contains(victim) {
		t.Error("the permanent is not in exile")
	}
}

// --- Saheeli Rai ------------------------------------------------

func TestSaheeliPlusOneScriesThenPingsEachOpponent(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	saheeli := pushWalkerForTest(g, owner.ID, "Saheeli Rai", saheeliRaiOracle, 3)
	advanceToMainOf(t, g, seat)

	life := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		life[p.ID] = p.Life
	}
	if err := g.ActivateCatalogAbility(owner.ID, saheeli, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	passPriorityAroundTable(t, g)

	// The scry queues a prompt; the damage is in its continuation,
	// so it lands once the scry is answered.
	resolveAnyScryFor(t, g, owner.ID)

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
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	saheeli := pushWalkerForTest(g, owner.ID, "Saheeli Rai", saheeliRaiOracle, 3)
	original := pushCreatureToBattlefieldForTest(g, owner.ID, "Bear")
	advanceToMainOf(t, g, seat)

	err := g.ActivateCatalogAbility(owner.ID, saheeli, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: original}},
	})
	if err != nil {
		t.Fatalf("−2: %v", err)
	}
	passPriorityAroundTable(t, g)

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

	advanceToStepOnThisTurn(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
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
		if got := saheeliArtifactInAddition(tc.in); got != tc.want {
			t.Errorf("saheeliArtifactInAddition(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestSaheeliDifferentNamesRejectsARepeat(t *testing.T) {
	sol := game.Card{Name: "Sol Ring"}
	mox := game.Card{Name: "Mox Diamond"}
	if !saheeliDifferentNames([]game.Card{sol, mox}) {
		t.Error("two different names were rejected")
	}
	if saheeliDifferentNames([]game.Card{sol, sol}) {
		t.Error("two copies of one name were accepted")
	}
	if !saheeliDifferentNames(nil) {
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
		if c.Kind == game.PendingChoiceSacrifice && c.Chooser == player {
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

// resolveAnyScryFor answers an open scry prompt for the player by
// putting everything back on top, which is what a test that does not
// care about the order wants.
func resolveAnyScryFor(t *testing.T, g *game.Game, player uuid.UUID) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c.Kind != game.PendingChoiceScry || c.Chooser != player {
			continue
		}
		if err := g.ResolveScry(c.ID, player, nil, c.ScryCards); err != nil {
			t.Fatalf("ResolveScry: %v", err)
		}
		return
	}
	t.Fatal("no scry prompt was queued")
}

// advanceToStepOnThisTurn walks forward to the named step without
// letting the turn roll over — the delayed triggers under test fire
// at "the next end step", which is this turn's.
func advanceToStepOnThisTurn(t *testing.T, g *game.Game, step game.Step) {
	t.Helper()
	for i := 0; i < 40; i++ {
		if g.Turn.Step == step {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep to %s: %v", step, err)
		}
	}
	t.Fatalf("never reached step %s (at %s)", step, g.Turn.Step)
}
