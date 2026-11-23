package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Kaya's Ghostform ---------------------------------------------

func TestKayasGhostformReturnsTheCreatureWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.AuraBear(g, me.ID)
	aura := effectstest.AuraCast(t, g, "Kaya's Ghostform", kayasGhostformOracle, bear)

	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.NamedOnBattlefield(g, "Bear", me.ID); n != 1 {
		t.Fatalf("%d Bears on the battlefield, want the one that came back", n)
	}
	if g.Battlefield.Contains(aura) {
		t.Error("the Aura should have gone to the graveyard with its host")
	}
}

func TestKayasGhostformReturnsTheCreatureWhenItIsExiled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.AuraBear(g, me.ID)
	effectstest.AuraCast(t, g, "Kaya's Ghostform", kayasGhostformOracle, bear)

	effectstest.ExileCreature(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.NamedOnBattlefield(g, "Bear", me.ID); n != 1 {
		t.Errorf("%d Bears on the battlefield after the exile, want 1", n)
	}
	if g.Exile.Contains(bear) {
		t.Error("the Bear is still in exile")
	}
}

func TestKayasGhostformIgnoresABounce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.AuraBear(g, me.ID)
	effectstest.AuraCast(t, g, "Kaya's Ghostform", kayasGhostformOracle, bear)

	g.WithWriteLock(func() { g.BounceCardsToHandForEffect([]uuid.UUID{bear}) })
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.NamedOnBattlefield(g, "Bear", me.ID); n != 0 {
		t.Errorf("a bounced Bear came back to the battlefield (%d)", n)
	}
}

// --- Gift of Immortality ------------------------------------------

func TestGiftOfImmortalityReturnsTheCreatureNowAndTheAuraAtEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.AuraBear(g, me.ID)
	aura := effectstest.AuraCast(t, g, "Gift of Immortality", giftOfImmortalityOracle, bear)

	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.NamedOnBattlefield(g, "Bear", me.ID); n != 1 {
		t.Fatalf("%d Bears on the battlefield, want the one that came back", n)
	}
	if g.Battlefield.Contains(aura) {
		t.Fatal("the Aura came back at once; it waits for the end step")
	}

	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(aura) {
		t.Fatal("the Aura did not return at the beginning of the next end step")
	}
	host := effectstest.AttachmentHostOf(t, g, aura)
	ids := effectstest.BattlefieldIDsNamed(g, "Bear")
	if len(ids) != 1 || host.Kind != game.TargetCard || host.ID != ids[0] {
		t.Errorf("the Aura is attached to %+v, want the returned Bear %v", host, ids)
	}
}

// "Under its owner's control": a Gift cast on an opponent's creature
// gives it back to that opponent.
func TestGiftOfImmortalityReturnsTheCreatureToItsOwner(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.AuraBear(g, opp.ID)
	effectstest.AuraCast(t, g, "Gift of Immortality", giftOfImmortalityOracle, theirs)

	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.NamedOnBattlefield(g, "Bear", opp.ID); n != 1 {
		t.Errorf("their Bear did not come back under their control (%d)", n)
	}
	if n := effectstest.NamedOnBattlefield(g, "Bear", me.ID); n != 0 {
		t.Errorf("the Bear came back under the Aura's controller (%d)", n)
	}
}

// If the creature is gone again by the end step there is nothing for
// the Aura to be attached to, and it stays in the graveyard.
func TestGiftOfImmortalityStaysInTheGraveyardIfTheCreatureLeftAgain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.AuraBear(g, me.ID)
	aura := effectstest.AuraCast(t, g, "Gift of Immortality", giftOfImmortalityOracle, bear)

	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	returned := effectstest.BattlefieldIDsNamed(g, "Bear")
	if len(returned) != 1 {
		t.Fatalf("%d Bears returned, want 1", len(returned))
	}
	effectstest.ExileCreature(t, g, returned[0])

	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(aura) {
		t.Error("the Aura returned although its creature is in exile")
	}
}

// --- Sheltered by Ghosts ------------------------------------------

func TestShelteredByGhostsExilesUntilItLeavesAndGrantsLifelinkAndWard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.AuraBear(g, me.ID)
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Rock", TypeLine: "Artifact",
		Owner: opp.ID, Controller: opp.ID,
	})
	land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Forest", TypeLine: "Basic Land — Forest",
		Owner: opp.ID, Controller: opp.ID,
	})
	aura := effectstest.CastCatalogSpell(t, g, "Sheltered by Ghosts", effectstest.AuraTypeLine, shelteredByGhostsOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if effectstest.HasID(p.PickTargetCards, land) {
			t.Error("a land is offered as a target for \"nonland permanent\"")
		}
		if !effectstest.HasID(p.PickTargetCards, rock) {
			t.Error("the opponent's artifact is not offered")
		}
	}
	effectstest.PickCard(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Exile.Contains(rock) {
		t.Fatal("the targeted permanent was not exiled")
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 3 {
		t.Errorf("power %d, want 3", p)
	}
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, bear), "lifelink") {
		t.Errorf("abilities %v lack lifelink", effectstest.EffectiveAbilities(t, g, bear))
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aura) })
	g.RunStateChecksForTest()
	if effectstest.NamedOnBattlefield(g, "Their Rock", opp.ID) != 1 {
		t.Error("the exiled permanent did not return under its owner's control when the Aura left")
	}
}

