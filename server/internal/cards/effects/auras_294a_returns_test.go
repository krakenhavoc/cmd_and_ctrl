package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// auras_294a_returns_test.go — the slice 294-a Auras that move cards
// between zones: Kaya's Ghostform, Gift of Immortality, Sheltered by
// Ghosts, Mantle of the Ancients, and Mechanized Production's token.

const (
	wildGrowthAuraOracle = "706ae742-1807-44b7-a4fa-f2e26f61519a"
	pacifismAuraOracle   = "5f5e0b10-c8cf-450c-bfd3-bcb0528ec330"
)

// namedOnBattlefield counts battlefield cards with the name under a
// controller.
func namedOnBattlefield(g *game.Game, name string, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == controller {
			n++
		}
	}
	return n
}

// killCreature moves a battlefield creature to its owner's graveyard
// through the public mutator, which runs the state-based actions on
// the way out — so a fallen-off Aura is swept before any trigger
// goes on the stack, as it is in play.
func killCreature(t *testing.T, g *game.Game, owner, id uuid.UUID) {
	t.Helper()
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield},
		game.ZoneRef{Kind: game.ZoneGraveyard, Owner: owner}, id); err != nil {
		t.Fatalf("MoveCardByID to the graveyard: %v", err)
	}
}

// exileCreature is killCreature for exile.
func exileCreature(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield},
		game.ZoneRef{Kind: game.ZoneExile}, id); err != nil {
		t.Fatalf("MoveCardByID to exile: %v", err)
	}
}

// --- Kaya's Ghostform ---------------------------------------------

func TestKayasGhostformReturnsTheCreatureWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := auraBear(g, me.ID)
	aura := auraCast(t, g, "Kaya's Ghostform", kayasGhostformOracle, bear)

	killCreature(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)

	if n := namedOnBattlefield(g, "Bear", me.ID); n != 1 {
		t.Fatalf("%d Bears on the battlefield, want the one that came back", n)
	}
	if g.Battlefield.Contains(aura) {
		t.Error("the Aura should have gone to the graveyard with its host")
	}
}

func TestKayasGhostformReturnsTheCreatureWhenItIsExiled(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := auraBear(g, me.ID)
	auraCast(t, g, "Kaya's Ghostform", kayasGhostformOracle, bear)

	exileCreature(t, g, bear)
	passPriorityAroundTable(t, g)

	if n := namedOnBattlefield(g, "Bear", me.ID); n != 1 {
		t.Errorf("%d Bears on the battlefield after the exile, want 1", n)
	}
	if g.Exile.Contains(bear) {
		t.Error("the Bear is still in exile")
	}
}

func TestKayasGhostformIgnoresABounce(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := auraBear(g, me.ID)
	auraCast(t, g, "Kaya's Ghostform", kayasGhostformOracle, bear)

	g.WithWriteLock(func() { g.BounceCardsToHandForEffect([]uuid.UUID{bear}) })
	passPriorityAroundTable(t, g)

	if n := namedOnBattlefield(g, "Bear", me.ID); n != 0 {
		t.Errorf("a bounced Bear came back to the battlefield (%d)", n)
	}
}

// --- Gift of Immortality ------------------------------------------

func TestGiftOfImmortalityReturnsTheCreatureNowAndTheAuraAtEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := auraBear(g, me.ID)
	aura := auraCast(t, g, "Gift of Immortality", giftOfImmortalityOracle, bear)

	killCreature(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)

	if n := namedOnBattlefield(g, "Bear", me.ID); n != 1 {
		t.Fatalf("%d Bears on the battlefield, want the one that came back", n)
	}
	if g.Battlefield.Contains(aura) {
		t.Fatal("the Aura came back at once; it waits for the end step")
	}

	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(aura) {
		t.Fatal("the Aura did not return at the beginning of the next end step")
	}
	host := attachmentHostOf(t, g, aura)
	ids := battlefieldIDsNamed(g, "Bear")
	if len(ids) != 1 || host.Kind != game.TargetCard || host.ID != ids[0] {
		t.Errorf("the Aura is attached to %+v, want the returned Bear %v", host, ids)
	}
}

