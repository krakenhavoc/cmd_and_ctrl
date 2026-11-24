package game

import (
	"errors"
	"testing"
)

// attack_power_hand_test.go pins #6087's attack restriction, Ensnaring
// Bridge's "Creatures with power greater than the number of cards in your
// hand can't attack" (CR 508.1c): power read through the layers, the
// Bridge controller's hand read live at declaration, every creature bound,
// the Bridge controller's own included.

const powerHandBridgeOracle = "test-power-hand-bridge-6087"

// withPowerHandBridge makes the card with powerHandBridgeOracle write the
// clause onto every creature, naming its controller.
func withPowerHandBridge(t *testing.T) {
	t.Helper()
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID != powerHandBridgeOracle {
			return nil
		}
		return []StaticAbility{{
			Layer: Layer6Ability,
			AppliesTo: func(target *Card, _ *Game, _ *Card) bool {
				return target.IsCreature()
			},
			Apply: func(c *Characteristic, _ *Card, _ *Game, source *Card) {
				c.AttackTargetRestrictions = append(c.AttackTargetRestrictions, AttackTargetRestriction{
					Source: source.InstanceID, SourceName: source.Name, PowerAtMostHandOf: source.Controller,
				})
			},
		}}
	})
}

// setHandSize leaves exactly n cards in p's hand.
func setHandSize(g *Game, p *Player, n int) {
	g.WithWriteLock(func() {
		p.Hand.Cards = nil
		for i := 0; i < n; i++ {
			p.Hand.PushTop(NewCard("Filler", p.ID))
		}
	})
}

func TestPowerAboveHandCantAttack(t *testing.T) {
	withPowerHandBridge(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// The Bridge is the OPPONENT's: its controller's hand is what counts.
	pushTyped(t, g, opp, "Bridge", "Artifact", powerHandBridgeOracle, 0, 0)
	big := pushCombatant(t, g, me, "Big Bear", 3, 3)
	even := pushCombatant(t, g, me, "Even Bear", 2, 2)
	small := pushCombatant(t, g, me, "Small Bear", 1, 1)
	setHandSize(g, me, 7) // the attacker's own hand does not matter
	setHandSize(g, opp, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)

	err := g.DeclareAttacker(big, opp.ID)
	if !errors.Is(err, ErrIllegalAttackTarget) {
		t.Fatalf("power 3 over a hand of 2: err = %v, want ErrIllegalAttackTarget", err)
	}
	var re *AttackTargetRestrictionError
	if !errors.As(err, &re) {
		t.Fatalf("refusal is %T", err)
	}
	var name string
	g.WithWriteLock(func() { name = g.playerNameLocked(opp.ID) })
	want := "Big Bear can't attack: its power (3) is greater than the number of cards in " + name + "'s hand (2) (Bridge)."
	if got := re.Sentence(); got != want {
		t.Errorf("sentence = %q, want %q", got, want)
	}
	if err := g.DeclareAttacker(even, opp.ID); err != nil {
		t.Fatalf("power equal to the hand may attack: %v", err)
	}
	if err := g.DeclareAttacker(small, opp.ID); err != nil {
		t.Fatalf("power below the hand may attack: %v", err)
	}
}

func TestPowerAboveHandFollowsTheHandAndThePower(t *testing.T) {
	withPowerHandBridge(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// The Bridge is MINE: it binds my own creatures, with my hand.
	pushTyped(t, g, me, "Bridge", "Artifact", powerHandBridgeOracle, 0, 0)
	bear := pushCombatant(t, g, me, "Bear", 2, 2)
	setHandSize(g, me, 1)
	advanceIntoStep(t, g, StepDeclareAttackers)

	var targets []AttackTargetRef
	var chips []DefenderRefusal
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c := findBattlefieldCard(g, bear)
		targets = g.AttackTargetsForAttackerForEffect(c)
		chips = g.DefenderRefusalsForEffect(c)
	})
	if len(targets) != 0 {
		t.Errorf("targets = %+v, want none while power 2 > hand 1", targets)
	}
	if len(chips) != 1 || chips[0].Player != opp.ID {
		t.Errorf("chips = %+v, want one row for the opponent", chips)
	}
	if err := g.DeclareAttacker(bear, opp.ID); !errors.Is(err, ErrIllegalAttackTarget) {
		t.Fatalf("power 2 over a hand of 1: err = %v", err)
	}
	// A card drawn is read at the next declaration.
	setHandSize(g, me, 2)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("power 2, hand 2: %v", err)
	}
}

func TestPowerAboveHandReadsLayeredPowerAndLeavesWithTheBridge(t *testing.T) {
	withPowerHandBridge(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bridge := pushTyped(t, g, opp, "Bridge", "Artifact", powerHandBridgeOracle, 0, 0)
	bear := pushCombatant(t, g, me, "Bear", 1, 1)
	setHandSize(g, opp, 1)
	// A +1/+1 counter takes it to 2, over a hand of 1.
	g.WithWriteLock(func() {
		c := findBattlefieldCard(g, bear)
		if c.Counters == nil {
			c.Counters = map[string]int{}
		}
		c.Counters[CounterPlusOne] = 1
		g.layerVersion.Add(1)
	})
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); !errors.Is(err, ErrIllegalAttackTarget) {
		t.Fatalf("layered power 2 over a hand of 1: err = %v", err)
	}
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneBattlefield}, ZoneRef{Kind: ZoneGraveyard, Owner: opp.ID}, bridge); err != nil {
		t.Fatalf("move the Bridge: %v", err)
	}
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("with the Bridge gone: %v", err)
	}
}
