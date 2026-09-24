package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ninjutsu_view_test.go — the wire half of #1227. Nothing on the wire
// is NEW: #1213 already ships `return_label` / `return_options` on
// ActivatedAbilityView and takes `return_ids` back on the payload, and
// #1221 already ships a hand card's rows as `zone_abilities`. What is
// new is that the two meet, and that the option list is now a COMBAT
// fact — so the assertion is the #544 one: the picker offers exactly
// the permanents the engine would accept, and no others.

// seatNinjutsuHandCard puts a ninjutsu card in a seat's hand: a hand
// ability whose cost is mana plus "return an unblocked attacker you
// control to hand".
func seatNinjutsuHandCard(p *game.Player) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id,
		Name:       "Ninja of the Deep Hours",
		TypeLine:   "Creature — Human Ninja",
		OracleID:   "00000000-0000-0000-0000-0000000000nj",
		Power:      2, Toughness: 2,
		Owner: p.ID, Controller: p.ID,
		KnownBy: map[uuid.UUID]bool{p.ID: true},
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label: "Ninjutsu {1}{U}",
			Cost: game.AbilityCost{
				Mana: "{1}{U}",
				ReturnToHand: &game.ReturnToHandCost{
					Count: 1,
					Filter: &game.TargetSpec{
						Mode:  "permanent",
						Label: "an unblocked attacker you control",
						Zones: []game.ZoneKind{game.ZoneBattlefield},
						CardOK: func(g *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
							return c.IsCreature() && g.UnblockedAttackerForEffect(c.InstanceID)
						},
						Min: 1, Max: 1,
					},
					Label: "an unblocked attacker you control",
				},
			},
			Zones: []game.ZoneKind{game.ZoneHand},
		}},
	})
	return id
}

// seatCreatureFor puts a plain creature on the battlefield.
func seatCreatureFor(g *game.Game, controller uuid.UUID, name string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Rat",
		Power: 1, Toughness: 1, Owner: controller, Controller: controller,
	})
	return id
}

// The picker's list is exactly the unblocked attackers the seat
// controls: not the creature sitting at home, not the blocked one, and
// not the opponent's attacker.
func TestNinjutsuReturnOptionsAreTheUnblockedAttackersOnly(t *testing.T) {
	g := buildActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	ninja := seatNinjutsuHandCard(me)
	unblocked := seatCreatureFor(g, me.ID, "Sneaky Rat")
	blocked := seatCreatureFor(g, me.ID, "Blocked Rat")
	seatCreatureFor(g, me.ID, "Homebody")
	theirs := seatCreatureFor(g, them.ID, "Their Rat")
	wall := seatCreatureFor(g, them.ID, "Wall")

	g.WithWriteLock(func() {
		g.Turn.Step = game.StepDeclareBlockers
		for i := range g.Battlefield.Cards {
			switch g.Battlefield.Cards[i].InstanceID {
			case unblocked, blocked:
				g.Battlefield.Cards[i].AttackingTarget = them.ID
			case theirs:
				g.Battlefield.Cards[i].AttackingTarget = me.ID
			case wall:
				g.Battlefield.Cards[i].BlockingTarget = blocked
			}
		}
	})
	g.BumpLayerVersionForTest()
	// #1279: an attacker is unblocked only once its defending player has
	// DECLARED — here, declared no blocks beyond what is staged.
	if err := g.FinishBlocks(them.ID); err != nil {
		t.Fatalf("FinishBlocks: %v", err)
	}

	v := FilterViewFor(ViewOfGame(g), me.ID.String())
	c := handCardView(t, v, 0, ninja)
	if len(c.ZoneAbilities) != 1 {
		t.Fatalf("zone_abilities = %+v, want the one ninjutsu row", c.ZoneAbilities)
	}
	ab := c.ZoneAbilities[0]
	if ab.ReturnLabel != "an unblocked attacker you control" {
		t.Errorf("return_label = %q, want the clause as printed", ab.ReturnLabel)
	}
	if ab.ReturnOptions == nil {
		t.Fatal("return_options is absent on an ability that prints the clause")
	}
	if ab.ReturnOptions.Min != 1 || ab.ReturnOptions.Max != 1 {
		t.Errorf("return_options min/max = %d/%d, want 1/1", ab.ReturnOptions.Min, ab.ReturnOptions.Max)
	}
	got := ab.ReturnOptions.Cards
	if len(got) != 1 || got[0] != unblocked.String() {
		t.Errorf("return_options.cards = %v, want just the unblocked attacker %s (not the blocked one, the homebody or the opponent's)",
			got, unblocked)
	}
}

// Outside the declare-blockers window nothing is unblocked, so the
// picker has nothing to offer — the same answer the engine and the
// enumerator give.
func TestNinjutsuReturnOptionsAreEmptyBeforeBlockersAreDeclared(t *testing.T) {
	g := buildActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	ninja := seatNinjutsuHandCard(me)
	attacker := seatCreatureFor(g, me.ID, "Sneaky Rat")

	g.WithWriteLock(func() {
		g.Turn.Step = game.StepDeclareAttackers
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == attacker {
				g.Battlefield.Cards[i].AttackingTarget = them.ID
			}
		}
	})
	g.BumpLayerVersionForTest()

	v := FilterViewFor(ViewOfGame(g), me.ID.String())
	ab := handCardView(t, v, 0, ninja).ZoneAbilities[0]
	if ab.ReturnOptions != nil && len(ab.ReturnOptions.Cards) != 0 {
		t.Errorf("return_options.cards = %v in the declare-attackers step, want none", ab.ReturnOptions.Cards)
	}
}
