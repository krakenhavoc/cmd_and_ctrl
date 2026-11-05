package heuristic_test

import (
	"context"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// haste_tap_test.go — #4324 (#4312's item 4): haste granted until end of
// turn lets a summoning-sick creature pay a {T} cost (CR 302.6, 702.10c),
// so a haste grant is also worth the {T} ability it unlocks.

// archivist is a seat-0 1/1 whose one row is "{T}: Draw a card." (a
// declared draw), summoning-sick unless opts say otherwise.
func archivist(id, cost string, opts ...cardOpt) protocol.CardView {
	c := creature(id, 0, "Archivist", 1, 1, append([]cardOpt{sick()}, opts...)...)
	c.ActivatedAbilities = []protocol.ActivatedAbilityView{{
		Index: 0, TapCost: true, ManaCost: cost, Label: cost + "{T}: Draw a card.",
		Purpose: &protocol.PurposeView{Draws: 1},
	}}
	return c
}

// expedite is Expedite as the catalog declares it: "Target creature gains
// haste until end of turn. Draw a card." for {R}.
func expedite(id string) protocol.CardView {
	c := spell(id, 0, "Expedite", "{R}")
	c.Purpose = &protocol.PurposeView{Draws: 1, Targets: &[]protocol.TargetPurposeView{{
		Slot: 0, UntilEndOfTurn: "benefit", Pump: &protocol.PumpView{Keywords: []string{"haste"}},
	}}}
	return c
}

func expediteMove(t *testing.T, id, target string) legal.Move {
	m := castMove(t, 0, id, "Cast Expedite", cardTarget(target))
	m.Cost = &legal.MoveCost{Mana: "{R}"}
	return m
}

func noHasteTaps() *heuristic.Policy {
	return withConfig(func(c *heuristic.Config) { c.PriceHasteTapRows = false })
}

// hasteBoard is the bot's precombat main with `lands` Mountains, the
// creatures given, and an opponent's untapped 3/3 that walls a 1/1's
// attack.
func hasteBoard(hand []protocol.CardView, lands int, mine ...protocol.CardView) protocol.GameView {
	bf := append([]protocol.CardView{}, mine...)
	for i := 0; i < lands; i++ {
		bf = append(bf, land(cardID(91+i), 0))
	}
	bf = append(bf, creature(cardID(2), 1, "Their Bear", 3, 3))
	return newView([]protocol.PlayerView{newSeat(0, withHand(hand...)), newSeat(1)},
		withBattlefield(bf...), withTurn(5, 0, "precombat_main"))
}

func castValue(t *testing.T, pol *heuristic.Policy, v protocol.GameView, moves []legal.Move, index int) float64 {
	t.Helper()
	for _, c := range pol.Rank(context.Background(), input(0, v, moves...)) {
		if c.Index == index {
			return c.Value
		}
	}
	t.Fatalf("move %d not ranked", index)
	return 0
}

// TestHasteGrantUnlocksATapAbility: the bot's fresh Archivist can't
// attack into the 3/3 and can't tap to draw. Expedite on it is its card
// back and the Archivist's draw this turn, and the bot casts it. Before
// #4324, and with PriceHasteTapRows off, the haste was worth only the
// attack it adds, nothing here, and the bot kept the card.
func TestHasteGrantUnlocksATapAbility(t *testing.T) {
	ex, arc := expedite(cardID(90)), cardID(3)
	v := hasteBoard([]protocol.CardView{ex}, 2, archivist(arc, ""))
	moves := []legal.Move{passMove(0), expediteMove(t, ex.InstanceID, arc)}
	if i, why := decideOn(t, heuristic.New(), v, moves...); i != 1 || !strings.Contains(why, "haste for a {T} ability") {
		t.Errorf("picked %d (%s), want Expedite on the Archivist for its draw", i, why)
	}
	if i, why := decideOn(t, noHasteTaps(), v, moves...); i != 0 {
		t.Errorf("with PriceHasteTapRows off the haste is meant to be worth only the attack: picked %d (%s)", i, why)
	}
}

// TestHasteTapNeedsTheRowsMana: the Archivist's row costs {2}. With two
// Mountains, Expedite's {R} leaves one, so the row cannot be paid this
// turn and the grant is priced as before; with three it can.
func TestHasteTapNeedsTheRowsMana(t *testing.T) {
	ex, arc := expedite(cardID(90)), cardID(3)
	moves := []legal.Move{passMove(0), expediteMove(t, ex.InstanceID, arc)}
	short := hasteBoard([]protocol.CardView{ex}, 2, archivist(arc, "{2}"))
	if on, off := castValue(t, heuristic.New(), short, moves, 1), castValue(t, noHasteTaps(), short, moves, 1); on != off {
		t.Errorf("with the row unaffordable the cast is priced %+.2f, want the old %+.2f", on, off)
	}
	enough := hasteBoard([]protocol.CardView{ex}, 3, archivist(arc, "{2}"))
	if i, why := decideOn(t, heuristic.New(), enough, moves...); i != 1 {
		t.Errorf("with the mana for the row the bot passed: picked %d (%s)", i, why)
	}
}

// TestHasteTapOnlyForASickCreature: an Archivist that has been on the
// battlefield since the turn began can already tap, and one with haste
// needs none, so the grant unlocks nothing on either.
func TestHasteTapOnlyForASickCreature(t *testing.T) {
	ex, arc := expedite(cardID(90)), cardID(3)
	moves := []legal.Move{passMove(0), expediteMove(t, ex.InstanceID, arc)}
	for name, c := range map[string]protocol.CardView{
		"not sick":       creatureWithRow(archivist(arc, ""), false),
		"already hasted": archivist(arc, "", keywords("haste")),
	} {
		t.Run(name, func(t *testing.T) {
			v := hasteBoard([]protocol.CardView{ex}, 2, c)
			if on, off := castValue(t, heuristic.New(), v, moves, 1), castValue(t, noHasteTaps(), v, moves, 1); on != off {
				t.Errorf("the cast is priced %+.2f, want the old %+.2f", on, off)
			}
		})
	}
}

func creatureWithRow(c protocol.CardView, sick bool) protocol.CardView {
	c.SummoningSick = sick
	return c
}

// TestHasteGrantStillPricesTheAttack: a fresh 5/5 with the same row, into
// an open board. Haste's attack is worth more than the draw, so the grant
// keeps the price it had: the {T} ability is the alternative to the
// attack, not an addition to it.
func TestHasteGrantStillPricesTheAttack(t *testing.T) {
	ex, big := expedite(cardID(90)), cardID(3)
	c := archivist(big, "")
	c.Power, c.Toughness = 5, 5
	v := newView([]protocol.PlayerView{newSeat(0, withHand(ex)), newSeat(1, withLife(10))},
		withBattlefield(c, land(cardID(91), 0), land(cardID(92), 0)), withTurn(5, 0, "precombat_main"))
	moves := []legal.Move{passMove(0), expediteMove(t, ex.InstanceID, big)}
	on, off := castValue(t, heuristic.New(), v, moves, 1), castValue(t, noHasteTaps(), v, moves, 1)
	if on != off {
		t.Errorf("the attack-priced grant moved from %+.2f to %+.2f", off, on)
	}
	if i, why := decideOn(t, heuristic.New(), v, moves...); i != 1 {
		t.Errorf("picked %d (%s), want Expedite for the 5/5's attack", i, why)
	}
}

// battlements is a seat-0 land with "{T}: Add {C}." and "{R}, {T}:
// Target creature gains haste until end of turn.", Hanweir Battlements'
// two rows.
func battlements(id string) protocol.CardView {
	c := land(id, 0)
	c.Name, c.TypeLine = "Hanweir Battlements", "Land"
	c.ManaAbilities = []protocol.ManaAbilityView{{Index: 0, TapCost: true, Produced: "{C}"}}
	c.ActivatedAbilities = []protocol.ActivatedAbilityView{{
		Index: 0, TapCost: true, ManaCost: "{R}", Label: "{R}, {T}: Target creature gains haste until end of turn.",
		Purpose: &protocol.PurposeView{Targets: &[]protocol.TargetPurposeView{{
			Slot: 0, UntilEndOfTurn: "benefit", Pump: &protocol.PumpView{Keywords: []string{"haste"}},
		}}},
	}}
	return c
}

// TestHasteRowCountsItsOwnTap: Hanweir Battlements' haste row taps the
// land itself, so its mana is not there for the Archivist's {1} row. With
// one Mountain beside it, the {R} and the tap leave nothing; with two,
// one is left and the bot hastes the Archivist.
func TestHasteRowCountsItsOwnTap(t *testing.T) {
	hb, arc := cardID(80), cardID(3)
	move := activateMove(t, 0, hb, "Hanweir Battlements: haste", nil, cardTarget(arc))
	moves := []legal.Move{passMove(0), move}
	short := hasteBoard(nil, 1, battlements(hb), archivist(arc, "{1}"))
	if i, why := decideOn(t, heuristic.New(), short, moves...); i != 0 {
		t.Errorf("with no mana left for the row the bot hasted anyway: picked %d (%s)", i, why)
	}
	enough := hasteBoard(nil, 2, battlements(hb), archivist(arc, "{1}"))
	if i, why := decideOn(t, heuristic.New(), enough, moves...); i != 1 || !strings.Contains(why, "haste for a {T} ability") {
		t.Errorf("picked %d (%s), want the haste row for the Archivist's draw", i, why)
	}
	if i, why := decideOn(t, noHasteTaps(), enough, moves...); i != 0 {
		t.Errorf("with PriceHasteTapRows off the row is meant to buy nothing here: picked %d (%s)", i, why)
	}
}