// "Under its owner's control": a Gift cast on an opponent's creature
// gives it back to that opponent.
func TestGiftOfImmortalityReturnsTheCreatureToItsOwner(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := auraBear(g, opp.ID)
	auraCast(t, g, "Gift of Immortality", giftOfImmortalityOracle, theirs)

	killCreature(t, g, opp.ID, theirs)
	passPriorityAroundTable(t, g)

	if n := namedOnBattlefield(g, "Bear", opp.ID); n != 1 {
		t.Errorf("their Bear did not come back under their control (%d)", n)
	}
	if n := namedOnBattlefield(g, "Bear", me.ID); n != 0 {
		t.Errorf("the Bear came back under the Aura's controller (%d)", n)
	}
}

// If the creature is gone again by the end step there is nothing for
// the Aura to be attached to, and it stays in the graveyard.
func TestGiftOfImmortalityStaysInTheGraveyardIfTheCreatureLeftAgain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := auraBear(g, me.ID)
	aura := auraCast(t, g, "Gift of Immortality", giftOfImmortalityOracle, bear)

	killCreature(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	returned := battlefieldIDsNamed(g, "Bear")
	if len(returned) != 1 {
		t.Fatalf("%d Bears returned, want 1", len(returned))
	}
	exileCreature(t, g, returned[0])

	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(aura) {
		t.Error("the Aura returned although its creature is in exile")
	}
}

// --- Sheltered by Ghosts ------------------------------------------

func TestShelteredByGhostsExilesUntilItLeavesAndGrantsLifelinkAndWard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := auraBear(g, me.ID)
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Rock", TypeLine: "Artifact",
		Owner: opp.ID, Controller: opp.ID,
	})
	land := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Forest", TypeLine: "Basic Land — Forest",
		Owner: opp.ID, Controller: opp.ID,
	})
	aura := castCatalogSpell(t, g, "Sheltered by Ghosts", auraTypeLine, shelteredByGhostsOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		if hasID(p.PickTargetCards, land) {
			t.Error("a land is offered as a target for \"nonland permanent\"")
		}
		if !hasID(p.PickTargetCards, rock) {
			t.Error("the opponent's artifact is not offered")
		}
	}
	pickCard(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)

	if !g.Exile.Contains(rock) {
		t.Fatal("the targeted permanent was not exiled")
	}
	if p := effectivePower(t, g, bear); p != 3 {
		t.Errorf("power %d, want 3", p)
	}
	if !containsString(effectiveAbilities(t, g, bear), "lifelink") {
		t.Errorf("abilities %v lack lifelink", effectiveAbilities(t, g, bear))
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aura) })
	g.RunStateChecksForTest()
	if namedOnBattlefield(g, "Their Rock", opp.ID) != 1 {
		t.Error("the exiled permanent did not return under its owner's control when the Aura left")
	}
}

// CR 610.3b: an Aura that is gone before its trigger resolves never
// starts the exile, so nothing is taken and nothing is left to return.
func TestShelteredByGhostsDoesNothingIfTheAuraLeavesFirst(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := auraBear(g, me.ID)
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Rock", TypeLine: "Artifact",
		Owner: opp.ID, Controller: opp.ID,
	})
	aura := castCatalogSpell(t, g, "Sheltered by Ghosts", auraTypeLine, shelteredByGhostsOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, rock)
	if triggerOnStack(g, aura) == nil {
		t.Fatal("the exile trigger is not on the stack")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aura) })
	passPriorityAroundTable(t, g)

	if g.Exile.Contains(rock) || !g.Battlefield.Contains(rock) {
		t.Error("the permanent was exiled by an Aura that had already left")
	}
}

// --- Mantle of the Ancients ---------------------------------------