// CR 610.3b: an Aura that is gone before its trigger resolves never
// starts the exile, so nothing is taken and nothing is left to return.
func TestShelteredByGhostsDoesNothingIfTheAuraLeavesFirst(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.AuraBear(g, me.ID)
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Rock", TypeLine: "Artifact",
		Owner: opp.ID, Controller: opp.ID,
	})
	aura := effectstest.CastCatalogSpell(t, g, "Sheltered by Ghosts", effectstest.AuraTypeLine, shelteredByGhostsOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, rock)
	if effectstest.TriggerOnStack(g, aura) == nil {
		t.Fatal("the exile trigger is not on the stack")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aura) })
	effectstest.PassPriorityAroundTable(t, g)

	if g.Exile.Contains(rock) || !g.Battlefield.Contains(rock) {
		t.Error("the permanent was exiled by an Aura that had already left")
	}
}

// --- Mantle of the Ancients ---------------------------------------

func TestMantleOfTheAncientsReturnsAurasAndEquipmentAttached(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.AuraBear(g, me.ID)
	pac := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: pac, Name: "Pacifism", TypeLine: effectstest.AuraTypeLine, OracleID: effectstest.PacifismAuraOracle,
		Owner: me.ID, Controller: me.ID,
	})
	hammer := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: hammer, Name: "Loxodon Warhammer", TypeLine: effectstest.EquipTypeLine, OracleID: effectstest.WarhammerOracle,
		Owner: me.ID, Controller: me.ID,
	})
	growth := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: growth, Name: "Wild Growth", TypeLine: effectstest.AuraTypeLine, OracleID: effectstest.WildGrowthAuraOracle,
		Owner: me.ID, Controller: me.ID,
	})

	mantle := effectstest.CastCatalogSpell(t, g, "Mantle of the Ancients", effectstest.AuraTypeLine, mantleOfTheAncientsOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerTriggerTargets(t, g, me.ID, pac, hammer, growth)
	effectstest.PassPriorityAroundTable(t, g)

	for name, id := range map[string]uuid.UUID{"Pacifism": pac, "Loxodon Warhammer": hammer} {
		if host := effectstest.AttachmentHostOf(t, g, id); host.ID != bear {
			t.Errorf("%s is attached to %+v, want the enchanted Bear", name, host)
		}
	}
	if g.Battlefield.Contains(growth) {
		t.Error("Wild Growth came back although it can't enchant a creature")
	}
	if host := effectstest.AttachmentHostOf(t, g, mantle); host.ID != bear {
		t.Fatalf("the Mantle is attached to %+v", host)
	}
	// Mantle +1/+1, Pacifism +1/+1 and the Warhammer +1/+1 for three
	// attachments, then the Warhammer's own +3/+0.
	if p, tt := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 8 || tt != 5 {
		t.Errorf("enchanted Bear is %d/%d, want 8/5", p, tt)
	}
}

// --- Mechanized Production ----------------------------------------

// mechanizedProductionBoard puts the Aura on an artifact of the
// second seat, so the upkeep that matters is the next one to arrive.
func mechanizedProductionBoard(g *game.Game, extraCopies int) (owner *game.Player, widget uuid.UUID) {
	owner = g.Seats[1]
	widget = effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Widget", TypeLine: "Artifact",
		Owner: owner.ID, Controller: owner.ID,
	})
	for i := 0; i < extraCopies; i++ {
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Widget", TypeLine: "Artifact",
			Owner: owner.ID, Controller: owner.ID,
		})
	}
	aura := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Mechanized Production", TypeLine: effectstest.AuraTypeLine,
		OracleID: mechanizedProductionOracle, Owner: owner.ID, Controller: owner.ID,
	})
	g.WithWriteLock(func() {
		_ = g.AttachForEffect(aura, game.TargetRef{Kind: game.TargetCard, ID: widget})
	})
	return owner, widget
}

func TestMechanizedProductionCopiesTheEnchantedArtifactEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, _ := mechanizedProductionBoard(g, 0)

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.NamedOnBattlefield(g, "Widget", owner.ID); n != 2 {
		t.Errorf("%d Widgets after one upkeep, want 2", n)
	}
	if g.State != game.StateActive {
		t.Error("the game ended with only two Widgets")
	}
}

func TestMechanizedProductionWinsAtEightArtifactsWithTheSameName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, _ := mechanizedProductionBoard(g, 6) // seven Widgets; the token is the eighth

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Outcome == nil || g.Outcome.Winner != owner.ID || g.Outcome.Cause != game.OutcomeCauseEffect {
		t.Errorf("outcome %+v, want a win by effect for the Aura's controller", g.Outcome)
	}
}

func TestMechanizedProductionDoesNotCountArtifactsWithDifferentNames(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, _ := mechanizedProductionBoard(g, 2) // three Widgets, four with the token
	for i := 0; i < 5; i++ {
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Gadget " + string(rune('A'+i)), TypeLine: "Artifact",
			Owner: owner.ID, Controller: owner.ID,
		})
	}

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)

	if g.State != game.StateActive {
		t.Error("eight artifacts of different names won the game")
	}
}
