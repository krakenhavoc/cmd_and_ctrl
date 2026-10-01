package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sacrifice_alt_cost_view_test.go — #1727: an alternative cost whose
// card component is a SACRIFICE ships `sacrifice_options`, the block
// the additional cost's sacrifice picker already reads, instead of
// `pay_options`. Four properties:
//
//   - an offer the caster cannot pay (two creatures for "sacrifice
//     three") is not stamped, so the card is not castable_here — the
//     client greys the flashback out without knowing the rule;
//   - the options are the CASTER's permanents only (CR 701.21a), in
//     payment order (tokens first), sized by the clause's count;
//   - the view and the bot enumerator agree on castable_here;
//   - a bystander sees the printed offer and not the caster's list
//     (#1172).
func TestSacrificeAlternativeCostStampsSacrificeOptions(t *testing.T) {
	const oracle = "test-1727-view-dread-return"
	g := busyTable(t, 0)
	me := g.Seats[g.Turn.ActiveSeat]
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceTo(t, g, game.StepPrecombatMain)

	withCastableZones(t, map[string][]game.ZoneKind{oracle: {game.ZoneGraveyard}})
	withAltCostsByOracle(t, map[string][]game.AlternativeCost{
		oracle: {{
			Key:                 "flashback",
			Label:               "Flashback—Sacrifice three creatures",
			FromZone:            game.ZoneGraveyard,
			ExileOnLeavingStack: true,
			PayLabel:            "three creatures",
			Sacrifice: &game.TargetSpec{
				Mode:  "permanent",
				Label: "three creatures",
				Zones: []game.ZoneKind{game.ZoneBattlefield},
				CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
					return c.IsCreature()
				},
				Min: 3, Max: 3,
			},
		}},
	})

	// busyTable seats a Llanowar Elves per player; this board counts
	// creatures exactly, so it starts with none.
	kept := g.Battlefield.Cards[:0]
	for _, c := range g.Battlefield.Cards {
		if !c.IsCreature() {
			kept = append(kept, c)
		}
	}
	g.Battlefield.Cards = kept
	me.Graveyard.Cards = nil
	id := graveyardCard(me, "Test Dread Return", oracle)
	knownToEveryone(g, me.Graveyard, id)

	push := func(p *game.Player, name, typeLine string) uuid.UUID {
		c := game.NewCard(name, p.ID)
		c.TypeLine = typeLine
		c.Power, c.Toughness = 2, 2
		g.Battlefield.PushTop(c)
		return c.InstanceID
	}
	bear := push(me, "Bear", "Creature — Bear")
	wolf := push(me, "Wolf", "Creature — Wolf")
	theirs := push(them, "Their Bear", "Creature — Bear")

	c := castSurfaceOf(t, g, me.ID, me.Seat, id)
	if c.CastableHere || len(c.AlternativeCosts) != 0 {
		t.Fatalf("two creatures: castable_here=%v offers=%+v, want an uncastable card with no offer",
			c.CastableHere, c.AlternativeCosts)
	}
	if offeredFromGraveyard(g, me.ID, id) {
		t.Errorf("two creatures: the enumerator offered a flashback the view withholds")
	}

	token := push(me, "Goblin Token", "Token Creature — Goblin")
	c = castSurfaceOf(t, g, me.ID, me.Seat, id)
	if !c.CastableHere {
		t.Fatalf("three creatures: the flashback is not castable_here")
	}
	if !offeredFromGraveyard(g, me.ID, id) {
		t.Errorf("three creatures: the view offers a flashback the enumerator does not")
	}
	if len(c.AlternativeCosts) != 1 {
		t.Fatalf("offers = %+v, want the flashback alone", c.AlternativeCosts)
	}
	offer := c.AlternativeCosts[0]
	if offer.PayOptions != nil {
		t.Errorf("a sacrifice offer shipped pay_options %+v — the client would open the card picker, not the sacrifice one", offer.PayOptions)
	}
	opts := offer.SacrificeOptions
	if opts == nil {
		t.Fatalf("the offer carries no sacrifice_options — the client has nothing to pick from")
	}
	if opts.Min != 3 || opts.Max != 3 {
		t.Errorf("sacrifice_options count %d..%d, want 3..3", opts.Min, opts.Max)
	}
	want := map[string]bool{bear.String(): true, wolf.String(): true, token.String(): true}
	if len(opts.Cards) != 3 {
		t.Fatalf("sacrifice_options = %v, want the caster's three creatures", opts.Cards)
	}
	for _, cid := range opts.Cards {
		if cid == theirs.String() {
			t.Errorf("an opponent's creature is offered as payment — CR 701.21a")
		}
		if !want[cid] {
			t.Errorf("sacrifice_options carries %s, which the caster does not control", cid)
		}
	}
	if opts.Cards[0] != token.String() {
		t.Errorf("sacrifice_options lead with %s, want the token first (payment order, #747)", opts.Cards[0])
	}
	if offer.PayLabel != "three creatures" {
		t.Errorf("pay_label = %q, want the clause", offer.PayLabel)
	}

	// The bystander: the printed offer survives on a public pile, the
	// caster's own list does not.
	b := castSurfaceOf(t, g, them.ID, me.Seat, id)
	for _, o := range b.AlternativeCosts {
		if o.SacrificeOptions != nil {
			t.Errorf("a bystander got the caster's sacrifice_options: %+v", o.SacrificeOptions)
		}
	}
}
