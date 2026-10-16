package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// planeswalker_gaps_3861_test.go — the clauses #3861 closed: Ugin's
// "put up to seven permanent cards from your hand", Teferi, Temporal
// Pilgrim's growing Spirit and −12, and Wrenn and Realmbreaker's −7
// emblem. Karn Liberated's +4 is in superfriends_batch2_test.go beside
// his other abilities.

// --- Ugin, the Spirit Dragon −10 ---------------------------------

// After the life and the seven cards, one pick over the permanent cards
// in hand: optional, at most seven, nonpermanents not offered, and the
// picks enter without being played (a land is not a land drop).
func TestUginMinusTenPutsUpToSevenPermanentsFromHand(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	ugin := pushWalkerForTest(g, owner.ID, "Ugin, the Spirit Dragon", uginSpiritDragonOracle, 12)
	advanceToMainOf(t, g, seat)
	bear := pushHandCardWithManaCost(owner, "Hand Bear", "Creature — Bear", "{1}{G}")
	forest := pushHandCardWithManaCost(owner, "Hand Forest", "Basic Land — Forest", "")
	bolt := pushHandCardWithManaCost(owner, "Hand Bolt", "Instant", "{R}")
	life := owner.Life

	if err := g.ActivateCatalogAbility(owner.ID, ugin, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−10: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := owner.Life - life; got != 7 {
		t.Errorf("life gained = %d, want 7", got)
	}

	pick := chooseCardsChoiceFor(g, owner.ID)
	if pick == nil {
		t.Fatal("no pick over the hand after the draw")
	}
	if pick.ChooseMin != 0 {
		t.Errorf("pick floor = %d, want 0 (\"up to seven\")", pick.ChooseMin)
	}
	if pick.ChooseMax > 7 {
		t.Errorf("pick ceiling = %d, want at most 7", pick.ChooseMax)
	}
	if !hasID(pick.ChooseCards, bear) || !hasID(pick.ChooseCards, forest) {
		t.Errorf("the permanent cards in hand are not offered: %v", pick.ChooseCards)
	}
	if hasID(pick.ChooseCards, bolt) {
		t.Error("an instant was offered to put onto the battlefield")
	}
	if err := g.ResolveChooseCards(pick.ID, owner.ID, []uuid.UUID{bear, forest}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !g.Battlefield.Contains(bear) || !g.Battlefield.Contains(forest) {
		t.Error("the picked cards did not reach the battlefield")
	}
	if !owner.Hand.Contains(bolt) {
		t.Error("the instant left the hand")
	}
	if got := g.LandsPlayedThisTurnFor(owner.ID); got != 0 {
		t.Errorf("land plays this turn = %d, want 0 (putting a land is not playing it)", got)
	}
}

// --- Teferi, Temporal Pilgrim ------------------------------------

// The −2's Spirit carries its own draw trigger: each of its
// controller's draws puts a +1/+1 counter on it, and an opponent's
// draw does not.
func TestTeferiPilgrimSpiritGrowsOnYourDraws(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	them := g.Seats[(seat+1)%len(g.Seats)]
	teferi := pushWalkerForTest(g, owner.ID, "Teferi, Temporal Pilgrim", teferiTemporalPilgrimOracle, 4)
	advanceToMainOf(t, g, seat)

	if err := g.ActivateCatalogAbility(owner.ID, teferi, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−2: %v", err)
	}
	passPriorityAroundTable(t, g)
	spirit := uuid.Nil
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Spirit" && c.Controller == owner.ID {
			spirit = c.InstanceID
			if c.TokenKey != game.TokenKey("spirit-temporal-pilgrim") {
				t.Errorf("Spirit token key = %q, want the catalog template's", c.TokenKey)
			}
		}
	}
	if spirit == uuid.Nil {
		t.Fatal("no Spirit token was created")
	}

	if err := g.DrawNForEffect(them.ID, 1); err != nil {
		t.Fatalf("opponent draw: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := plusOneCounters(g, spirit); got != 0 {
		t.Errorf("+1/+1 counters after an opponent's draw = %d, want 0", got)
	}

	if err := g.DrawNForEffect(owner.ID, 2); err != nil {
		t.Fatalf("draw: %v", err)
	}
	// Teferi's own draw trigger fires beside the Spirit's, so the
	// owner orders them (CR 603.3b) before the stack can drain.
	for i := 0; i < 8; i++ {
		passPriorityAroundTable(t, g)
		order := openTriggerOrder(g, owner.ID)
		if order == nil {
			break
		}
		if err := g.ResolveTriggerOrder(order.ID, owner.ID, append([]uuid.UUID(nil), order.TriggerOrderIDs...)); err != nil {
			t.Fatalf("ResolveTriggerOrder: %v", err)
		}
	}
	if got := plusOneCounters(g, spirit); got != 2 {
		t.Errorf("+1/+1 counters after drawing two = %d, want 2", got)
	}
	if got := effectivePowerWithCounters(t, g, spirit); got != 4 {
		t.Errorf("Spirit power = %d, want 4", got)
	}
}

// The −12: the target opponent picks a permanent they control and it
// goes to its owner's hand; then every nonland permanent they still
// control goes into its OWNER's library, and lands stay.
func TestTeferiPilgrimMinusTwelveBouncesThenShufflesNonlandAway(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	them := g.Seats[(seat+1)%len(g.Seats)]
	// 13, so he survives paying 12 and the test can tell him from a
	// permanent the −12 moved.
	teferi := pushWalkerForTest(g, owner.ID, "Teferi, Temporal Pilgrim", teferiTemporalPilgrimOracle, 13)
	advanceToMainOf(t, g, seat)
	land := b12Permanent(g, them.ID, "Their Land", "Land")
	bear := b12Permanent(g, them.ID, "Their Bear", "Creature — Bear")
	rock := b12Permanent(g, them.ID, "Their Rock", "Artifact")
	// A creature of mine they have stolen: it goes into MY library.
	stolen := b12Permanent(g, owner.ID, "My Stolen Bear", "Creature — Bear")
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == stolen {
			g.Battlefield.Cards[i].Controller = them.ID
		}
	}
	mine := b12Permanent(g, owner.ID, "My Rock", "Artifact")

	if err := g.ActivateCatalogAbility(owner.ID, teferi, 2, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: owner.ID}},
	}); err == nil {
		t.Fatal("the −12 targeted its own controller; it says target opponent")
	}
	if err := g.ActivateCatalogAbility(owner.ID, teferi, 2, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: them.ID}},
	}); err != nil {
		t.Fatalf("−12: %v", err)
	}
	passPriorityAroundTable(t, g)

	if chooseCardsChoiceFor(g, owner.ID) != nil {
		t.Error("Teferi's controller was asked to choose; the opponent chooses")
	}
	pick := chooseCardsChoiceFor(g, them.ID)
	if pick == nil {
		t.Fatal("the opponent was not asked to choose a permanent")
	}
	if pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Errorf("pick bounds = [%d, %d], want exactly one", pick.ChooseMin, pick.ChooseMax)
	}
	for _, id := range []uuid.UUID{land, bear, rock, stolen} {
		if !hasID(pick.ChooseCards, id) {
			t.Errorf("a permanent they control is not offered: %v", id)
		}
	}
	if hasID(pick.ChooseCards, mine) || hasID(pick.ChooseCards, teferi) {
		t.Error("a permanent they do not control was offered")
	}
	themLib, ownerLib := them.Library.Size(), owner.Library.Size()
	if err := g.ResolveChooseCards(pick.ID, them.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}

	if !them.Hand.Contains(bear) {
		t.Error("the chosen permanent did not return to its owner's hand")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("a land they did not choose left the battlefield")
	}
	if !them.Library.Contains(rock) {
		t.Error("their nonland permanent was not shuffled into their library")
	}
	if !owner.Library.Contains(stolen) {
		t.Error("the stolen creature was not shuffled into its OWNER's library")
	}
	if got := them.Library.Size() - themLib; got != 1 {
		t.Errorf("their library grew by %d, want 1", got)
	}
	if got := owner.Library.Size() - ownerLib; got != 1 {
		t.Errorf("my library grew by %d, want 1", got)
	}
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(teferi) {
		t.Error("a permanent the opponent does not control was moved")
	}
}