func TestMantleOfTheAncientsReturnsAurasAndEquipmentAttached(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := auraBear(g, me.ID)
	pac := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: pac, Name: "Pacifism", TypeLine: auraTypeLine, OracleID: pacifismAuraOracle,
		Owner: me.ID, Controller: me.ID,
	})
	hammer := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: hammer, Name: "Loxodon Warhammer", TypeLine: equipTypeLine, OracleID: warhammerOracle,
		Owner: me.ID, Controller: me.ID,
	})
	growth := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: growth, Name: "Wild Growth", TypeLine: auraTypeLine, OracleID: wildGrowthAuraOracle,
		Owner: me.ID, Controller: me.ID,
	})

	mantle := castCatalogSpell(t, g, "Mantle of the Ancients", auraTypeLine, mantleOfTheAncientsOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	answerTriggerTargets(t, g, me.ID, pac, hammer, growth)
	passPriorityAroundTable(t, g)

	for name, id := range map[string]uuid.UUID{"Pacifism": pac, "Loxodon Warhammer": hammer} {
		if host := attachmentHostOf(t, g, id); host.ID != bear {
			t.Errorf("%s is attached to %+v, want the enchanted Bear", name, host)
		}
	}
	if g.Battlefield.Contains(growth) {
		t.Error("Wild Growth came back although it can't enchant a creature")
	}
	if host := attachmentHostOf(t, g, mantle); host.ID != bear {
		t.Fatalf("the Mantle is attached to %+v", host)
	}
	// Mantle +1/+1, Pacifism +1/+1 and the Warhammer +1/+1 for three
	// attachments, then the Warhammer's own +3/+0.
	if p, tt := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 8 || tt != 5 {
		t.Errorf("enchanted Bear is %d/%d, want 8/5", p, tt)
	}
}

// --- Mechanized Production ----------------------------------------

// mechanizedProductionBoard puts the Aura on an artifact of the
// second seat, so the upkeep that matters is the next one to arrive.
func mechanizedProductionBoard(g *game.Game, extraCopies int) (owner *game.Player, widget uuid.UUID) {
	owner = g.Seats[1]
	widget = pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Widget", TypeLine: "Artifact",
		Owner: owner.ID, Controller: owner.ID,
	})
	for i := 0; i < extraCopies; i++ {
		pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Widget", TypeLine: "Artifact",
			Owner: owner.ID, Controller: owner.ID,
		})
	}
	aura := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Mechanized Production", TypeLine: auraTypeLine,
		OracleID: mechanizedProductionOracle, Owner: owner.ID, Controller: owner.ID,
	})
	g.WithWriteLock(func() {
		_ = g.AttachForEffect(aura, game.TargetRef{Kind: game.TargetCard, ID: widget})
	})
	return owner, widget
}

func TestMechanizedProductionCopiesTheEnchantedArtifactEachUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	owner, _ := mechanizedProductionBoard(g, 0)

	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)

	if n := namedOnBattlefield(g, "Widget", owner.ID); n != 2 {
		t.Errorf("%d Widgets after one upkeep, want 2", n)
	}
	if g.State != game.StateActive {
		t.Error("the game ended with only two Widgets")
	}
}

func TestMechanizedProductionWinsAtEightArtifactsWithTheSameName(t *testing.T) {
	g := newCatalogGame(t)
	owner, _ := mechanizedProductionBoard(g, 6) // seven Widgets; the token is the eighth

	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)

	if g.Outcome == nil || g.Outcome.Winner != owner.ID || g.Outcome.Cause != game.OutcomeCauseEffect {
		t.Errorf("outcome %+v, want a win by effect for the Aura's controller", g.Outcome)
	}
}

func TestMechanizedProductionDoesNotCountArtifactsWithDifferentNames(t *testing.T) {
	g := newCatalogGame(t)
	owner, _ := mechanizedProductionBoard(g, 2) // three Widgets, four with the token
	for i := 0; i < 5; i++ {
		pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Gadget " + string(rune('A'+i)), TypeLine: "Artifact",
			Owner: owner.ID, Controller: owner.ID,
		})
	}

	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)

	if g.State != game.StateActive {
		t.Error("eight artifacts of different names won the game")
	}
}
