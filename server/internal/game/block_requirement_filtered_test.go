package game

import (
	"testing"
)

// block_requirement_filtered_test.go pins #6087's "must be blocked by a
// Dalek if able" (CR 509.1c): one requirement on the attacker, obeyed only
// by a matching blocker; a non-matching blocker may block it freely; the
// search finds the matching blocker the declaration owes, and respects
// menace's minimum.

const filteredMustBeBlockedOracle = "test-filtered-must-be-blocked-6087"

// withDalekMustBeBlocked makes the card with filteredMustBeBlockedOracle
// carry "must be blocked by a Dalek if able".
func withDalekMustBeBlocked(t *testing.T) {
	t.Helper()
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID != filteredMustBeBlockedOracle {
			return nil
		}
		return []StaticAbility{{
			Layer: Layer6Ability,
			AppliesTo: func(target *Card, _ *Game, source *Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *Characteristic, _ *Card, _ *Game, source *Card) {
				c.BlockRequirements = append(c.BlockRequirements, BlockRequirement{
					Kind: BlockRequirementMustBeBlocked, Filter: BlockerFilterSubtype("Dalek"),
					Source: source.InstanceID, SourceName: source.Name,
				})
			},
		}}
	})
}

func TestMustBeBlockedByADalekNeedsADalek(t *testing.T) {
	withDalekMustBeBlocked(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := pushTyped(t, g, me, "Batter", "Creature — Human", filteredMustBeBlockedOracle, 3, 3)
	other := pushCombatant(t, g, me, "Other Bear", 2, 2)
	dalek := pushTyped(t, g, opp, "Dalek Drone", "Artifact Creature — Dalek", "", 2, 2)
	bear := pushTyped(t, g, opp, "Grizzly Bears", "Creature — Bear", "", 2, 2)
	declareAttacks(t, g, atk, other)

	var owed []BlockDeclaration
	g.WithWriteLock(func() { owed = g.blockRequirementWitnessLocked(opp.ID) })
	if len(owed) != 1 || owed[0].Blocker != dalek || owed[0].Attacker != atk {
		t.Fatalf("witness = %+v, want the Dalek on the Batter", owed)
	}
	// The Dalek sent elsewhere gives the requirement up.
	br := blockRequirementErr(t, block(t, g, dalek, other))
	if got := br.Sentence(opp.ID); got != "Batter must be blocked by a Dalek if able." {
		t.Errorf("sentence = %q", got)
	}
	// A non-Dalek may block it, and obeys nothing.
	if err := block(t, g, bear, atk); err != nil {
		t.Fatalf("a non-Dalek blocking it: %v", err)
	}
	passToDefender(t, g)
	blockRequirementErr(t, g.PassPriority())
	blockRequirementErr(t, g.FinishBlocks(opp.ID))
	if err := block(t, g, dalek, atk); err != nil {
		t.Fatalf("the Dalek joins: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with a Dalek blocking: %v", err)
	}
}

func TestMustBeBlockedByADalekWithoutADalekAsksNothing(t *testing.T) {
	withDalekMustBeBlocked(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := pushTyped(t, g, me, "Batter", "Creature — Human", filteredMustBeBlockedOracle, 3, 3)
	pushTyped(t, g, opp, "Grizzly Bears", "Creature — Bear", "", 2, 2)
	declareAttacks(t, g, atk)
	passToDefender(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("no Dalek, so nothing is owed: %v", err)
	}
}

func TestMustBeBlockedByADalekCountsAChangeling(t *testing.T) {
	withDalekMustBeBlocked(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := pushTyped(t, g, me, "Batter", "Creature — Human", filteredMustBeBlockedOracle, 3, 3)
	shape := pushTyped(t, g, opp, "Shapeshifter", "Creature — Shapeshifter", "", 1, 1, "changeling")
	declareAttacks(t, g, atk)
	var owed []BlockDeclaration
	g.WithWriteLock(func() { owed = g.blockRequirementWitnessLocked(opp.ID) })
	if len(owed) != 1 || owed[0].Blocker != shape {
		t.Fatalf("witness = %+v, want the changeling (every creature type)", owed)
	}
}

func TestMustBeBlockedByADalekRespectsMenace(t *testing.T) {
	withDalekMustBeBlocked(t)
	t.Run("a lone Dalek owes nothing", func(t *testing.T) {
		g := newActiveGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		atk := pushTyped(t, g, me, "Batter", "Creature — Human", filteredMustBeBlockedOracle, 3, 3, "menace")
		pushTyped(t, g, opp, "Dalek Drone", "Artifact Creature — Dalek", "", 2, 2)
		declareAttacks(t, g, atk)
		if got := g.BlockDeclarationStatusOf(opp.ID); got != BlockDeclarationDeclared {
			t.Fatalf("status = %q: a lone blocker can't block a menace attacker", got)
		}
	})
	t.Run("a Dalek and a Bear owe the pair", func(t *testing.T) {
		g := newActiveGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		atk := pushTyped(t, g, me, "Batter", "Creature — Human", filteredMustBeBlockedOracle, 3, 3, "menace")
		dalek := pushTyped(t, g, opp, "Dalek Drone", "Artifact Creature — Dalek", "", 2, 2)
		bear := pushTyped(t, g, opp, "Grizzly Bears", "Creature — Bear", "", 2, 2)
		declareAttacks(t, g, atk)
		var owed []BlockDeclaration
		g.WithWriteLock(func() { owed = g.blockRequirementWitnessLocked(opp.ID) })
		if len(owed) != 2 {
			t.Fatalf("witness = %+v, want the Dalek and the Bear", owed)
		}
		if err := block(t, g, dalek, atk, bear, atk); err != nil {
			t.Fatalf("the pair: %v", err)
		}
		passToDefender(t, g)
		if err := g.PassPriority(); err != nil {
			t.Fatalf("pass with the pair blocking: %v", err)
		}
	})
}

func TestBlockerFilterSubtypeKeys(t *testing.T) {
	if !KnownBlockerFilter(BlockerFilterSubtype("Dalek")) {
		t.Error("a subtype key is not known")
	}
	if got := BlockerFilterNoun(BlockerFilterSubtype("Elf")); got != "an Elf" {
		t.Errorf("noun = %q", got)
	}
	if got := BlockerFilterNoun(BlockerFilterSubtype("Dalek")); got != "a Dalek" {
		t.Errorf("noun = %q", got)
	}
}
