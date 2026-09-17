package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// doubling_season_test.go — the S17 sprint exit-criterion test.
//
// > Doubling Season + Hardened Scales in play → add a +1/+1 counter
// > → prompt for order → 4 counters land per chosen order.
//
// Both permutations exercised explicitly:
//   - [HS, DS] → (1 + 1) * 2 = 4
//   - [DS, HS] → (1 * 2) + 1 = 3
//
// Coverage also asserts the single-replacement fast path (Doubling
// Season alone → 2 counters, no prompt) and the predicate gates
// (opponent's creature, foreign counter name).

const (
	doublingSeasonOracle     = "01546b7d-a233-4176-8843-d732074dc5b6"
	hardenedScalesOracle     = "a1f3da21-af6d-450e-bf0b-985d158418e6"
	branchingEvolutionOracle = "28fe909b-06e0-424c-9f75-c824a25f5865"
)

// seedCreature pushes a 2/2 creature onto the battlefield via the
// timestamp helper (fires EventZoneMove so the S16 listener bumps
// layerVersion + stamps EnteredBattlefieldAt).
func seedCreature(g *game.Game, name string, controller uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       name,
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      controller,
		Controller: controller,
	})
}

// seedReplacementPermanent pushes a catalog card onto the
// battlefield so its Replacements become active. OracleID gates
// the CatalogReplacements lookup.
func seedReplacementPermanent(g *game.Game, oracleID, name string, controller uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       name,
		TypeLine:   "Enchantment",
		OracleID:   oracleID,
		Owner:      controller,
		Controller: controller,
	})
}

// countersOn returns the +1/+1 count on the given card. Zero when
// the counter map is nil or absent.
func countersOn(g *game.Game, cardID uuid.UUID, name string) int {
	var n int
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == cardID {
				n = c.Counters[name]
				return
			}
		}
	})
	return n
}

