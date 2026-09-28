package game

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// block_requirements_1684_test.go pins #1684's two additions to the
// CR 509.1c block requirements (ADR 0045 amendment of 2026-09-28,
// Decision 59): the fifth kind, "blocks THAT attacker if able", which
// names one attacking object (Provoke, Grappling Hook, Turntimber
// Basilisk), and the filtered Lure, which binds only the blockers a
// registered filter key matches (Marble Priest's Walls, Talruum
// Piper's flyers). The cards are pinned in cards/effects.

// objectRefOf is the attacking object `id` is right now.
func objectRefOf(t *testing.T, g *Game, id uuid.UUID) ObjectRef {
	t.Helper()
	var ref ObjectRef
	g.WithWriteLock(func() {
		c := findBattlefieldCard(g, id)
		if c == nil {
			t.Fatalf("%s is not on the battlefield", id)
		}
		ref = ObjectRef{ID: c.InstanceID, Epoch: c.ObjectEpoch}
	})
	return ref
}

// provoke registers "blocker blocks attacker this turn if able" as the
// data record Provoke's resolution writes.
func provoke(t *testing.T, g *Game, blocker uuid.UUID, attacker ObjectRef) {
	t.Helper()
	registerScopedEffectForTest(t, g, blocker, []Mod{BlocksAttackerMod(attacker)}, IndefiniteDuration())
}

// TestBlocksAttackerBindsOnlyThatAttacker — a provoked creature must
// block the provoker, and no other attacker, if able. Blocking the
// other attacker is refused at the verb, declaring nothing is refused
// at the checkpoint, and blocking the provoker is accepted.
func TestBlocksAttackerBindsOnlyThatAttacker(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	provoker := pushCombatant(t, g, me, "Goblin Grappler", 1, 1)
	other := pushCombatant(t, g, me, "Other Bear", 2, 2)
	bear := pushCombatant(t, g, opp, "Grizzly Bears", 2, 2)
	free := pushCombatant(t, g, opp, "Free Wall", 0, 4)
	provoke(t, g, bear, objectRefOf(t, g, provoker))
	declareAttacks(t, g, provoker, other)

	br := blockRequirementErr(t, block(t, g, bear, other))
	if br.Blocker != bear || br.Attacker != provoker {
		t.Errorf("refusal names %s -> %s, want the provoked Bears -> the provoker", br.Blocker, br.Attacker)
	}
	if br.Requirement.Kind != BlockRequirementBlocksAttacker {
		t.Errorf("requirement kind = %q, want blocksAttacker", br.Requirement.Kind)
	}
	if got := br.Sentence(opp.ID); got != "Grizzly Bears must block Goblin Grappler if able." {
		t.Errorf("sentence = %q", got)
	}
	// The free creature blocking the other attacker is fine — the
	// requirement is on the Bears alone.
	if err := block(t, g, free, other); err != nil {
		t.Fatalf("an unprovoked creature blocking elsewhere: %v", err)
	}
	var owed map[uuid.UUID]uuid.UUID
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked(); owed = g.MustBlockForEffect() })
	if len(owed) != 1 || owed[bear] != provoker {
		t.Errorf("must_block = %v, want exactly the Bears on the provoker", owed)
	}
	passToDefender(t, g)
	br = blockRequirementErr(t, g.PassPriority())
	if br.Blocker != bear || br.Attacker != provoker {
		t.Errorf("the pass's refusal names %s -> %s, want the Bears -> the provoker", br.Blocker, br.Attacker)
	}
	if err := block(t, g, bear, provoker); err != nil {
		t.Fatalf("the provoked block: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the provoked block made: %v", err)
	}
}

// TestBlocksAttackerAgainstAnIllegalPairAsksNothing — Provoke against a
// creature that can't legally block the provoker (a flyer it can't
// reach) creates no requirement it has to obey: nothing is owed, the
// pass is accepted, and the creature may block another attacker.
func TestBlocksAttackerAgainstAnIllegalPairAsksNothing(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flyer := pushCombatant(t, g, me, "Swooping Talon", 2, 6, "flying")
	other := pushCombatant(t, g, me, "Other Bear", 2, 2)
	bear := pushCombatant(t, g, opp, "Grizzly Bears", 2, 2)
	provoke(t, g, bear, objectRefOf(t, g, flyer))
	declareAttacks(t, g, flyer, other)

	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		if w := g.blockRequirementWitnessLocked(opp.ID); w != nil {
			t.Errorf("witness = %+v: a creature that can't block the flyer owes nothing", w)
		}
		if owed := g.MustBlockForEffect(); owed != nil {
			t.Errorf("must_block = %v, want none", owed)
		}
	})
	if err := block(t, g, bear, other); err != nil {
		t.Fatalf("blocking the other attacker when the provoker can't be blocked: %v", err)
	}
	passToDefender(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass: %v", err)
	}
}

