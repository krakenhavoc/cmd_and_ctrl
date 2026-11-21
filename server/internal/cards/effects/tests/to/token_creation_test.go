package teststo

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestParallelLivesDoublesOneCreationInstruction — CR 701.7b: "create
// two Treasures" is ONE event, so Parallel Lives makes four, in one
// batch, with no prompt.
func TestParallelLivesDoublesOneCreationInstruction(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0].ID
	effectstest.PushTokenReplacementCard(g, effectstest.ParallelLivesOracle, "Parallel Lives", "Enchantment", me)

	effectstest.MakeTokens(t, g, me, effects.TreasureToken(), 2)

	if len(g.PendingChoices) != 0 {
		t.Fatalf("one doubler queued %d prompt(s), want none", len(g.PendingChoices))
	}
	if got := effectstest.OnBattlefieldNamed(g, "Treasure"); got != 4 {
		t.Errorf("Treasures = %d, want 4 (two doubled)", got)
	}
}

// TestTwoAnointedProcessionsQuadrupleWithNoPrompt — #792: two copies
// of one declared effect are not an ordering question. Four tokens,
// nobody asked anything.
func TestTwoAnointedProcessionsQuadrupleWithNoPrompt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0].ID
	effectstest.PushTokenReplacementCard(g, effectstest.AnointedProcessionOracl, "Anointed Procession", "Enchantment", me)
	effectstest.PushTokenReplacementCard(g, effectstest.AnointedProcessionOracl, "Anointed Procession", "Enchantment", me)

	effectstest.MakeTokens(t, g, me, effects.TreasureToken(), 1)

	if len(g.PendingChoices) != 0 {
		t.Fatalf("two Anointed Processions queued %d prompt(s), want none", len(g.PendingChoices))
	}
	if got := effectstest.OnBattlefieldNamed(g, "Treasure"); got != 4 {
		t.Errorf("Treasures = %d, want 4 (×2 then ×2)", got)
	}
}

// TestMondrakDoublesYourTokensOnly — the fourth printing of the same
// effect, and the controller clause that keeps it off the opponent's
// Saprolings.
func TestMondrakDoublesYourTokensOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	effectstest.PushTokenReplacementCard(g, effectstest.MondrakOracle, "Mondrak, Glory Dominus", "Legendary Creature — Phyrexian Horror", me)

	effectstest.MakeTokens(t, g, me, effects.TreasureToken(), 1)
	effectstest.MakeTokens(t, g, opp, effects.ClueToken(), 1)

	if got := effectstest.OnBattlefieldNamed(g, "Treasure"); got != 2 {
		t.Errorf("your Treasures = %d, want 2", got)
	}
	if got := effectstest.OnBattlefieldNamed(g, "Clue"); got != 1 {
		t.Errorf("an opponent's Clues = %d, want 1 (Mondrak is not symmetrical)", got)
	}
}

// TestPrimalVigorDoublesEverybodysTokens — the symmetrical one, which
// is the whole reason it is not simply a worse Doubling Season.
func TestPrimalVigorDoublesEverybodysTokens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	effectstest.PushTokenReplacementCard(g, effectstest.PrimalVigorOracle, "Primal Vigor", "Enchantment", me)

	effectstest.MakeTokens(t, g, me, effects.TreasureToken(), 1)
	effectstest.MakeTokens(t, g, opp, effects.ClueToken(), 1)

	if got := effectstest.OnBattlefieldNamed(g, "Treasure"); got != 2 {
		t.Errorf("your Treasures = %d, want 2", got)
	}
	if got := effectstest.OnBattlefieldNamed(g, "Clue"); got != 2 {
		t.Errorf("an opponent's Clues = %d, want 2 — Primal Vigor doubles for everyone", got)
	}
}

// TestAcademyManufactorTurnsOneKindIntoThree — the card that decides
// the event carries GROUPS rather than a count. One Treasure becomes a
// Clue, a Food and a Treasure; the count rides along, so two Treasures
// become two of each.
func TestAcademyManufactorTurnsOneKindIntoThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0].ID
	effectstest.PushTokenReplacementCard(g, effectstest.AcademyManufactorOracle, "Academy Manufactor", "Artifact Creature — Assembly-Worker", me)

	effectstest.MakeTokens(t, g, me, effects.TreasureToken(), 2)

	for _, name := range []string{"Clue", "Food", "Treasure"} {
		if got := effectstest.OnBattlefieldNamed(g, name); got != 2 {
			t.Errorf("%ss = %d, want 2 (two of each)", name, got)
		}
	}
}

// TestAcademyManufactorLeavesOtherKindsAlone — "if you would create a
// Clue, Food, or Treasure token" is a clause about the token, not
// about the instruction: a creation it does not name is untouched.
func TestAcademyManufactorLeavesOtherKindsAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0].ID
	effectstest.PushTokenReplacementCard(g, effectstest.AcademyManufactorOracle, "Academy Manufactor", "Artifact Creature — Assembly-Worker", me)

	effectstest.MakeTokens(t, g, me, effects.RedGoblinToken(), 1)

	if got := effectstest.OnBattlefieldNamed(g, "Goblin"); got != 1 {
		t.Errorf("Goblins = %d, want 1", got)
	}
	if got := effectstest.OnBattlefieldNamed(g, "Clue"); got != 0 {
		t.Errorf("Clues = %d, want 0 — a Goblin is not a Clue, Food or Treasure", got)
	}
}