// TestDoublingSeasonAloneDoubles — single-replacement fast path.
// No prompt, the apply-loop runs to completion inline, and the
// counter doubles (1 → 2).
func TestDoublingSeasonAloneDoubles(t *testing.T) {
	g := newCatalogGame(t)
	p0 := g.Seats[0].ID

	_ = seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", p0)
	bear := seedCreature(g, "Bears", p0)

	if err := g.AddCounter(bear, "+1/+1", 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if got := countersOn(g, bear, "+1/+1"); got != 2 {
		t.Errorf("counters = %d, want 2 (1 original × 2 Doubling Season)", got)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("unexpected pending choice with only one replacement: %d", len(g.PendingChoices))
	}
}

// TestHardenedScalesAloneAddsOne — symmetric fast path for the
// narrower predicate. 1 → 2 via +1.
func TestHardenedScalesAloneAddsOne(t *testing.T) {
	g := newCatalogGame(t)
	p0 := g.Seats[0].ID

	_ = seedReplacementPermanent(g, hardenedScalesOracle, "Hardened Scales", p0)
	bear := seedCreature(g, "Bears", p0)

	if err := g.AddCounter(bear, "+1/+1", 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if got := countersOn(g, bear, "+1/+1"); got != 2 {
		t.Errorf("counters = %d, want 2 (1 original + 1 Hardened Scales)", got)
	}
}

// TestDoublingSeasonSkipsOpponentCreature — controller predicate
// gates the effect. Doubling Season on p0's board does NOT affect
// counters placed on p1's creature.
func TestDoublingSeasonSkipsOpponentCreature(t *testing.T) {
	g := newCatalogGame(t)
	p0 := g.Seats[0].ID
	p1 := g.Seats[1].ID

	_ = seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", p0)
	oppBear := seedCreature(g, "OpponentBear", p1)

	if err := g.AddCounter(oppBear, "+1/+1", 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if got := countersOn(g, oppBear, "+1/+1"); got != 1 {
		t.Errorf("counters = %d, want 1 (opponent's creature, Doubling Season skipped)", got)
	}
}

// TestHardenedScalesGatesOnCounterName — Hardened Scales only fires
// on +1/+1. A loyalty counter passes through untouched.
func TestHardenedScalesGatesOnCounterName(t *testing.T) {
	g := newCatalogGame(t)
	p0 := g.Seats[0].ID

	_ = seedReplacementPermanent(g, hardenedScalesOracle, "Hardened Scales", p0)
	pw := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Emperor",
		TypeLine:   "Legendary Planeswalker — Emperor",
		Owner:      p0,
		Controller: p0,
	})

	if err := g.AddCounter(pw, "loyalty", 3); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if got := countersOn(g, pw, "loyalty"); got != 3 {
		t.Errorf("loyalty = %d, want 3 (Hardened Scales doesn't touch loyalty)", got)
	}
}

// TestExitCriterionDoublingSeasonPlusHardenedScalesOrderMatters —
// the S17 sprint exit criterion. With both replacements in play,
// adding a +1/+1 queues a CR 616 prompt; the two possible orders
// produce different counter counts (4 vs 3).
func TestExitCriterionDoublingSeasonPlusHardenedScalesOrderMatters(t *testing.T) {
	t.Run("HS then DS → 4 counters", func(t *testing.T) {
		g := newCatalogGame(t)
		p0 := g.Seats[0].ID
		hsID := seedReplacementPermanent(g, hardenedScalesOracle, "Hardened Scales", p0)
		dsID := seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", p0)
		bear := seedCreature(g, "Bears", p0)

		if err := g.AddCounter(bear, "+1/+1", 1); err != nil {
			t.Fatalf("AddCounter: %v", err)
		}
		// Prompt queued; no counters placed yet.
		if got := countersOn(g, bear, "+1/+1"); got != 0 {
			t.Fatalf("counters = %d, want 0 (prompt pending)", got)
		}
		pc := g.PendingChoices
		if len(pc) != 1 {
			t.Fatalf("pending choices = %d, want 1", len(pc))
		}
		prompt := pc[0]
		if prompt.Kind != game.PendingChoiceReplacementOrder {
			t.Fatalf("prompt kind = %q, want %q", prompt.Kind, game.PendingChoiceReplacementOrder)
		}
		if prompt.Chooser != p0 {
			t.Errorf("chooser = %s, want %s", prompt.Chooser, p0)
		}

		// Find which ID corresponds to HS and DS.
		hsEffectID, dsEffectID := replacementIDsForSources(t, g, prompt.ReplacementEffectIDs, hsID, dsID)

		// Order: HS first, then DS. (1+1)*2 = 4.
		if err := g.ResolveReplacementOrder(prompt.ID, p0, []game.ReplacementEffectID{hsEffectID, dsEffectID}); err != nil {
			t.Fatalf("ResolveReplacementOrder: %v", err)
		}
		if got := countersOn(g, bear, "+1/+1"); got != 4 {
			t.Errorf("counters = %d, want 4 ((1+1)*2 with HS first)", got)
		}
	})

	t.Run("DS then HS → 3 counters", func(t *testing.T) {
		g := newCatalogGame(t)
		p0 := g.Seats[0].ID
		hsID := seedReplacementPermanent(g, hardenedScalesOracle, "Hardened Scales", p0)
		dsID := seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", p0)
		bear := seedCreature(g, "Bears", p0)

		if err := g.AddCounter(bear, "+1/+1", 1); err != nil {
			t.Fatalf("AddCounter: %v", err)
		}
		prompt := g.PendingChoices[0]
		hsEffectID, dsEffectID := replacementIDsForSources(t, g, prompt.ReplacementEffectIDs, hsID, dsID)

		// Order: DS first, then HS. (1*2)+1 = 3.
		if err := g.ResolveReplacementOrder(prompt.ID, p0, []game.ReplacementEffectID{dsEffectID, hsEffectID}); err != nil {
			t.Fatalf("ResolveReplacementOrder: %v", err)
		}
		if got := countersOn(g, bear, "+1/+1"); got != 3 {
			t.Errorf("counters = %d, want 3 ((1*2)+1 with DS first)", got)
		}
	})
}

// replacementIDsForSources maps a prompt's ReplacementEffectIDs
// back to (hsID, dsID) by reading each ID's metadata and matching
// on source card. Avoids hard-coding the engine's ID-minting
// scheme in the test.
func replacementIDsForSources(t *testing.T, g *game.Game, ids []game.ReplacementEffectID, hsCardID, dsCardID uuid.UUID) (hsEff, dsEff game.ReplacementEffectID) {
	t.Helper()
	return replacementIDForSource(t, g, ids, hsCardID), replacementIDForSource(t, g, ids, dsCardID)
}

// replacementIDForSource is the one-card half of the above: the
// prompt entry contributed by that permanent. Fatal when the prompt
// has none, which is always a test bug.
func replacementIDForSource(t *testing.T, g *game.Game, ids []game.ReplacementEffectID, cardID uuid.UUID) game.ReplacementEffectID {
	t.Helper()
	for _, id := range ids {
		if _, srcCardID := g.ReplacementOptionMetaForEffect(id); srcCardID == cardID {
			return id
		}
	}
	t.Fatalf("no prompt entry for source card %s", cardID)
	return 0
}

// TestThreeReplacementsSinglePrompt — with Doubling Season +
// Hardened Scales + Branching Evolution all in play, adding a
// +1/+1 counter should queue EXACTLY ONE prompt listing all 3,
// and submitting the order should apply all 3 in that order
// without re-prompting. Earlier drafts fired only the first pick
// and re-queued for the rest, forcing the user to submit the same
// order N times.
func TestThreeReplacementsSinglePrompt(t *testing.T) {
	g := newCatalogGame(t)
	p0 := g.Seats[0].ID

	hsID := seedReplacementPermanent(g, hardenedScalesOracle, "Hardened Scales", p0)
	beID := seedReplacementPermanent(g, branchingEvolutionOracle, "Branching Evolution", p0)
	dsID := seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", p0)
	bear := seedCreature(g, "Bears", p0)

	if err := g.AddCounter(bear, "+1/+1", 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want 1 prompt listing all 3 effects", len(g.PendingChoices))
	}
	prompt := g.PendingChoices[0]
	if len(prompt.ReplacementEffectIDs) != 3 {
		t.Fatalf("prompt has %d effect IDs, want 3", len(prompt.ReplacementEffectIDs))
	}

	// Find the three IDs by source card.
	var hsEff, beEff, dsEff game.ReplacementEffectID
	for _, id := range prompt.ReplacementEffectIDs {
		_, src := g.ReplacementOptionMetaForEffect(id)
		switch src {
		case hsID:
			hsEff = id
		case beID:
			beEff = id
		case dsID:
			dsEff = id
		}
	}

	// Order: HS → BE → DS. (1+1)*2*2 = 8.
	if err := g.ResolveReplacementOrder(prompt.ID, p0, []game.ReplacementEffectID{hsEff, beEff, dsEff}); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}

	// Critical regression: no second prompt queued.
	if len(g.PendingChoices) != 0 {
		t.Errorf("pending choices after resolve = %d, want 0 (single-submit should apply all 3)", len(g.PendingChoices))
	}
	if got := countersOn(g, bear, "+1/+1"); got != 8 {
		t.Errorf("counters = %d, want 8 ((1+1)*2*2 with HS→BE→DS)", got)
	}
}

