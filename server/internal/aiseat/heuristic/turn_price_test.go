package heuristic_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// turn_price_test.go — ADR 0188 Decision 10, #3566: a whole turn gained
// (CR 500.7) or lost (CR 614.10) is priced at Weights.Turn, from the
// purpose the catalog declares. Meditate is not in the catalog yet (the
// ADR's slice 7), so its purpose is declared on a fixture, exactly as
// the slice will.

func timeWarp(id string, controller int) protocol.CardView {
	return sorcery(id, controller, "Time Warp", "{3}{U}{U}",
		targetEntries(protocol.TargetPurposeView{Slot: 0, ExtraTurns: 1}))
}

func temporalManipulation(id string, controller int) protocol.CardView {
	return sorcery(id, controller, "Temporal Manipulation", "{3}{U}{U}", &protocol.PurposeView{ExtraTurns: 1})
}

// meditate is "{2}{U} Draw four cards. You skip your next turn."
func meditate(id string, controller int) protocol.CardView {
	return sorcery(id, controller, "Meditate", "{2}{U}", &protocol.PurposeView{Draws: 4, SkipsTurns: 1})
}

func evenBoard(hand protocol.CardView, lands int) protocol.GameView {
	return newView([]protocol.PlayerView{newSeat(0, withHand(hand)), newSeat(1), newSeat(2), newSeat(3)},
		withBattlefield(manaLands(lands, 0, 100)...), withTurn(6, 0, "precombat_main"))
}

// A Time Warp is worth a whole turn at the bot and a whole turn LOST at
// an opponent, so the bot aims it at itself and casts it.
func TestTimeWarpIsAimedAtTheBotAndCast(t *testing.T) {
	id := cardID(1)
	v := evenBoard(timeWarp(id, 0), 5)
	me := castMove(t, 0, id, "Time Warp at me", playerTarget(0))
	them := castMove(t, 0, id, "Time Warp at them", playerTarget(1))
	in := input(0, v, passMove(0), me, them)

	pol := heuristic.New()
	w := pol.Config().Weights
	if w.Turn <= 0 {
		t.Fatalf("Weights.Turn = %v, want a positive price", w.Turn)
	}
	// A whole turn at the bot, less the card the spell spends.
	if got, want := rankValue(t, pol, in, me.Label), w.Turn-w.Hand; !nearly(got, want) {
		t.Errorf("Time Warp at the bot priced %.3f, want %.3f", got, want)
	}
	if m, o := rankValue(t, pol, in, me.Label), rankValue(t, pol, in, them.Label); m <= o {
		t.Errorf("Time Warp at the bot %.3f is no better than at an opponent %.3f", m, o)
	}
	if o := rankValue(t, pol, in, them.Label); o >= 0 {
		t.Errorf("Time Warp at an opponent priced %.3f, want below zero", o)
	}
	d, err := pol.Decide(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	if got := in.Moves[d.Index].Label; got != me.Label {
		t.Errorf("the bot chose %q, want %q", got, me.Label)
	}
}

// The declared turn replaces the mana-value proxy, and the baseline keeps
// the proxy.
func TestAnExtraTurnSpellIsPricedByAWholeTurn(t *testing.T) {
	id := cardID(1)
	v := evenBoard(temporalManipulation(id, 0), 5)
	cast := castMove(t, 0, id, "Cast Temporal Manipulation")
	in := input(0, v, passMove(0), cast)

	pol := heuristic.New()
	w := pol.Config().Weights
	if got, want := rankValue(t, pol, in, cast.Label), w.Turn-w.Hand; !nearly(got, want) {
		t.Errorf("Temporal Manipulation priced %.3f, want a turn less the card %.3f", got, want)
	}
	base := heuristic.NewWithConfig(heuristic.BaselineConfig())
	if got, want := rankValue(t, base, in, cast.Label), 0.6*5-1.2; !nearly(got, want) {
		t.Errorf("baseline priced it %.3f, want the proxy %.3f", got, want)
	}
	if bw := heuristic.BaselineConfig().Weights.Turn; bw != 0 {
		t.Errorf("BaselineConfig.Weights.Turn = %v, want 0 (turns priced at nothing)", bw)
	}
}

// Meditate draws four cards and skips a turn. At an even board the turn
// is worth more than the cards, so the bot holds it.
func TestMeditateIsNotCastAtAnEvenBoard(t *testing.T) {
	id := cardID(1)
	v := evenBoard(meditate(id, 0), 5)
	cast := castMove(t, 0, id, "Cast Meditate")
	in := input(0, v, passMove(0), cast)

	pol := heuristic.New()
	w := pol.Config().Weights
	if got, want := rankValue(t, pol, in, cast.Label), 4*w.Hand-w.Turn-w.Hand; !nearly(got, want) {
		t.Errorf("Meditate priced %.3f, want four cards less a turn less the card %.3f", got, want)
	}
	d, err := pol.Decide(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	if in.Moves[d.Index].Kind != legal.KindPass {
		t.Errorf("the bot cast Meditate at an even board: %+v", d)
	}
	// Without the price (the baseline) the four cards win, which is the
	// bug the weight exists to fix.
	base := heuristic.NewWithConfig(heuristic.BaselineConfig())
	if got := rankValue(t, base, in, cast.Label); got <= rankValue(t, pol, in, cast.Label) {
		t.Errorf("baseline prices Meditate %.3f, no higher than with the turn charged", got)
	}
}

// A land that skips its controller's next turn for a counter is never
// activated, and the row that returns it for an extra turn is worth one.
func TestAnActivatedTurnRowIsPricedByTheTurn(t *testing.T) {
	src := rock(cardID(1), 0, "Magosi, the Waterveil", "", protocol.ManaAbilityView{})
	src.ManaAbilities = nil
	src.TypeLine = "Land"
	src.ActivatedAbilities = []protocol.ActivatedAbilityView{
		{Index: 0, TapCost: true, ManaCost: "{U}", Purpose: &protocol.PurposeView{SkipsTurns: 1}},
		{Index: 1, TapCost: true, Purpose: &protocol.PurposeView{ExtraTurns: 1}},
	}
	move := func(i int, label string) legal.Move {
		return legal.Move{
			Type: legal.TypeActivateAbility, Player: seatID(0), Kind: legal.KindActivate,
			Label: label, Source: uuid.MustParse(src.InstanceID),
			Params: mustJSON(t, map[string]any{"source_card_id": src.InstanceID, "ability_index": i}),
		}
	}
	skip, extra := move(0, "Magosi: skip"), move(1, "Magosi: extra turn")
	bf := append(manaLands(2, 0, 100), src)
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(bf...), withTurn(6, 0, "precombat_main"))
	in := input(0, v, passMove(0), skip, extra)

	pol := heuristic.New()
	w := pol.Config().Weights
	if got, want := rankValue(t, pol, in, skip.Label), -w.Turn; got > want+1e-6 {
		t.Errorf("the skip row priced %.3f, want at most a turn lost %.3f", got, want)
	}
	if got := rankValue(t, pol, in, extra.Label); got < w.Turn-1e-6 {
		t.Errorf("the extra-turn row priced %.3f, want at least a turn %.3f", got, w.Turn)
	}
}
