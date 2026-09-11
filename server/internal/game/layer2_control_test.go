package game

import (
	"testing"

	"github.com/google/uuid"
)

// layer2_control_test.go pins the engine half of the S24 control
// change: the baseline capture, the materialisation back onto
// Card.Controller, and the two CR consequences that ride with it.
// The catalog half (Mind Control) is pinned in
// cards/effects/mind_control_test.go.
//
// These drive a stubbed CatalogStaticAbilities hook rather than a
// catalog card, so they pin the engine contract on its own.

// controlGrabStatic returns a layer-2 static that hands control of
// every battlefield card matching `pred` to the source's controller
// — the shape of every "you control enchanted creature" effect,
// without the attachment.
func controlGrabStatic(pred func(target *Card) bool) StaticAbility {
	return StaticAbility{
		Layer: Layer2Control,
		AppliesTo: func(target *Card, _ *Game, _ *Card) bool {
			return pred(target)
		},
		Apply: func(c *Characteristic, _ *Card, _ *Game, source *Card) {
			c.Controller = source.Controller
		},
	}
}

// Every battlefield card gets a seeded Characteristic.Controller,
// whether or not any layer-2 effect exists. Without that, the
// materialisation step would have to guess what uuid.Nil meant, and
// the first recompute of a game with no control effects in it would
// zero every controller on the board.
func TestEveryEffectiveCharacteristicCarriesItsController(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushAttachTestCard(g, me.ID, "Bear", "Creature — Bear")
	theirs := pushAttachTestCard(g, opp.ID, "Ox", "Creature — Ox")

	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			eff := c.Effective()
			if c.InstanceID == mine && eff.Controller != me.ID {
				t.Errorf("mine: effective controller %s, want %s", eff.Controller, me.ID)
			}
			if c.InstanceID == theirs && eff.Controller != opp.ID {
				t.Errorf("theirs: effective controller %s, want %s", eff.Controller, opp.ID)
			}
			if c.Controller != eff.Controller {
				t.Errorf("%s: Card.Controller %s disagrees with effective %s",
					c.Name, c.Controller, eff.Controller)
			}
		}
	})
}

// The baseline is what makes control revert with no bookkeeping.
func TestControlRevertsWhenTheEffectEnds(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushAttachTestCard(g, opp.ID, "Bear", "Creature — Bear")
	thiefID := pushAttachTestCard(g, me.ID, "Thief", "Enchantment")

	active := true
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if !active {
			return nil
		}
		return []StaticAbility{controlGrabStatic(func(target *Card) bool {
			return target.InstanceID == victim
		})}
	})
	// Only the thief carries the static: give it an oracle ID the
	// stub can key on, and leave the victim without one.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == thiefID {
				g.Battlefield.Cards[i].OracleID = "thief"
			}
		}
	})
	g.BumpLayerVersionForTest()

	if got := controllerOfCard(t, g, victim); got != me.ID {
		t.Fatalf("stolen: controller %s, want %s", got, me.ID)
	}
	// CR 302.6 — the steal costs the new controller a turn of
	// summoning sickness.
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == victim && !g.Battlefield.Cards[i].SummonedThisTurn {
				t.Error("a creature that changed control should be summoning-sick")
			}
		}
	})

	active = false
	g.BumpLayerVersionForTest()
	if got := controllerOfCard(t, g, victim); got != opp.ID {
		t.Errorf("reverted: controller %s, want the baseline %s", got, opp.ID)
	}
}

// The baseline is battlefield-scoped: leaving clears it so a
// re-entry under a different player captures the new truth rather
// than dragging the old one back.
func TestControlBaselineIsRecapturedOnReEntry(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	card := pushAttachTestCard(g, opp.ID, "Bear", "Creature — Bear")

	// Force the first capture.
	g.ReadSnapshot(func() {})
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == card && g.Battlefield.Cards[i].BaseController != opp.ID {
				t.Errorf("baseline %s, want %s", g.Battlefield.Cards[i].BaseController, opp.ID)
			}
		}
	})

	// Flicker it back in under the other seat.
	g.WithWriteLock(func() {
		moved, err := MoveCard(g.Battlefield, opp.Hand, card)
		if err != nil {
			t.Fatalf("MoveCard out: %v", err)
		}
		if moved.BaseController != uuid.Nil {
			t.Errorf("baseline survived battlefield exit: %s", moved.BaseController)
		}
		if _, err := MoveCard(opp.Hand, g.Battlefield, card); err != nil {
			t.Fatalf("MoveCard in: %v", err)
		}
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == card {
				g.Battlefield.Cards[i].Controller = me.ID
			}
		}
		g.EmitEvent(Event{Kind: EventZoneMove, CardID: card, OldZone: ZoneHand, NewZone: ZoneBattlefield})
	})
	g.ReadSnapshot(func() {})
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == card && g.Battlefield.Cards[i].BaseController != me.ID {
				t.Errorf("re-entry baseline %s, want %s",
					g.Battlefield.Cards[i].BaseController, me.ID)
			}
		}
	})
}

func controllerOfCard(t *testing.T, g *Game, id uuid.UUID) uuid.UUID {
	t.Helper()
	var out uuid.UUID
	var found bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				out, found = c.Controller, true
				return
			}
		}
	})
	if !found {
		t.Fatalf("card %s not on battlefield", id)
	}
	return out
}
