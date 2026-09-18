package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// token_creation_test.go — the catalog half of #762. The engine's own
// notion of a creation event is pinned in game/token_create_test.go;
// what is here is the printed cards, and the boards two of them make
// together.

const (
	parallelLivesOracle     = "84dc94b2-95fb-4d53-aaa2-191cb645639f"
	anointedProcessionOracl = "7246d45b-2185-4cdd-981b-5419b7d52bce"
	academyManufactorOracle = "f36d1d8b-8303-44a9-ab56-531931641ea2"
	mondrakOracle           = "fe83087d-c6c1-40be-9295-baaa1c6b2db1"
	primalVigorOracle       = "c665544f-557b-4631-a1dc-39571470ca2e"
)

// pushTokenReplacementCard puts a catalog permanent on the battlefield
// so its Replacements become live. The type line is the printed one so
// nothing downstream has to pretend.
func pushTokenReplacementCard(g *game.Game, oracleID, name, typeLine string, controller uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   oracleID,
		Owner:      controller,
		Controller: controller,
	})
}

// makeTokens runs one creation instruction and returns with the
// pipeline settled (or paused, which the caller then checks).
func makeTokens(t *testing.T, g *game.Game, controller uuid.UUID, template game.Card, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(controller, template, n); err != nil {
			t.Fatalf("CreateTokenForEffect: %v", err)
		}
	})
}

func onBattlefieldNamed(g *game.Game, name string) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name {
				n++
			}
		}
	})
	return n
}

// TestParallelLivesDoublesOneCreationInstruction — CR 701.7b: "create
// two Treasures" is ONE event, so Parallel Lives makes four, in one
// batch, with no prompt.
func TestParallelLivesDoublesOneCreationInstruction(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	pushTokenReplacementCard(g, parallelLivesOracle, "Parallel Lives", "Enchantment", me)

	makeTokens(t, g, me, TreasureToken(), 2)

	if len(g.PendingChoices) != 0 {
		t.Fatalf("one doubler queued %d prompt(s), want none", len(g.PendingChoices))
	}
	if got := onBattlefieldNamed(g, "Treasure"); got != 4 {
		t.Errorf("Treasures = %d, want 4 (two doubled)", got)
	}
}

// TestTwoAnointedProcessionsQuadrupleWithNoPrompt — #792: two copies
// of one declared effect are not an ordering question. Four tokens,
// nobody asked anything.
func TestTwoAnointedProcessionsQuadrupleWithNoPrompt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	pushTokenReplacementCard(g, anointedProcessionOracl, "Anointed Procession", "Enchantment", me)
	pushTokenReplacementCard(g, anointedProcessionOracl, "Anointed Procession", "Enchantment", me)

	makeTokens(t, g, me, TreasureToken(), 1)

	if len(g.PendingChoices) != 0 {
		t.Fatalf("two Anointed Processions queued %d prompt(s), want none", len(g.PendingChoices))
	}
	if got := onBattlefieldNamed(g, "Treasure"); got != 4 {
		t.Errorf("Treasures = %d, want 4 (×2 then ×2)", got)
	}
}

// TestMondrakDoublesYourTokensOnly — the fourth printing of the same
// effect, and the controller clause that keeps it off the opponent's
// Saprolings.
func TestMondrakDoublesYourTokensOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	pushTokenReplacementCard(g, mondrakOracle, "Mondrak, Glory Dominus", "Legendary Creature — Phyrexian Horror", me)

	makeTokens(t, g, me, TreasureToken(), 1)
	makeTokens(t, g, opp, ClueToken(), 1)

	if got := onBattlefieldNamed(g, "Treasure"); got != 2 {
		t.Errorf("your Treasures = %d, want 2", got)
	}
	if got := onBattlefieldNamed(g, "Clue"); got != 1 {
		t.Errorf("an opponent's Clues = %d, want 1 (Mondrak is not symmetrical)", got)
	}
}

// TestPrimalVigorDoublesEverybodysTokens — the symmetrical one, which
// is the whole reason it is not simply a worse Doubling Season.
func TestPrimalVigorDoublesEverybodysTokens(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	pushTokenReplacementCard(g, primalVigorOracle, "Primal Vigor", "Enchantment", me)

	makeTokens(t, g, me, TreasureToken(), 1)
	makeTokens(t, g, opp, ClueToken(), 1)

	if got := onBattlefieldNamed(g, "Treasure"); got != 2 {
		t.Errorf("your Treasures = %d, want 2", got)
	}
	if got := onBattlefieldNamed(g, "Clue"); got != 2 {
		t.Errorf("an opponent's Clues = %d, want 2 — Primal Vigor doubles for everyone", got)
	}
}

