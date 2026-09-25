package game

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// trigger_target_timing_test.go — #809. A targeted triggered ability
// that fires while a spell is RESOLVING must not have its legal-target
// set frozen against a stack that still holds that spell. CR 603.3:
// the ability goes on the stack the next time a player would receive
// priority, and CR 603.3d chooses its targets then — by which point
// CR 608.2n has put the resolving instant or sorcery in its owner's
// graveyard.

const (
	tt809WizardOracle    = "test-809-wizard"
	tt809ReanimateOracle = "test-809-reanimate"
)

// tt809StackSpellSpec is Dualcaster Mage's clause in miniature:
// target instant or sorcery spell (a card in the stack zone).
func tt809StackSpellSpec() *TargetSpec {
	return &TargetSpec{
		Mode:  "stack_spell",
		Label: "target instant or sorcery spell",
		Zones: []ZoneKind{ZoneStack},
		Min:   1,
		Max:   1,
		CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
			return strings.Contains(c.TypeLine, "Instant") || strings.Contains(c.TypeLine, "Sorcery")
		},
	}
}

// tt809Game seeds a game whose graveyard holds a creature with a
// targeted ETB reaching the stack zone, and whose "Reanimate" instant
// returns it to the battlefield on resolution. Returns the game, the
// controller and the creature's instance ID.
func tt809Game(t *testing.T) (*Game, *Player, uuid.UUID) {
	t.Helper()
	g := newActiveGame(t)
	me := g.Seats[0]
	wizard := uuid.New()
	me.Graveyard.PushTop(Card{
		InstanceID: wizard,
		Name:       "Test Dualcaster",
		TypeLine:   "Creature — Test Wizard",
		OracleID:   tt809WizardOracle,
		Power:      2,
		Toughness:  2,
		Owner:      me.ID,
		Controller: me.ID,
	})
	withCatalogTriggers(t, func(id string) []TriggeredAbility {
		if id != tt809WizardOracle {
			return nil
		}
		return []TriggeredAbility{{
			Watches: []EventKind{EventETB},
			AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
				return ev.CardID == source.InstanceID
			},
			Targets: tt809StackSpellSpec(),
			Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				return newTriggeredItemForTest(source, "copy target instant or sorcery spell",
					func(_ *Game, _ *StackItem) error { return nil })
			},
		}}
	})
	withEffectHooks(t,
		func(g *Game, _ *StackItem, oracleID string) error {
			if oracleID != tt809ReanimateOracle {
				return nil
			}
			return g.ReturnFromGraveyardForEffect(wizard, ZoneBattlefield)
		},
		nil,
		func(oracleID string) bool { return oracleID == tt809ReanimateOracle },
	)
	return g, me, wizard
}

// tt809Cast puts a spell in the caster's hand and casts it.
func tt809Cast(t *testing.T, g *Game, p *Player, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   oracle,
		Owner:      p.ID,
		Controller: p.ID,
	})
	if err := g.CastSpell(p.ID, id, CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// tt809ResolveTop passes priority around the two-seat table once, so
// the top of the stack resolves.
func tt809ResolveTop(t *testing.T, g *Game) {
	t.Helper()
	for i := 0; i < len(g.Seats); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority %d: %v", i, err)
		}
	}
}

// The wedge itself (#809). The reanimating spell is the ONLY thing on
// the stack, so once it has resolved and gone to the graveyard
// (CR 608.2n) the ETB trigger has no legal target at all. CR 603.3d
// removes it from the stack; it must not leave a pick_target prompt
// offering the spell that just left, because since #791 an unanswered
// prompt gates the whole table.
func TestTriggerWithNoTargetLeftAfterResolutionIsRemovedNotPrompted(t *testing.T) {
	g, me, wizard := tt809Game(t)
	tt809Cast(t, g, me, "Reanimate", "Instant", tt809ReanimateOracle)
	tt809ResolveTop(t, g)

	if !g.Battlefield.Contains(wizard) {
		t.Fatal("the creature should have been returned to the battlefield")
	}
	if p := findPickTarget(g, me.ID); p != nil {
		t.Errorf("CR 603.3d: the trigger has no legal target and must be removed, got a prompt offering %v", p.PickTargetCards)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("no prompt of any kind should be pending, got %d", len(g.PendingChoices))
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("the table is wedged: PassPriority: %v", err)
	}
}

// The other half: with a spell still underneath, the prompt survives —
// but it offers ONLY that spell, not the one that just finished
// resolving and left the stack.
func TestTriggerTargetsAreRecomputedAfterTheResolvingSpellLeaves(t *testing.T) {
	g, me, _ := tt809Game(t)
	below := tt809Cast(t, g, me, "Bolt", "Instant", "")
	reanimate := tt809Cast(t, g, me, "Reanimate", "Instant", tt809ReanimateOracle)
	tt809ResolveTop(t, g)

	p := findPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the spell below is still a legal target, so the trigger should prompt")
	}
	if !hasTargetID(p.PickTargetCards, below) {
		t.Errorf("the spell still on the stack must be offered: %v", p.PickTargetCards)
	}
	if hasTargetID(p.PickTargetCards, reanimate) {
		t.Errorf("the resolved spell is in the graveyard (CR 608.2n) and must not be offered: %v", p.PickTargetCards)
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, TargetRef{Kind: TargetCard, ID: below}); err != nil {
		t.Fatalf("answering with the surviving spell: %v", err)
	}
	if err := g.PassPriority(); err != nil && errors.Is(err, ErrChoicePending) {
		t.Fatalf("the table is still gated: %v", err)
	}
}

func hasTargetID(ids []uuid.UUID, id uuid.UUID) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}
