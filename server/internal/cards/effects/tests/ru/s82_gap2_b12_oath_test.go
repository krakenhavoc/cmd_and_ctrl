package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap2_b12_oath_test.go — S82 GAP2 batch 12: Oath of Druids, through
// the real upkeep trigger, the upkeep player's target pick and the reveal.

const gap2B12Oath = "cf0dcb27-f02e-4065-81dc-75fbbb09ea3f"

func TestGap2B12OathOfDruidsIsFull(t *testing.T) {
	spec, ok := effects.Lookup(gap2B12Oath)
	if !ok {
		t.Fatal("Oath of Druids is not registered")
	}
	if spec.Name != "Oath of Druids" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
	}
}

// The upkeep player (not the Oath's controller) picks an opponent of theirs
// with more creatures, then reveals until a creature card.
func TestGap2B12OathOfDruidsUpkeepPlayerPicksAndReveals(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, upkeep, rich, poor := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	effectstest.PushPermanentForTest(g, owner.ID, "Oath of Druids", gap2B12Oath, "Enchantment")
	effectstest.PushVanillaCreature(g, upkeep.ID, "Lone Elf", 1, 1)
	effectstest.PushVanillaCreature(g, rich.ID, "Bear A", 2, 2)
	effectstest.PushVanillaCreature(g, rich.ID, "Bear B", 2, 2)
	effectstest.PushVanillaCreature(g, owner.ID, "Owner Bear A", 2, 2)
	effectstest.PushVanillaCreature(g, owner.ID, "Owner Bear B", 2, 2)
	_ = poor

	creature := effectstest.PutOnTopOfLibrary(g, upkeep, "Craw Wurm", "Creature — Wurm", "")
	land1 := effectstest.PutOnTopOfLibrary(g, upkeep, "Forest", "Basic Land — Forest", "")
	land2 := effectstest.PutOnTopOfLibrary(g, upkeep, "Island", "Basic Land — Island", "")

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	ch := effectstest.PickTargetPrompt(t, g)
	if ch.Chooser != upkeep.ID {
		t.Fatalf("the prompt is for %s, want the upkeep player", ch.Chooser)
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range ch.PickTargetPlayers {
		offered[id] = true
	}
	if !offered[rich.ID] || !offered[owner.ID] || offered[upkeep.ID] || offered[poor.ID] || len(offered) != 2 {
		t.Fatalf("offered %v, want only the opponents with more creatures than the upkeep player", ch.PickTargetPlayers)
	}
	if err := g.ResolvePickTarget(ch.ID, upkeep.ID, game.TargetRef{Kind: game.TargetPlayer, ID: poor.ID}); err == nil {
		t.Fatal("a player with fewer creatures was accepted")
	}
	if err := g.ResolvePickTarget(ch.ID, upkeep.ID, game.TargetRef{Kind: game.TargetPlayer, ID: rich.ID}); err != nil {
		t.Fatalf("picking the richer opponent: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerConfirmFor(t, g, upkeep, true)
	if n := effectstest.NamedOnBattlefield(g, "Craw Wurm", upkeep.ID); n != 1 {
		t.Errorf("Craw Wurm: %d under the upkeep player's control, want 1", n)
	}
	for _, id := range []uuid.UUID{land1, land2} {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard || z.Owner != upkeep.ID {
			t.Errorf("a revealed land is in %+v, want its owner's graveyard", z)
		}
	}
	_ = creature
}

// Nobody with more creatures: the trigger never goes on the stack.
func TestGap2B12OathOfDruidsDoesNothingWithoutARicherOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, upkeep := g.Seats[0], g.Seats[1]
	effectstest.PushPermanentForTest(g, owner.ID, "Oath of Druids", gap2B12Oath, "Enchantment")
	effectstest.PushVanillaCreature(g, upkeep.ID, "Lone Elf", 1, 1)
	effectstest.PushVanillaCreature(g, owner.ID, "Owner Bear", 2, 2)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if p := effectstest.LatestPickTarget(g, upkeep.ID); p != nil {
		t.Fatalf("the upkeep player was asked for a target with nobody eligible: %+v", p)
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceConfirm {
			t.Fatalf("a reveal was offered: %+v", c)
		}
	}
}

// Declining reveals nothing; the controller is never asked.
func TestGap2B12OathOfDruidsTheUpkeepPlayerMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, upkeep := g.Seats[0], g.Seats[1]
	effectstest.PushPermanentForTest(g, owner.ID, "Oath of Druids", gap2B12Oath, "Enchantment")
	effectstest.PushVanillaCreature(g, owner.ID, "Owner Bear", 2, 2)
	top := effectstest.PutOnTopOfLibrary(g, upkeep, "Craw Wurm", "Creature — Wurm", "")
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	ch := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, upkeep.ID, game.TargetRef{Kind: game.TargetPlayer, ID: owner.ID}); err != nil {
		t.Fatalf("picking: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceConfirm && c.Chooser == owner.ID {
			t.Fatal("the Oath's controller was asked to reveal")
		}
	}
	effectstest.AnswerConfirmFor(t, g, upkeep, false)
	if z := g.FindCardZoneForEffect(top); z == nil || z.Kind != game.ZoneLibrary {
		t.Errorf("the top card moved to %+v after a decline", z)
	}
}
