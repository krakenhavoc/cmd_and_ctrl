package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const theScarabGodOracle = "c75e55f2-fd6c-4816-9d96-21eeb8369aff"

func pushScarabGod(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Scarab God", TypeLine: "Legendary Creature — God",
		OracleID: theScarabGodOracle, Power: 5, Toughness: 5, Owner: owner, Controller: owner,
	})
}

// TestScarabGodUpkeepDrainsAndScriesByZombieCount.
func TestScarabGodUpkeepDrainsAndScriesByZombieCount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "Bottom Me", "Bottom Too")
	pushScarabGod(g, me.ID)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Zombie One", TypeLine: "Creature — Zombie",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Zombie Two", TypeLine: "Creature — Zombie",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	oppLife := opp.Life

	effectstest.AdvanceToUpkeepOf(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)

	if opp.Life != oppLife-2 {
		t.Errorf("each opponent should lose X=2 life: %d, want %d", opp.Life, oppLife-2)
	}
	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no scry-2 prompt: %+v", g.PendingChoices)
	}
	if len(c.ScryCards) != 2 {
		t.Errorf("scry X=2 should offer two cards: %d", len(c.ScryCards))
	}
}

// TestScarabGodUpkeepDoesNothingWithNoZombies.
func TestScarabGodUpkeepDoesNothingWithNoZombies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushScarabGod(g, me.ID)
	oppLife := opp.Life

	effectstest.AdvanceToUpkeepOf(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)

	if opp.Life != oppLife {
		t.Errorf("X=0 should drain nothing: %d, want %d", opp.Life, oppLife)
	}
	if c := effectstest.ScryChoiceFor(g, me.ID); c != nil {
		t.Error("X=0 should not queue a scry prompt")
	}
}

// TestScarabGodReanimatesAsAFourFourBlackZombie exercises the
// activated ability: exile a creature card from a graveyard, create a
// token copy that's a 4/4 black Zombie instead of its printed types.
func TestScarabGodReanimatesAsAFourFourBlackZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	scarab := pushScarabGod(g, me.ID)
	target := effectstest.GraveCreature(opp, "Their Dead Bear", "{1}{G}")

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, scarab, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if opp.Graveyard.Contains(target) {
		t.Error("the copied creature card should have been exiled")
	}
	var token *game.Card
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == me.ID && c.Name == "Their Dead Bear" {
			token = c
		}
	}
	if token == nil {
		t.Fatalf("no Zombie token copy on the battlefield: %+v", g.Battlefield.Cards)
	}
	if token.Power != 4 || token.Toughness != 4 {
		t.Errorf("the token should be a 4/4: %d/%d", token.Power, token.Toughness)
	}
	if len(token.Colors) != 1 || token.Colors[0] != "B" {
		t.Errorf("the token should be black: %v", token.Colors)
	}
	if got := effectstest.EffectiveSubtypes(t, g, token.InstanceID); !effectstest.ContainsStr(got, "Zombie") || effectstest.ContainsStr(got, "Bear") {
		t.Errorf("the token's creature type should be REPLACED with Zombie: %v", got)
	}
}

// TestScarabGodReturnsToHandAtTheNextEndStepWhenItDies.
func TestScarabGodReturnsToHandAtTheNextEndStepWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	scarab := pushScarabGod(g, me.ID)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(scarab) })
	effectstest.PassPriorityAroundTable(t, g)

	if !me.Graveyard.Contains(scarab) {
		t.Fatal("The Scarab God should be in the graveyard right after dying")
	}

	effectstest.AdvanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)

	if me.Graveyard.Contains(scarab) {
		t.Error("The Scarab God should have returned to its owner's hand by the next end step")
	}
	if !me.Hand.Contains(scarab) {
		t.Error("The Scarab God should be in its owner's hand")
	}
}