// TestBlocksAttackerNamesTheObject — the requirement names an attacking
// OBJECT: a record naming another epoch of the same card (the creature
// left and came back, CR 400.7) asks nothing of the declaration.
func TestBlocksAttackerNamesTheObject(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	provoker := pushCombatant(t, g, me, "Goblin Grappler", 1, 1)
	bear := pushCombatant(t, g, opp, "Grizzly Bears", 2, 2)
	stale := objectRefOf(t, g, provoker)
	stale.Epoch++
	provoke(t, g, bear, stale)
	declareAttacks(t, g, provoker)
	passToDefender(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("a requirement naming a different object of the card was enforced: %v", err)
	}
}

// TestBlocksAttackerFlowPicksTheProvoker — the search weighs each pair
// by what that pair obeys: under a one-block limit, the only slot goes
// to the provoked creature on the provoker (its "blocks each combat"
// plus its "blocks that attacker"), and spending the slot elsewhere is
// refused.
func TestBlocksAttackerFlowPicksTheProvoker(t *testing.T) {
	stubCombatLimits(t, nil, []BlockRule{arbiterBlockLimit(1)})
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushLimitSource(t, g, me, "Silent Arbiter")
	plain := pushCombatant(t, g, me, "Plain", 2, 2)
	provoker := pushCombatant(t, g, me, "Provoker", 2, 2)
	watch := pushCombatant(t, g, opp, "Watchdog", 1, 2)
	withBlockRequirement(t, g, watch, BlockRequirementBlocks)
	provoke(t, g, watch, objectRefOf(t, g, provoker))
	declareAttacks(t, g, plain, provoker)
	var owed []BlockDeclaration
	g.WithWriteLock(func() { owed = g.blockRequirementWitnessLocked(opp.ID) })
	if len(owed) != 1 || owed[0].Blocker != watch || owed[0].Attacker != provoker {
		t.Fatalf("witness = %+v, want the Watchdog on the provoker (two requirements)", owed)
	}
	br := blockRequirementErr(t, block(t, g, watch, plain))
	if br.Requirement.Kind != BlockRequirementBlocksAttacker {
		t.Errorf("blocking the plain attacker gives up %q, want blocksAttacker", br.Requirement.Kind)
	}
	if err := block(t, g, watch, provoker); err != nil {
		t.Fatalf("the Watchdog on the provoker: %v", err)
	}
}

// TestBlocksAttackerSurvivesUndoAndSnapshot — the requirement and the
// object it names are data: an undo and a JSON round trip both keep
// enforcing it.
func TestBlocksAttackerSurvivesUndoAndSnapshot(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	provoker := pushCombatant(t, g, me, "Goblin Grappler", 1, 1)
	other := pushCombatant(t, g, me, "Other Bear", 2, 2)
	bear := pushCombatant(t, g, opp, "Grizzly Bears", 2, 2)
	provoke(t, g, bear, objectRefOf(t, g, provoker))
	declareAttacks(t, g, provoker, other)
	passToDefender(t, g)
	saved := g.Clone()

	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"objects":[{"id":"`+provoker.String()) {
		t.Errorf("the snapshot does not carry the provoker as the record's object")
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	restored, err := snap.Restore()
	if err != nil {
		t.Fatal(err)
	}
	blockRequirementErr(t, restored.PassPriority())
	blockRequirementErr(t, block(t, restored, bear, other))

	if err := block(t, g, bear, provoker); err != nil {
		t.Fatal(err)
	}
	g.RestoreFrom(saved)
	blockRequirementErr(t, g.PassPriority())
	if err := block(t, g, bear, provoker); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass after the undo and the block: %v", err)
	}
}

// TestBlocksAttackerModIsValidated — the record's object is required
// on blocksAttacker and refused everywhere else, at registration.
func TestBlocksAttackerModIsValidated(t *testing.T) {
	ref := ObjectRef{ID: uuid.New()}
	cases := []struct {
		name string
		mod  Mod
		bad  bool
	}{
		{"blocksAttacker with its object", BlocksAttackerMod(ref), false},
		{"blocksAttacker with no object", AddBlockRequirementMod(BlockRequirementBlocksAttacker), true},
		{"lure naming an object", Mod{Kind: ModAddBlockRequirement, Text: string(BlockRequirementLure), Objects: []ObjectRef{ref}}, true},
		{"another mod naming an object", Mod{Kind: ModModifyPT, Power: 1, Objects: []ObjectRef{ref}}, true},
		{"unknown kind", AddBlockRequirementMod("blocksEverything"), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := blockRequirementModProblem(tc.mod) != ""; got != tc.bad {
				t.Errorf("problem = %q, want refused=%v", blockRequirementModProblem(tc.mod), tc.bad)
			}
		})
	}
}

// filteredLureOracle is the stub oracle the filtered-Lure tests hang a
// static on.
const filteredLureOracle = "test-filtered-lure-1684"

// withFilteredLure makes the card with filteredLureOracle carry "All
// <filter> able to block this creature do so".
func withFilteredLure(t *testing.T, filter string) {
	t.Helper()
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID != filteredLureOracle {
			return nil
		}
		return []StaticAbility{{
			Layer: Layer6Ability,
			AppliesTo: func(target *Card, _ *Game, source *Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *Characteristic, _ *Card, _ *Game, source *Card) {
				c.BlockRequirements = append(c.BlockRequirements, BlockRequirement{
					Kind: BlockRequirementLure, Filter: filter, Source: source.InstanceID, SourceName: source.Name,
				})
			},
		}}
	})
}

