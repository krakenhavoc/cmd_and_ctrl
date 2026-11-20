package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// peace_talks_test.go — #5718: Peace Talks, driven through the real turn
// machine on a four-seat table. Cast on seat 0's turn, it lasts through
// seat 0's turn and seat 1's, and is gone on seat 2's.

const peaceTalksOracle = "b0f87e25-6604-4723-9467-1e4fedc1157a"

func castPeaceTalks(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	id := castCatalogSpell(t, g, "Peace Talks", "Sorcery", peaceTalksOracle, nil)
	passPriorityAroundTable(t, g)
	return id
}

// boltErr is the error from casting Lightning Bolt for the active seat at
// `target`, nil when the cast was allowed.
func boltErr(t *testing.T, g *game.Game, target game.TargetRef) error {
	t.Helper()
	return castCatalogSpellErr(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{target})
}

// Every creature on the table, read live: none can attack on the turn it
// resolves or on the next, a creature that arrives after it resolved
// included, and they attack again on the turn after.
func TestPeaceTalksCreaturesCantAttackThisTurnAndNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, next, third := g.Seats[0], g.Seats[1], g.Seats[2]
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, next.ID, "Their Bear", 2, 2)
	later := pushVanillaCreature(g, third.ID, "Third Bear", 2, 2)
	castPeaceTalks(t, g)
	arrived := pushVanillaCreature(g, next.ID, "Arrived Bear", 2, 2)

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(mine, next.ID); err == nil {
		t.Error("a creature attacked on the turn Peace Talks resolved")
	}

	advanceToDeclareAttackersOf(t, g, 1)
	if err := g.DeclareAttacker(theirs, me.ID); err == nil {
		t.Error("a creature attacked on the next turn")
	}
	if err := g.DeclareAttacker(arrived, me.ID); err == nil {
		t.Error("a creature that arrived after Peace Talks resolved attacked on the next turn")
	}

	advanceToDeclareAttackersOf(t, g, 2)
	if err := g.DeclareAttacker(later, me.ID); err != nil {
		t.Errorf("a creature could not attack on the turn after the next: %v", err)
	}
}

// Spells (an Aura spell among them) and activated abilities can't target
// any player or permanent, the caster's own included, this turn and next
// turn; a triggered ability still targets; and the turn after, a spell
// targets again.
func TestPeaceTalksBarsSpellsAndActivatedAbilitiesButNotTriggeredAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me, next, third := g.Seats[0], g.Seats[1], g.Seats[2]
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, next.ID, "Their Bear", 2, 2)
	far := pushVanillaCreature(g, third.ID, "Far Bear", 2, 2)
	splitter := pushCatalogPermanent(g, me.ID, "Bonesplitter", "Artifact — Equipment", bonesplitterOracle, false)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Erhnam Djinn", TypeLine: "Creature — Djinn",
		Power: 4, Toughness: 5, OracleID: erhnamDjinnOracle, Owner: next.ID, Controller: next.ID,
	})
	castPeaceTalks(t, g)

	// Each attempt is made on a copy of the table, so a cast that went
	// through leaves nothing on the stack for the next one.
	for name, ref := range map[string]game.TargetRef{
		"an opponent":               playerRef(next.ID),
		"its own caster":            playerRef(me.ID),
		"an opponent's creature":    cardRef(theirs),
		"the caster's own creature": cardRef(mine),
	} {
		if err := boltErr(t, g.Clone(), ref); err == nil {
			t.Errorf("a spell targeted %s", name)
		}
	}
	if err := castCatalogSpellErr(t, g.Clone(), "Pacifism", "Enchantment — Aura", pacifismOracle, []game.TargetRef{cardRef(theirs)}); err == nil {
		t.Error("an Aura spell targeted a creature (CR 303.4a)")
	}
	equip := len(game.ActivatedAbilitiesForCard(game.Card{OracleID: bonesplitterOracle})) - 1
	if err := g.Clone().ActivateCatalogAbility(me.ID, splitter, equip, game.ActivateAbilityParams{Targets: []game.TargetRef{cardRef(mine)}}); err == nil {
		t.Error("an activated ability (equip) targeted a creature")
	}

	// Seat 1's turn is the next turn. Erhnam Djinn's upkeep trigger is a
	// triggered ability, so it targets.
	advanceToUpkeepOf(t, g, 1)
	if latestPickTarget(g, next.ID) == nil {
		t.Fatal("Erhnam Djinn's trigger asked for no target under Peace Talks")
	}
	pickCard(t, g, next.ID, far)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, far, "forestwalk") {
		t.Error("the triggered ability's target did not gain forestwalk")
	}
	advanceToMainOf(t, g, 1)
	if err := boltErr(t, g.Clone(), playerRef(me.ID)); err == nil {
		t.Error("a spell targeted a player on the next turn")
	}
	if err := boltErr(t, g.Clone(), cardRef(far)); err == nil {
		t.Error("a spell targeted a permanent on the next turn")
	}

	// Seat 2's turn: the effect is over.
	advanceToMainOf(t, g, 2)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{cardRef(theirs)})
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, theirs) != nil {
		t.Error("Lightning Bolt on the turn after the next did not kill its target")
	}
}

// CR 608.2b: a spell already aimed at a creature when Peace Talks resolves
// finds its target illegal as it resolves. Leyline of Anticipation lets
// Peace Talks be cast over the Bolt.
func TestPeaceTalksMakesATargetIllegalForASpellAlreadyOnTheStack(t *testing.T) {
	g := newCatalogGame(t)
	me, next := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Leyline of Anticipation", "Enchantment", leylineOfAnticipationOracle, false)
	theirs := pushVanillaCreature(g, next.ID, "Their Bear", 2, 2)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{cardRef(theirs)})
	castCatalogSpell(t, g, "Peace Talks", "Sorcery", peaceTalksOracle, nil)
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, theirs) == nil {
		t.Error("the Bolt resolved against a target Peace Talks made illegal")
	}
}

// A restore point holds both records, and the restored game still ends
// them after the next turn.
func TestPeaceTalksSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	me, next := g.Seats[0], g.Seats[1]
	theirs := pushVanillaCreature(g, next.ID, "Their Bear", 2, 2)
	castPeaceTalks(t, g)
	advanceToMainOf(t, g, 1)

	g = restoreRoundTrip(t, g, true)
	if err := boltErr(t, g, playerRef(me.ID)); err == nil {
		t.Error("the restored game let a spell target a player")
	}
	advanceToDeclareAttackersOf(t, g, 1)
	if err := g.DeclareAttacker(theirs, me.ID); err == nil {
		t.Error("the restored game let a creature attack")
	}
	advanceToMainOf(t, g, 2)
	if err := boltErr(t, g, playerRef(me.ID)); err != nil {
		t.Errorf("the restored records outlived the next turn: %v", err)
	}
}

func TestPeaceTalksIsFull(t *testing.T) {
	spec, ok := Lookup(peaceTalksOracle)
	if !ok {
		t.Fatal("Peace Talks is not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("completeness %v, caveats %q; want Full with none", spec.Completeness, spec.Caveats)
	}
}