// TestAcademyManufactorTurnsOneKindIntoThree — the card that decides
// the event carries GROUPS rather than a count. One Treasure becomes a
// Clue, a Food and a Treasure; the count rides along, so two Treasures
// become two of each.
func TestAcademyManufactorTurnsOneKindIntoThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	pushTokenReplacementCard(g, academyManufactorOracle, "Academy Manufactor", "Artifact Creature — Assembly-Worker", me)

	makeTokens(t, g, me, TreasureToken(), 2)

	for _, name := range []string{"Clue", "Food", "Treasure"} {
		if got := onBattlefieldNamed(g, name); got != 2 {
			t.Errorf("%ss = %d, want 2 (two of each)", name, got)
		}
	}
}

// TestAcademyManufactorLeavesOtherKindsAlone — "if you would create a
// Clue, Food, or Treasure token" is a clause about the token, not
// about the instruction: a creation it does not name is untouched.
func TestAcademyManufactorLeavesOtherKindsAlone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	pushTokenReplacementCard(g, academyManufactorOracle, "Academy Manufactor", "Artifact Creature — Assembly-Worker", me)

	makeTokens(t, g, me, RedGoblinToken(), 1)

	if got := onBattlefieldNamed(g, "Goblin"); got != 1 {
		t.Errorf("Goblins = %d, want 1", got)
	}
	if got := onBattlefieldNamed(g, "Clue"); got != 0 {
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
		g := newCatalogGame(t)
		me := g.Seats[0].ID
		ds := pushTokenReplacementCard(g, doublingSeasonOracle, "Doubling Season", "Enchantment", me)
		am := pushTokenReplacementCard(g, academyManufactorOracle, "Academy Manufactor", "Artifact Creature — Assembly-Worker", me)
		return g, me, ds, am
	}
	assertTwoOfEach := func(t *testing.T, g *game.Game) {
		t.Helper()
		for _, name := range []string{"Clue", "Food", "Treasure"} {
			if got := onBattlefieldNamed(g, name); got != 2 {
				t.Errorf("%ss = %d, want 2", name, got)
			}
		}
	}

	t.Run("Doubling Season first", func(t *testing.T) {
		g, me, ds, am := seed(t)
		makeTokens(t, g, me, TreasureToken(), 1)
		prompt := g.PendingChoices[0]
		if prompt.Kind != game.PendingChoiceReplacementOrder || prompt.Chooser != me {
			t.Fatalf("prompt = %q for %s, want a replacement order for the creator", prompt.Kind, prompt.Chooser)
		}
		first := replacementIDForSource(t, g, prompt.ReplacementEffectIDs, ds)
		second := replacementIDForSource(t, g, prompt.ReplacementEffectIDs, am)
		if err := g.ResolveReplacementOrder(prompt.ID, me, []game.ReplacementEffectID{first, second}); err != nil {
			t.Fatalf("ResolveReplacementOrder: %v", err)
		}
		assertTwoOfEach(t, g)
	})

	t.Run("Academy Manufactor first", func(t *testing.T) {
		g, me, ds, am := seed(t)
		makeTokens(t, g, me, TreasureToken(), 1)
		prompt := g.PendingChoices[0]
		first := replacementIDForSource(t, g, prompt.ReplacementEffectIDs, am)
		second := replacementIDForSource(t, g, prompt.ReplacementEffectIDs, ds)
		if err := g.ResolveReplacementOrder(prompt.ID, me, []game.ReplacementEffectID{first, second}); err != nil {
			t.Fatalf("ResolveReplacementOrder: %v", err)
		}
		assertTwoOfEach(t, g)
	})
}

// TestDoublingSeasonStillDoublesCounters — the half that shipped in
// S17 must not have moved when the token half landed beside it.
func TestDoublingSeasonStillDoublesCountersWithTheTokenHalfPresent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	pushTokenReplacementCard(g, doublingSeasonOracle, "Doubling Season", "Enchantment", me)
	bear := seedCreature(g, "Bears", me)

	if err := g.AddCounter(bear, "+1/+1", 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if got := countersOn(g, bear, "+1/+1"); got != 2 {
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
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	original := pushTokenReplacementCard(g, "53c730c6-2f8c-4af8-b400-b9d573a71e60",
		"Adaptive Automaton", "Artifact Creature — Construct", me)

	g.WithWriteLock(func() {
		tmpl, ok := TokenCopyTemplate(g, original)
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
