package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spell_control_test.go — the card half of ADR 0104 (#1745): what a
// stolen spell DOES. The engine half (the stack step, the hand-off,
// CR 800.4a) is in game/spell_control_test.go.

const spellControlDivinationOracle = "273b339c-964b-4a18-8eb5-ceb8abcdfd9e"

// stealWith applies GainControlOfSpell as if a resolving spell
// controlled by `thief` did it.
func stealWith(t *testing.T, g *game.Game, spell, thief uuid.UUID, chooseNewTargets bool) {
	t.Helper()
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{
			ID: uuid.New(), Kind: game.StackItemTriggered,
			Controller: thief, Owner: thief, SourceCardID: uuid.New(),
		})
		if err := (GainControlOfSpell{Spell: spell, ChooseNewTargets: chooseNewTargets}).Apply(ctx); err != nil {
			t.Fatalf("GainControlOfSpell: %v", err)
		}
	})
}

// TestStolenDivinationDrawsForTheThief — CR 608.2c / 109.5: the
// controller follows the instructions, and "you" is the controller.
func TestStolenDivinationDrawsForTheThief(t *testing.T) {
	g := newCatalogGame(t)
	caster, thief := g.Seats[0], g.Seats[1]
	id := castCatalogSpell(t, g, "Divination", "Sorcery", spellControlDivinationOracle, nil)
	casterHand, thiefHand := len(caster.Hand.Cards), len(thief.Hand.Cards)

	stealWith(t, g, id, thief.ID, false)
	passPriorityAroundTable(t, g)

	if got := len(thief.Hand.Cards); got != thiefHand+2 {
		t.Errorf("thief's hand = %d, want %d — a stolen Divination draws for its new controller", got, thiefHand+2)
	}
	if got := len(caster.Hand.Cards); got != casterHand {
		t.Errorf("caster's hand = %d, want %d — the caster no longer controls the spell", got, casterHand)
	}
	if !caster.Graveyard.Contains(id) {
		t.Error("the resolved sorcery goes to its OWNER's graveyard (CR 608.2n)")
	}
}

// TestStolenBoltOffersItsNewControllerNewTargets — "you may choose new
// targets for it", asked of the thief after the steal, with the legal
// set judged for the thief.
func TestStolenBoltOffersItsNewControllerNewTargets(t *testing.T) {
	g := newCatalogGame(t)
	caster, thief := g.Seats[0], g.Seats[1]
	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: thief.ID}})

	stealWith(t, g, bolt, thief.ID, true)
	prompt := latestRetarget(g, thief.ID)
	if prompt == nil {
		t.Fatal("no retarget prompt for the new controller")
	}
	if !hasID(prompt.PickTargetPlayers, caster.ID) {
		t.Fatalf("the thief cannot aim the Bolt at its caster: %v", prompt.PickTargetPlayers)
	}
	if err := g.ResolveRetarget(prompt.ID, thief.ID,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: caster.ID}}); err != nil {
		t.Fatalf("ResolveRetarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, caster.ID); got != 37 {
		t.Errorf("caster's life = %d, want 37", got)
	}
	if got := lifeOf(g, thief.ID); got != 40 {
		t.Errorf("thief's life = %d, want 40", got)
	}
}

// TestStolenOneRingIsNotCastByItsController — "if you cast it": the
// thief controls the Ring and did not cast it (ADR 0104 §6). The same
// Ring cast by its controller still passes.
func TestStolenOneRingIsNotCastByItsController(t *testing.T) {
	g := newCatalogGame(t)
	thief := g.Seats[1]
	stolen := castCatalogSpell(t, g, "The One Ring", "Legendary Artifact", theOneRingOracle, nil)
	stealWith(t, g, stolen, thief.ID, false)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(stolen) {
		t.Fatal("the stolen Ring did not resolve")
	}
	var got bool
	g.ReadSnapshot(func() { got = b16EnteredFromStack(g, stolen) })
	if got {
		t.Error(`"if you cast it" is true for a thief who did not cast the Ring`)
	}

	g2 := newCatalogGame(t)
	own := castCatalogSpell(t, g2, "The One Ring", "Legendary Artifact", theOneRingOracle, nil)
	passPriorityAroundTable(t, g2)
	g2.ReadSnapshot(func() { got = b16EnteredFromStack(g2, own) })
	if !got {
		t.Error(`"if you cast it" is false for the player who cast the Ring`)
	}
}

// TestExchangeControlOfSpellAndOffersTheNewControllerTargets — the
// Sudden Substitution / Perplexing Chimera shape: control swaps both
// ways and the spell's NEW controller is offered new targets.
func TestExchangeControlOfSpellAndOffersTheNewControllerTargets(t *testing.T) {
	g := newCatalogGame(t)
	caster, other := g.Seats[0], g.Seats[1]
	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: other.ID}})
	creature := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{InstanceID: creature, Name: "Chimera", TypeLine: "Creature — Chimera",
			Power: 3, Toughness: 3, Owner: other.ID, Controller: other.ID})
	})
	var did bool
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{ID: uuid.New(), Kind: game.StackItemTriggered,
			Controller: other.ID, Owner: other.ID, SourceCardID: creature})
		var err error
		did, err = ExchangeControlOfSpellAnd{Spell: bolt, Permanent: creature, ChooseNewTargets: true}.ApplyAndReport(ctx)
		if err != nil {
			t.Fatalf("exchange: %v", err)
		}
	})
	if !did {
		t.Fatal("the exchange did not happen")
	}
	if got := g.StackMeta[bolt].Controller; got != other.ID {
		t.Errorf("Bolt controller = %s, want %s", got, other.ID)
	}
	var creatureCtl uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == creature {
				creatureCtl = c.Controller
			}
		}
	})
	if creatureCtl != caster.ID {
		t.Errorf("creature controller = %s, want %s", creatureCtl, caster.ID)
	}
	if latestRetarget(g, other.ID) == nil {
		t.Error("the spell's new controller was not offered new targets")
	}
}