// pushTyped seeds a combatant (not summoning sick, like pushCombatant)
// with a type line and oracle ID of its own.
func pushTyped(t *testing.T, g *Game, owner *Player, name, typeLine, oracleID string, power, toughness int, keywords ...string) uuid.UUID {
	t.Helper()
	id := pushCombatant(t, g, owner, name, power, toughness, keywords...)
	g.WithWriteLock(func() {
		c := findBattlefieldCard(g, id)
		c.TypeLine = typeLine
		c.OracleID = oracleID
		c.effective = printedEffectiveWith(*c)
		g.layerVersion.Add(1)
	})
	return id
}

// TestFilteredLureBindsOnlyMatchingBlockers — "All Walls able to block
// this creature do so" binds the Walls and nobody else; "all creatures
// with flying …" binds the flyers.
func TestFilteredLureBindsOnlyMatchingBlockers(t *testing.T) {
	t.Run("walls", func(t *testing.T) {
		withFilteredLure(t, BlockerFilterWall)
		g := newActiveGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		priest := pushTyped(t, g, me, "Marble Priest", "Artifact Creature — Cleric", filteredLureOracle, 3, 3)
		other := pushCombatant(t, g, me, "Other Bear", 2, 2)
		wall := pushTyped(t, g, opp, "Wall of Stone", "Creature — Wall", "", 0, 8, "defender")
		bear := pushTyped(t, g, opp, "Grizzly Bears", "Creature — Bear", "", 2, 2)
		declareAttacks(t, g, priest, other)
		var owed []BlockDeclaration
		g.WithWriteLock(func() { owed = g.blockRequirementWitnessLocked(opp.ID) })
		if len(owed) != 1 || owed[0].Blocker != wall || owed[0].Attacker != priest {
			t.Fatalf("witness = %+v, want only the Wall on Marble Priest", owed)
		}
		if err := block(t, g, bear, other); err != nil {
			t.Fatalf("a non-Wall is free to block elsewhere: %v", err)
		}
		br := blockRequirementErr(t, block(t, g, wall, other))
		if got := br.Sentence(opp.ID); got != "Wall of Stone must block Marble Priest if able." {
			t.Errorf("sentence = %q", got)
		}
		if err := block(t, g, wall, priest); err != nil {
			t.Fatalf("the Wall on the Priest: %v", err)
		}
		passToDefender(t, g)
		if err := g.PassPriority(); err != nil {
			t.Fatalf("pass with the Wall blocking: %v", err)
		}
	})
	t.Run("flyers", func(t *testing.T) {
		withFilteredLure(t, BlockerFilterFlying)
		g := newActiveGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		piper := pushTyped(t, g, me, "Talruum Piper", "Creature — Minotaur", filteredLureOracle, 3, 3)
		bird := pushTyped(t, g, opp, "Birds", "Creature — Bird", "", 1, 1, "flying")
		pushTyped(t, g, opp, "Grizzly Bears", "Creature — Bear", "", 2, 2)
		declareAttacks(t, g, piper)
		var owed []BlockDeclaration
		g.WithWriteLock(func() { owed = g.blockRequirementWitnessLocked(opp.ID) })
		if len(owed) != 1 || owed[0].Blocker != bird {
			t.Fatalf("witness = %+v, want only the flyer", owed)
		}
		passToDefender(t, g)
		blockRequirementErr(t, g.PassPriority())
		if err := block(t, g, bird, piper); err != nil {
			t.Fatal(err)
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("pass with the flyer blocking and the Bears home: %v", err)
		}
	})
	t.Run("an unregistered key binds nobody", func(t *testing.T) {
		withFilteredLure(t, "no-such-filter")
		g := newActiveGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		priest := pushTyped(t, g, me, "Marble Priest", "Artifact Creature — Cleric", filteredLureOracle, 3, 3)
		pushTyped(t, g, opp, "Wall of Stone", "Creature — Wall", "", 0, 8, "defender")
		declareAttacks(t, g, priest)
		passToDefender(t, g)
		if err := g.PassPriority(); err != nil {
			t.Fatalf("an unregistered filter bound a blocker: %v", err)
		}
	})
}

// TestBlockerFilterRegistry — the two built-in keys are registered, the
// empty key is "no filter", and a duplicate registration panics.
func TestBlockerFilterRegistry(t *testing.T) {
	for _, k := range []string{"", BlockerFilterWall, BlockerFilterFlying} {
		if !KnownBlockerFilter(k) {
			t.Errorf("%q is not known", k)
		}
	}
	if KnownBlockerFilter("no-such-filter") {
		t.Error("an unregistered key reads as known")
	}
	defer func() {
		if recover() == nil {
			t.Error("a duplicate registration did not panic")
		}
	}()
	RegisterBlockerFilter(BlockerFilterWall, func(*Card) bool { return true })
}