// --- Wrenn and Realmbreaker −7 -----------------------------------

// The emblem lets its owner play lands and cast permanent spells from
// their own graveyard, and nothing else: not an instant, and not an
// opponent's graveyard.
func TestWrennAndRealmbreakerEmblemPlaysLandsAndCastsPermanentsFromGraveyard(t *testing.T) {
	spec, ok := Lookup(wrennAndRealmbreakerOracle)
	if !ok {
		t.Fatal("Wrenn and Realmbreaker is not registered")
	}
	if len(spec.Activated) != 3 || spec.Emblem == nil {
		t.Fatalf("Wrenn declares %d abilities and emblem %v, want 3 and the emblem", len(spec.Activated), spec.Emblem)
	}

	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	wrenn := pushWalkerForTest(g, owner.ID, "Wrenn and Realmbreaker", wrennAndRealmbreakerOracle, 7)
	advanceToMainOf(t, g, seat)
	land := graveyardCardOf(g, owner, "Dead Forest", "Basic Land — Forest", "", "")
	bear := graveyardCardOf(g, owner, "Dead Bear", "Creature — Bear", "{1}{G}", "")
	bolt := graveyardCardOf(g, owner, "Dead Bolt", "Instant", "{R}", "")
	theirs := graveyardCardOf(g, opp, "Their Dead Bear", "Creature — Bear", "{1}{G}", "")

	if grantedPermissionOn(g, owner.ID, bear, game.ZoneGraveyard).Granted() {
		t.Fatal("a graveyard cast was open before the emblem existed")
	}
	if err := g.ActivateCatalogAbility(owner.ID, wrenn, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−7: %v", err)
	}
	passPriorityAroundTable(t, g)
	// Wrenn paid all seven and is gone (CR 704.5i); the emblem stays.
	emblems := emblemsOf(g, owner.ID)
	if len(emblems) != 1 || emblems[0].Text != "You may play lands and cast permanent spells from your graveyard." {
		t.Fatalf("emblems after the −7 = %+v", emblems)
	}

	if grantedPermissionOn(g, owner.ID, bolt, game.ZoneGraveyard).Granted() {
		t.Error("the emblem opened an instant in the graveyard")
	}
	if grantedPermissionOn(g, owner.ID, theirs, game.ZoneGraveyard).Granted() {
		t.Error("the emblem opened an opponent's graveyard")
	}

	if err := g.CastSpell(owner.ID, land, game.CastSpellParams{FromZone: string(game.ZoneGraveyard)}); err != nil {
		t.Fatalf("play the land from the graveyard: %v", err)
	}
	if !g.Battlefield.Contains(land) {
		t.Error("the land never reached the battlefield")
	}
	if got := g.LandsPlayedThisTurnFor(owner.ID); got != 1 {
		t.Errorf("land plays this turn = %d, want 1 (the land drop is spent)", got)
	}

	payMana(t, g, owner, "{G}{G}")
	if err := g.CastSpell(owner.ID, bear, game.CastSpellParams{FromZone: string(game.ZoneGraveyard)}); err != nil {
		t.Fatalf("cast the creature from the graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the creature cast from the graveyard did not resolve onto the battlefield")
	}
}