// TestDoublingSeasonAndAcademyManufactorAskForTheOrder — two DIFFERENT
// declared effects in one window, so CR 616 gives the creation's
// controller the ordering, and the two orders really do produce
// different boards: doubling first makes two Treasures and then two of
// each kind; rewriting the kinds first makes one of each and then
// doubles all three.
func TestDoublingSeasonAndAcademyManufactorAskForTheOrder(t *testing.T) {
	seed := func(t *testing.T) (*game.Game, uuid.UUID, uuid.UUID, uuid.UUID) {
		t.Helper()
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0].ID
		ds := effectstest.PushTokenReplacementCard(g, effectstest.DoublingSeasonOracle, "Doubling Season", "Enchantment", me)
		am := effectstest.PushTokenReplacementCard(g, effectstest.AcademyManufactorOracle, "Academy Manufactor", "Artifact Creature — Assembly-Worker", me)
		return g, me, ds, am
	}
	assertTwoOfEach := func(t *testing.T, g *game.Game) {
		t.Helper()
		for _, name := range []string{"Clue", "Food", "Treasure"} {
			if got := effectstest.OnBattlefieldNamed(g, name); got != 2 {
				t.Errorf("%ss = %d, want 2", name, got)
			}
		}
	}

	t.Run("Doubling Season first", func(t *testing.T) {
		g, me, ds, am := seed(t)
		effectstest.MakeTokens(t, g, me, effects.TreasureToken(), 1)
		prompt := g.PendingChoices[0]
		if prompt.Kind != game.PendingChoiceReplacementOrder || prompt.Chooser != me {
			t.Fatalf("prompt = %q for %s, want a replacement order for the creator", prompt.Kind, prompt.Chooser)
		}
		first := effectstest.ReplacementIDForSource(t, g, prompt.ReplacementEffectIDs, ds)
		second := effectstest.ReplacementIDForSource(t, g, prompt.ReplacementEffectIDs, am)
		if err := g.ResolveReplacementOrder(prompt.ID, me, []game.ReplacementEffectID{first, second}); err != nil {
			t.Fatalf("ResolveReplacementOrder: %v", err)
		}
		assertTwoOfEach(t, g)
	})

	t.Run("Academy Manufactor first", func(t *testing.T) {
		g, me, ds, am := seed(t)
		effectstest.MakeTokens(t, g, me, effects.TreasureToken(), 1)
		prompt := g.PendingChoices[0]
		first := effectstest.ReplacementIDForSource(t, g, prompt.ReplacementEffectIDs, am)
		second := effectstest.ReplacementIDForSource(t, g, prompt.ReplacementEffectIDs, ds)
		if err := g.ResolveReplacementOrder(prompt.ID, me, []game.ReplacementEffectID{first, second}); err != nil {
			t.Fatalf("ResolveReplacementOrder: %v", err)
		}
		assertTwoOfEach(t, g)
	})
}

// TestDoublingSeasonStillDoublesCounters — the half that shipped in
// S17 must not have moved when the token half landed beside it.
func TestDoublingSeasonStillDoublesCountersWithTheTokenHalfPresent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0].ID
	effectstest.PushTokenReplacementCard(g, effectstest.DoublingSeasonOracle, "Doubling Season", "Enchantment", me)
	bear := effectstest.SeedCreature(g, "Bears", me)

	if err := g.AddCounter(bear, "+1/+1", 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if got := effectstest.CountersOn(g, bear, "+1/+1"); got != 2 {
		t.Errorf("counters = %d, want 2", got)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("the counter half prompted: %d choice(s)", len(g.PendingChoices))
	}
}

// TestATokenCopyRunsTheCopiedCardsAsEnters — #762 put a created token
// on the shared entry path, and fireETBHookLocked is on that path, so
// a token copy now gets the copied card's CR 614.12 "as this enters"
// clause. It never did before, and token_copy.go said so in its own
// comment.
func TestATokenCopyRunsTheCopiedCardsAsEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0].ID
	original := effectstest.PushTokenReplacementCard(g, "53c730c6-2f8c-4af8-b400-b9d573a71e60",
		"Adaptive Automaton", "Artifact Creature — Construct", me)

	g.WithWriteLock(func() {
		tmpl, ok := effects.TokenCopyTemplate(g, original)
		if !ok {
			t.Fatal("TokenCopyTemplate failed")
		}
		if err := g.CreateTokenForEffect(me, tmpl, 1); err != nil {
			t.Fatalf("CreateTokenForEffect: %v", err)
		}
	})

	if len(g.PendingChoices) == 0 {
		t.Fatal("the token copy did not run Adaptive Automaton's as-enters clause")
	}
}