// TestTwoDoublingSeasonsNeedNoPrompt — #792's headline case. Two
// copies of ONE card are two objects contributing one effect, so
// every order the CR 616 prompt could offer applies the same
// modification twice: 1 → 4. The engine skips the prompt and applies
// them inline, which also matters because #730's gate means an
// unanswered prompt holds up the whole table.
func TestTwoDoublingSeasonsNeedNoPrompt(t *testing.T) {
	g := newCatalogGame(t)
	p0 := g.Seats[0].ID

	_ = seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", p0)
	_ = seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", p0)
	bear := seedCreature(g, "Bears", p0)

	if err := g.AddCounter(bear, "+1/+1", 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if len(g.PendingChoices) != 0 {
		t.Fatalf("two copies of ONE replacement queued %d prompts, want 0 (#792)", len(g.PendingChoices))
	}
	if got := countersOn(g, bear, "+1/+1"); got != 4 {
		t.Errorf("counters = %d, want 4 (1 × 2 × 2, both Seasons applied)", got)
	}
}

// TestTwoHardenedScalesNeedNoPrompt — the additive half of the same
// rule, and the second card #792 names. Two Scales is +2, not a
// question.
func TestTwoHardenedScalesNeedNoPrompt(t *testing.T) {
	g := newCatalogGame(t)
	p0 := g.Seats[0].ID

	_ = seedReplacementPermanent(g, hardenedScalesOracle, "Hardened Scales", p0)
	_ = seedReplacementPermanent(g, hardenedScalesOracle, "Hardened Scales", p0)
	bear := seedCreature(g, "Bears", p0)

	if err := g.AddCounter(bear, "+1/+1", 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if len(g.PendingChoices) != 0 {
		t.Fatalf("two copies of ONE replacement queued %d prompts, want 0 (#792)", len(g.PendingChoices))
	}
	if got := countersOn(g, bear, "+1/+1"); got != 3 {
		t.Errorf("counters = %d, want 3 (1 + 1 + 1, both Scales applied)", got)
	}
}

// TestTwoDoublingSeasonsPlusHardenedScalesStillPrompts — the mixed
// window, and the reason #792 stops at "ALL of them are the same
// effect". Collapsing the two Seasons into one entry here would force
// them to fire back to back, and CR 616.1 lets the affected player
// interleave: Season, Scales, Season is 6 counters, which neither
// Season-Season-Scales (5) nor Scales-Season-Season (8) can reach. So
// the prompt still lists all three, and this checks that the answer
// the player could only reach by interleaving actually lands.
func TestTwoDoublingSeasonsPlusHardenedScalesStillPrompts(t *testing.T) {
	g := newCatalogGame(t)
	p0 := g.Seats[0].ID

	firstDS := seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", p0)
	hsID := seedReplacementPermanent(g, hardenedScalesOracle, "Hardened Scales", p0)
	secondDS := seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", p0)
	bear := seedCreature(g, "Bears", p0)

	if err := g.AddCounter(bear, "+1/+1", 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want 1 (a mixed window still prompts)", len(g.PendingChoices))
	}
	prompt := g.PendingChoices[0]
	if len(prompt.ReplacementEffectIDs) != 3 {
		t.Fatalf("prompt lists %d effects, want all 3", len(prompt.ReplacementEffectIDs))
	}
	// Season → Scales → Season: ((1 * 2) + 1) * 2 = 6.
	order := []game.ReplacementEffectID{
		replacementIDForSource(t, g, prompt.ReplacementEffectIDs, firstDS),
		replacementIDForSource(t, g, prompt.ReplacementEffectIDs, hsID),
		replacementIDForSource(t, g, prompt.ReplacementEffectIDs, secondDS),
	}
	if err := g.ResolveReplacementOrder(prompt.ID, p0, order); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
	if got := countersOn(g, bear, "+1/+1"); got != 6 {
		t.Errorf("counters = %d, want 6 (((1*2)+1)*2 — only reachable by interleaving)", got)
	}
}

// TestBranchingEvolutionStacksWithDoublingSeason — both are
// "creatures only, +1/+1 only" doublers. With both + a third
// counter on a creature, the prompt has 2 entries (same shape,
// different source cards), and either order produces 4.
//
// Two DIFFERENT cards, so #792 does not collapse them: the rules say
// the affected player chooses, and only an ordering nobody could
// observe is safe to skip. Commuting by arithmetic accident is not
// the same thing as being one effect.
func TestBranchingEvolutionStacksWithDoublingSeason(t *testing.T) {
	g := newCatalogGame(t)
	p0 := g.Seats[0].ID

	beID := seedReplacementPermanent(g, branchingEvolutionOracle, "Branching Evolution", p0)
	dsID := seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", p0)
	bear := seedCreature(g, "Bears", p0)

	if err := g.AddCounter(bear, "+1/+1", 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if len(g.PendingChoices) != 1 {
		t.Fatalf("prompt not queued")
	}
	prompt := g.PendingChoices[0]
	// Pick whichever order — both doublers multiply, so (1*2)*2 = 4 regardless.
	if err := g.ResolveReplacementOrder(prompt.ID, p0, prompt.ReplacementEffectIDs); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
	if got := countersOn(g, bear, "+1/+1"); got != 4 {
		t.Errorf("counters = %d, want 4 ((1*2)*2 from BE + DS)", got)
	}
	_ = beID
	_ = dsID
}
