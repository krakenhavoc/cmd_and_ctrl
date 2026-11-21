package testsso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// stealWith applies GainControlOfSpell as if a resolving spell
// controlled by `thief` did it.
func stealWith(t *testing.T, g *game.Game, spell, thief uuid.UUID, chooseNewTargets bool) {
	t.Helper()
	g.WithWriteLock(func() {
		ctx := effects.NewContext(g, &game.StackItem{
			ID: uuid.New(), Kind: game.StackItemTriggered,
			Controller: thief, Owner: thief, SourceCardID: uuid.New(),
		})
		if err := (effects.GainControlOfSpell{Spell: spell, ChooseNewTargets: chooseNewTargets}).Apply(ctx); err != nil {
			t.Fatalf("GainControlOfSpell: %v", err)
		}
	})
}

// TestStolenDivinationDrawsForTheThief — CR 608.2c / 109.5: the
// controller follows the instructions, and "you" is the controller.
func TestStolenDivinationDrawsForTheThief(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, thief := g.Seats[0], g.Seats[1]
	id := effectstest.CastCatalogSpell(t, g, "Divination", "Sorcery", effectstest.SpellControlDivinationOracle, nil)
	casterHand, thiefHand := len(caster.Hand.Cards), len(thief.Hand.Cards)

	stealWith(t, g, id, thief.ID, false)
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	caster, thief := g.Seats[0], g.Seats[1]
	bolt := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: thief.ID}})

	stealWith(t, g, bolt, thief.ID, true)
	prompt := effectstest.LatestRetarget(g, thief.ID)
	if prompt == nil {
		t.Fatal("no retarget prompt for the new controller")
	}
	if !effectstest.HasID(prompt.PickTargetPlayers, caster.ID) {
		t.Fatalf("the thief cannot aim the Bolt at its caster: %v", prompt.PickTargetPlayers)
	}
	if err := g.ResolveRetarget(prompt.ID, thief.ID,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: caster.ID}}); err != nil {
		t.Fatalf("ResolveRetarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, caster.ID); got != 37 {
		t.Errorf("caster's life = %d, want 37", got)
	}
	if got := effectstest.LifeOf(g, thief.ID); got != 40 {
		t.Errorf("thief's life = %d, want 40", got)
	}
}

// TestStolenOneRingIsNotCastByItsController — "if you cast it": the
// thief controls the Ring and did not cast it (ADR 0104 §6). The same
// Ring cast by its controller still passes.
func TestStolenOneRingIsNotCastByItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	thief := g.Seats[1]
	stolen := effectstest.CastCatalogSpell(t, g, "The One Ring", "Legendary Artifact", effectstest.TheOneRingOracle, nil)
	stealWith(t, g, stolen, thief.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(stolen) {
		t.Fatal("the stolen Ring did not resolve")
	}
	var got bool
	g.ReadSnapshot(func() { got = effects.B16EnteredFromStack(g, stolen) })
	if got {
		t.Error(`"if you cast it" is true for a thief who did not cast the Ring`)
	}

	g2 := effectstest.NewCatalogGame(t)
	own := effectstest.CastCatalogSpell(t, g2, "The One Ring", "Legendary Artifact", effectstest.TheOneRingOracle, nil)
	effectstest.PassPriorityAroundTable(t, g2)
	g2.ReadSnapshot(func() { got = effects.B16EnteredFromStack(g2, own) })
	if !got {
		t.Error(`"if you cast it" is false for the player who cast the Ring`)
	}
}

// TestExchangeControlOfSpellAndOffersTheNewControllerTargets — the
// Sudden Substitution / Perplexing Chimera shape: control swaps both
// ways and the spell's NEW controller is offered new targets.
func TestExchangeControlOfSpellAndOffersTheNewControllerTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, other := g.Seats[0], g.Seats[1]
	bolt := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: other.ID}})
	creature := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{InstanceID: creature, Name: "Chimera", TypeLine: "Creature — Chimera",
			Power: 3, Toughness: 3, Owner: other.ID, Controller: other.ID})
	})
	var did bool
	g.WithWriteLock(func() {
		ctx := effects.NewContext(g, &game.StackItem{ID: uuid.New(), Kind: game.StackItemTriggered,
			Controller: other.ID, Owner: other.ID, SourceCardID: creature})
		var err error
		did, err = effects.ExchangeControlOfSpellAnd{Spell: bolt, Permanent: creature, ChooseNewTargets: true}.ApplyAndReport(ctx)
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
	if effectstest.LatestRetarget(g, other.ID) == nil {
		t.Error("the spell's new controller was not offered new targets")
	}
}
