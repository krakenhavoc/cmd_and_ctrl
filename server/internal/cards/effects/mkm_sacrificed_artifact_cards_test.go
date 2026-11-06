package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mkm_sacrificed_artifact_cards_test.go — "if you've sacrificed an
// artifact this turn" (#5224) and the Murders at Karlov Manor cards that
// read it. The tally cell itself is pinned in
// game/turn_tally_sacrificed_card_type_test.go; these play each card's
// real catalog entry.

const (
	detectivesSatchelOracle = "e4ecfccd-b8b4-4b18-b989-30de79b4971e"
	furtiveCourierOracle    = "92d24b75-419b-4068-89c1-db9945f74972"
)

func TestSacrificedArtifactCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		detectivesSatchelOracle: "Detective's Satchel",
		furtiveCourierOracle:    "Furtive Courier",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// mkmSacrifice sacrifices a permanent `who` controls.
func mkmSacrifice(t *testing.T, g *game.Game, who, id uuid.UUID) {
	t.Helper()
	if err := g.SacrificePermanent(who, id); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
}

// --- the condition helper ---------------------------------------------------

// The three readers agree: nothing before a sacrifice, a non-artifact
// does not count, an opponent's artifact is theirs, and a token artifact
// of mine does.
func TestSacrificedAnArtifactThisTurnReaders(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	holds := func() (plain, activation, cond bool) {
		g.WithWriteLock(func() {
			plain = SacrificedAnArtifactThisTurn(g, me.ID)
			activation = YouSacrificedAnArtifactThisTurn(g, me.ID, uuid.Nil)
			cond = CondSacrificedAnArtifact{}.Holds(&Context{Game: g, Item: &game.StackItem{Controller: me.ID}})
		})
		return
	}
	if a, b, c := holds(); a || b || c {
		t.Fatal("true before anything was sacrificed")
	}
	mkmSacrifice(t, g, me.ID, b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2))
	mkmSacrifice(t, g, opp.ID, b12Creature(g, opp.ID, "Their Rock", "Artifact", 0, 0))
	if a, b, c := holds(); a || b || c {
		t.Error("true after a creature of mine and an opponent's artifact")
	}
	var clue uuid.UUID
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me.ID, ClueToken(), 1); err != nil {
			t.Fatal(err)
		}
	})
	for _, c := range tokensNamed(g, "Clue") {
		clue = c.InstanceID
	}
	mkmSacrifice(t, g, me.ID, clue)
	if a, b, c := holds(); !a || !b || !c {
		t.Errorf("after a Clue token: plain %v, activation %v, condition %v; want all true", a, b, c)
	}
	step, ok := CondSacrificedAnArtifact{}.condition()
	if !ok || step.Cond != game.FootprintCondArtifactSacrificed {
		t.Errorf("footprint %+v (%v), want %s", step, ok, game.FootprintCondArtifactSacrificed)
	}
}

// --- Detective's Satchel -----------------------------------------------------

// It investigates twice as it enters, and its Thopter ability is refused
// until an artifact has been sacrificed this turn: cracking one of its
// own Clues (a token, gone at once) is enough.
func TestDetectivesSatchel(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	satchel := celebrationEnterFromHand(t, g, me.ID, "Detective's Satchel", "Artifact", detectivesSatchelOracle)
	passPriorityAroundTable(t, g)
	clues := tokensNamed(g, "Clue")
	if len(clues) != 2 {
		t.Fatalf("%d Clues after it entered, want 2", len(clues))
	}

	if err := g.ActivateCatalogAbility(me.ID, satchel, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated with no artifact sacrificed this turn")
	}
	mkmSacrifice(t, g, me.ID, clues[0].InstanceID)
	if err := g.ActivateCatalogAbility(me.ID, satchel, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate after sacrificing a Clue: %v", err)
	}
	passPriorityAroundTable(t, g)
	thopters := tokensNamed(g, "Thopter")
	if len(thopters) != 1 {
		t.Fatalf("%d Thopters, want 1", len(thopters))
	}
	th := thopters[0]
	if th.Controller != me.ID || !th.IsArtifact() || !th.IsCreature() || th.Power != 1 || th.Toughness != 1 ||
		!game.HasKeyword(&th, "flying") || len(th.Colors) != 0 {
		t.Errorf("Thopter %+v, want a 1/1 colorless artifact creature with flying", th)
	}
	if c, ok := battlefieldCard(g, satchel); !ok || !c.Tapped {
		t.Error("the Satchel is not tapped after its {T} ability")
	}
}

// --- Furtive Courier ---------------------------------------------------------

// With nothing sacrificed it can be blocked; its attack loots either way.
func TestFurtiveCourierCanBeBlockedAndLoots(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "Drawn")
	courier := b12Push(g, me.ID, "Furtive Courier", "Creature — Merfolk Advisor", furtiveCourierOracle, 3, 2)
	wall := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 4)
	advanceToMain(t, g)
	// An opponent's artifact sacrifice is theirs.
	mkmSacrifice(t, g, opp.ID, b12Creature(g, opp.ID, "Their Rock", "Artifact", 0, 0))
	hand := len(me.Hand.Cards)
	pr7bAttack(t, g, opp.ID, courier)
	if len(me.Hand.Cards) != hand+1 {
		t.Fatalf("hand %d → %d after the attack trigger, want one drawn", hand, len(me.Hand.Cards))
	}
	answerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
	passPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != hand {
		t.Errorf("hand %d → %d, want draw then discard", hand, len(me.Hand.Cards))
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, courier); err != nil {
		t.Errorf("with no artifact of mine sacrificed the Courier can be blocked: %v", err)
	}
}

// After an artifact creature of its controller's is sacrificed, it can't
// be blocked.
func TestFurtiveCourierIsUnblockableAfterAnArtifactSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "Drawn")
	courier := b12Push(g, me.ID, "Furtive Courier", "Creature — Merfolk Advisor", furtiveCourierOracle, 3, 2)
	wall := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 4)
	advanceToMain(t, g)
	mkmSacrifice(t, g, me.ID, b12Creature(g, me.ID, "Ornithopter", "Artifact Creature — Thopter", 0, 2))
	pr7bAttack(t, g, opp.ID, courier)
	discardFromHand(t, g, me.ID)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, courier); err == nil {
		t.Fatal("after an artifact sacrifice this turn the Courier can't be blocked")
	}
}
