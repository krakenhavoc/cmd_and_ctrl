package heuristic_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// purpose_test.go is ADR 0126 §4, §6's prices and the discard half of
// §7: a spell or a row priced by the purpose the catalog declares, a
// wipe by what it removes, and a discarded card by what it is worth.

func sorcery(id string, controller int, name, cost string, p *protocol.PurposeView) protocol.CardView {
	c := spell(id, controller, name, cost)
	c.TypeLine = "Sorcery"
	c.Purpose = p
	return c
}

func wrathOfGod(id string, controller int) protocol.CardView {
	return sorcery(id, controller, "Wrath of God", "{2}{W}{W}",
		&protocol.PurposeView{Sweep: &protocol.SweepView{Matches: "creatures", How: "destroy"}})
}

// bigArmy is n 4/4s for one seat.
func bigArmy(controller, n, from int) []protocol.CardView {
	out := make([]protocol.CardView, n)
	for i := range out {
		out[i] = creature(cardID(from+i), controller, "Beast", 4, 4)
	}
	return out
}

func wipeInput(t *testing.T, wrath protocol.CardView, bf []protocol.CardView, moves ...legal.Move) (protocol.GameView, []legal.Move) {
	t.Helper()
	v := newView([]protocol.PlayerView{newSeat(0, withHand(wrath)), newSeat(1), newSeat(2), newSeat(3)},
		withBattlefield(bf...), withTurn(5, 0, "precombat_main"))
	return v, append([]legal.Move{passMove(0)}, moves...)
}

// ADR 0126 §4: the wipe the bot used to cast onto its own winning
// board, and the one it should cast onto a losing one.
func TestAWipeIsPricedByWhatItRemoves(t *testing.T) {
	wrath := cardID(1)
	lands := manaLands(4, 0, 100)

	t.Run("do not wipe your own winning board", func(t *testing.T) {
		bf := append(append([]protocol.CardView{}, lands...), bigArmy(0, 3, 200)...)
		bf = append(bf, creature(cardID(300), 1, "Bear", 2, 2))
		v, moves := wipeInput(t, wrathOfGod(wrath, 0), bf, castMove(t, 0, wrath, "Cast Wrath of God"))
		in := input(0, v, moves...)
		if got := rankValue(t, heuristic.New(), in, "Cast Wrath of God"); got >= 0 {
			t.Errorf("a wipe of the bot's own three 4/4s priced %.3f, want below zero", got)
		}
		if d, _ := heuristic.New().Decide(context.Background(), in); in.Moves[d.Index].Kind != legal.KindPass {
			t.Errorf("the heuristic wiped its own board: %+v", d)
		}
		// The old proxy cast it: 0.60 × 4 − 1.20.
		base := heuristic.NewWithConfig(heuristic.BaselineConfig())
		if got := rankValue(t, base, in, "Cast Wrath of God"); !nearly(got, 0.6*4-1.2) {
			t.Errorf("baseline priced the wipe %.3f, want %.3f", got, 0.6*4-1.2)
		}
	})

	t.Run("wipe a losing board", func(t *testing.T) {
		bf := append(append([]protocol.CardView{}, lands...), bigArmy(1, 3, 200)...)
		bf = append(bf, bigArmy(2, 2, 300)...)
		v, moves := wipeInput(t, wrathOfGod(wrath, 0), bf, castMove(t, 0, wrath, "Cast Wrath of God"))
		in := input(0, v, moves...)
		got := rankValue(t, heuristic.New(), in, "Cast Wrath of God")
		if got <= 0.6*4-1.2 {
			t.Errorf("a wipe of five opposing 4/4s priced %.3f, want more than the old proxy", got)
		}
		if d, _ := heuristic.New().Decide(context.Background(), in); in.Moves[d.Index].Label != "Cast Wrath of God" {
			t.Errorf("the heuristic did not wipe a losing board: %+v", d)
		}
	})
}

// sweepPrice is the price of casting a one-mode sorcery with this sweep
// onto `bf`, against the same cast of a sweep that removes nothing.
func sweepPrice(t *testing.T, s protocol.SweepView, x int, bf []protocol.CardView) float64 {
	t.Helper()
	id := cardID(1)
	c := sorcery(id, 0, "Sweep", "{2}{W}{W}", &protocol.PurposeView{Sweep: &s})
	v, _ := wipeInput(t, c, bf)
	m := castMove(t, 0, id, "Cast Sweep")
	if x > 0 {
		m.Params = mustJSON(t, map[string]any{"instance_id": id, "from_zone": "hand", "x_value": x})
	}
	return rankValue(t, heuristic.New(), input(0, v, passMove(0), m), "Cast Sweep") + 1.2
}

// What a sweep spares, by how it removes and what it matches.
func TestASweepSparesWhatItDoesNotRemove(t *testing.T) {
	theirs := creature(cardID(200), 1, "Bear", 2, 2)
	god := creature(cardID(201), 1, "God", 6, 6, keywords("indestructible"))
	mine := creature(cardID(202), 0, "Elf", 2, 2)
	nothing := 0.0

	if got := sweepPrice(t, protocol.SweepView{Matches: "creatures", How: "destroy"}, 0, []protocol.CardView{god}); !nearly(got, nothing) {
		t.Errorf("destroy priced an indestructible creature at %.3f, want %.3f", got, nothing)
	}
	if got := sweepPrice(t, protocol.SweepView{Matches: "creatures", How: "exile"}, 0, []protocol.CardView{god}); got <= 0 {
		t.Errorf("exile spared an indestructible creature: %.3f", got)
	}
	if got := sweepPrice(t, protocol.SweepView{Matches: "creatures", How: "damage", Amount: 3}, 0, []protocol.CardView{god}); !nearly(got, nothing) {
		t.Errorf("3 damage priced a 6/6 at %.3f, want %.3f", got, nothing)
	}
	if got := sweepPrice(t, protocol.SweepView{Matches: "creatures", How: "damage", AmountIsX: true}, 0, []protocol.CardView{theirs}); !nearly(got, nothing) {
		t.Errorf("X=0 damage priced a 2/2 at %.3f, want %.3f", got, nothing)
	}
	if got := sweepPrice(t, protocol.SweepView{Matches: "creatures", How: "damage", AmountIsX: true}, 2, []protocol.CardView{theirs}); got <= 0 {
		t.Errorf("X=2 damage spared a 2/2: %.3f", got)
	}
	if got := sweepPrice(t, protocol.SweepView{Matches: "artifacts", How: "destroy"}, 0, []protocol.CardView{theirs}); !nearly(got, nothing) {
		t.Errorf("an artifact sweep priced a creature at %.3f", got)
	}
	both := []protocol.CardView{theirs, mine}
	all := sweepPrice(t, protocol.SweepView{Matches: "creatures", How: "destroy"}, 0, both)
	opp := sweepPrice(t, protocol.SweepView{Matches: "creatures", How: "destroy", OpponentsOnly: true}, 0, both)
	if opp <= all {
		t.Errorf("an opponents-only sweep (%.3f) is worth no more than one that also kills the bot's elf (%.3f)", opp, all)
	}
	full := sweepPrice(t, protocol.SweepView{Matches: "creatures", How: "destroy"}, 0, []protocol.CardView{theirs})
	bounce := sweepPrice(t, protocol.SweepView{Matches: "creatures", How: "bounce"}, 0, []protocol.CardView{theirs})
	partial := sweepPrice(t, protocol.SweepView{Matches: "creatures", How: "destroy", Partial: true}, 0, []protocol.CardView{theirs})
	if !(bounce > 0 && bounce < full) || !(partial > 0 && partial < full) {
		t.Errorf("bounce %.3f and partial %.3f should each be a share of the full sweep %.3f", bounce, partial, full)
	}
}

// ADR 0126 §4: a modal wipe is priced by the mode the move names, and
// an overloaded one by the cost it claims.
func TestAWipeIsPricedByTheModeOrCostItNames(t *testing.T) {
	id := cardID(1)
	rift := protocol.CardView{
		InstanceID: id, Name: "Cyclonic Rift", Owner: seatID(0).String(), Controller: seatID(0).String(),
		TypeLine: "Instant", ManaCost: "{1}{U}", KnownByYou: true,
		AlternativeCosts: []protocol.AlternativeCostView{{
			Key: "overload", ManaCost: "{6}{U}",
			Purpose: &protocol.PurposeView{Sweep: &protocol.SweepView{Matches: "nonland_permanents", How: "bounce", OpponentsOnly: true}},
		}},
	}
	bf := append(manaLands(7, 0, 100), bigArmy(1, 3, 200)...)
	v, _ := wipeInput(t, rift, bf)
	hard := castMove(t, 0, id, "Cast Cyclonic Rift")
	over := castMove(t, 0, id, "Cast Cyclonic Rift (overload)")
	over.Params = mustJSON(t, map[string]any{"instance_id": id, "from_zone": "hand", "alternative_cost": "overload"})
	in := input(0, v, passMove(0), hard, over)
	pol := heuristic.New()
	if h, o := rankValue(t, pol, in, hard.Label), rankValue(t, pol, in, over.Label); o <= h {
		t.Errorf("overload priced %.3f, no better than the hard cast %.3f", o, h)
	}

	farewell := protocol.CardView{
		InstanceID: id, Name: "Farewell", Owner: seatID(0).String(), Controller: seatID(0).String(),
		TypeLine: "Sorcery", ManaCost: "{4}{W}{W}", KnownByYou: true,
		Modes: &protocol.ModeSpecView{Min: 1, Max: 4, Options: []protocol.ModeOptionView{
			{Label: "Exile all artifacts.", Purpose: &protocol.PurposeView{Sweep: &protocol.SweepView{Matches: "artifacts", How: "exile"}}},
			{Label: "Exile all creatures.", Purpose: &protocol.PurposeView{Sweep: &protocol.SweepView{Matches: "creatures", How: "exile"}}},
		}},
	}
	v, _ = wipeInput(t, farewell, bf)
	arts := castMove(t, 0, id, "Cast Farewell (artifacts)")
	arts.Params = mustJSON(t, map[string]any{"instance_id": id, "from_zone": "hand", "modes": []int{0}})
	crits := castMove(t, 0, id, "Cast Farewell (creatures)")
	crits.Params = mustJSON(t, map[string]any{"instance_id": id, "from_zone": "hand", "modes": []int{1}})
	in = input(0, v, passMove(0), arts, crits)
	if a, c := rankValue(t, pol, in, arts.Label), rankValue(t, pol, in, crits.Label); !nearly(a, -1.2) || c <= 0 {
		t.Errorf("Farewell: the artifact mode priced %.3f (want −1.20, nothing to exile), the creature mode %.3f (want positive)", a, c)
	}
}

// ADR 0126 §6: the amounts, in place of the mana-value proxy.
func TestASpellIsPricedByItsDeclaredPurpose(t *testing.T) {
	pol := heuristic.New()
	cfg := pol.Config()
	price := func(c protocol.CardView, bf []protocol.CardView, hand ...protocol.CardView) float64 {
		t.Helper()
		v := newView([]protocol.PlayerView{newSeat(0, withHand(append([]protocol.CardView{c}, hand...)...)), newSeat(1)},
			withBattlefield(bf...), withTurn(3, 0, "precombat_main"))
		return rankValue(t, pol, input(0, v, passMove(0), castMove(t, 0, c.InstanceID, "Cast "+c.Name)), "Cast "+c.Name)
	}
	id := cardID(1)
	lands := manaLands(3, 0, 100)

	whisper := sorcery(id, 0, "Night's Whisper", "{1}{B}", &protocol.PurposeView{Draws: 2})
	if got, want := price(whisper, lands), 2*1.2-1.2; !nearly(got, want) {
		t.Errorf("Night's Whisper priced %.3f, want %.3f", got, want)
	}
	tutor := sorcery(id, 0, "Demonic Tutor", "{1}{B}", &protocol.PurposeView{Tutors: 1})
	if got, want := price(tutor, lands), 1.2*cfg.TutorWeight-1.2; !nearly(got, want) {
		t.Errorf("Demonic Tutor priced %.3f, want %.3f", got, want)
	}
	entomb := sorcery(id, 0, "Entomb", "{B}", &protocol.PurposeView{SelfMillTutor: 1})
	if got, want := price(entomb, lands), 1.2*cfg.SelfMillWeight-1.2; !nearly(got, want) {
		t.Errorf("Entomb priced %.3f, want %.3f", got, want)
	}
	looting := sorcery(id, 0, "Faithless Looting", "{R}", &protocol.PurposeView{Draws: 2, Discards: 2})
	if got, want := price(looting, lands), 2*1.2-2*cfg.DiscardWeight-1.2; !nearly(got, want) {
		t.Errorf("Faithless Looting priced %.3f, want %.3f", got, want)
	}

	// A ramp spell: a land onto the battlefield, plus PR 3's premium
	// while the bot is short of its hand.
	growth := sorcery(id, 0, "Rampant Growth", "{1}{G}", &protocol.PurposeView{Lands: 1})
	five := rampFiveDrop(cardID(9), 0)
	if got, want := price(growth, manaLands(2, 0, 100), five), 1.0+1.0-1.2; !nearly(got, want) {
		t.Errorf("Rampant Growth short of a five-drop priced %.3f, want %.3f", got, want)
	}
	if got, want := price(growth, manaLands(6, 0, 100), five), 1.0-1.2; !nearly(got, want) {
		t.Errorf("Rampant Growth with the five-drop covered priced %.3f, want %.3f", got, want)
	}

	// A permanent's declared enters effect is added to its body.
	elves := creature(id, 0, "Wood Elves", 1, 1)
	elves.ManaCost = "{2}{G}"
	elves.Purpose = &protocol.PurposeView{Lands: 1}
	plain := elves
	plain.Purpose = nil
	if with, without := price(elves, manaLands(6, 0, 100)), price(plain, manaLands(6, 0, 100)); !nearly(with-without, 1.0) {
		t.Errorf("Wood Elves' land added %.3f to its body, want 1.0", with-without)
	}

	// The baseline keeps the proxy.
	base := heuristic.NewWithConfig(heuristic.BaselineConfig())
	v := newView([]protocol.PlayerView{newSeat(0, withHand(tutor)), newSeat(1)}, withBattlefield(lands...), withTurn(3, 0, "precombat_main"))
	if got := rankValue(t, base, input(0, v, passMove(0), castMove(t, 0, id, "Cast Demonic Tutor")), "Cast Demonic Tutor"); !nearly(got, 0.6*2-1.2) {
		t.Errorf("baseline priced Demonic Tutor %.3f, want the proxy %.3f", got, 0.6*2-1.2)
	}
}

// ADR 0126 §6: a row of the bot's own that declares a purpose is priced
// by it, and pays for its source when it sacrifices it.
func TestAnActivatedRowIsPricedByItsDeclaredPurpose(t *testing.T) {
	pol := heuristic.New()
	src := rock(cardID(1), 0, "Wayfarer's Bauble", "{1}", protocol.ManaAbilityView{})
	src.ManaAbilities = nil
	src.ActivatedAbilities = []protocol.ActivatedAbilityView{{
		Index: 0, TapCost: true, SacrificeSelf: true, ManaCost: "{2}",
		Purpose: &protocol.PurposeView{Lands: 1},
	}}
	act := legal.Move{
		Type: legal.TypeActivateAbility, Player: seatID(0), Kind: legal.KindActivate,
		Label: "Activate Wayfarer's Bauble", Source: uuid.MustParse(src.InstanceID),
		Params: mustJSON(t, map[string]any{"source_card_id": src.InstanceID, "ability_index": 0}),
	}
	five := rampFiveDrop(cardID(9), 0)
	at := func(lands int) float64 {
		bf := append(manaLands(lands, 0, 100), src)
		v := newView([]protocol.PlayerView{newSeat(0, withHand(five)), newSeat(1)}, withBattlefield(bf...), withTurn(3, 0, "precombat_main"))
		return rankValue(t, pol, input(0, v, passMove(0), act), act.Label)
	}
	// A land (1.0) and the premium (1.0) for the Bauble (1.2).
	if got, want := at(2), 1.0+1.0-1.2; !nearly(got, want) {
		t.Errorf("the Bauble short of a five-drop priced %.3f, want %.3f", got, want)
	}
	if got, want := at(6), 1.0-1.2; !nearly(got, want) {
		t.Errorf("the Bauble with the five-drop covered priced %.3f, want %.3f", got, want)
	}
}

// A declared loot on a creature that could attack costs that attack
// before combat, so the bot swings with its commander rather than
// tapping it to loot in its first main phase.
func TestALootDoesNotReplaceTheAttack(t *testing.T) {
	mary := creature(cardID(1), 0, "Mary Read and Anne Bonny", 3, 3)
	mary.ActivatedAbilities = []protocol.ActivatedAbilityView{{
		Index: 0, TapCost: true, Purpose: &protocol.PurposeView{Draws: 1, Discards: 1},
	}}
	loot := legal.Move{
		Type: legal.TypeActivateAbility, Player: seatID(0), Kind: legal.KindActivate,
		Label: "Loot", Source: uuid.MustParse(mary.InstanceID),
		Params: mustJSON(t, map[string]any{"source_card_id": mary.InstanceID, "ability_index": 0}),
	}
	at := func(step string) float64 {
		v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(mary), withTurn(5, 0, step))
		return rankValue(t, heuristic.New(), input(0, v, passMove(0), loot), "Loot")
	}
	if got, want := at("postcombat_main"), 1.2-0.6-0.3; !nearly(got, want) {
		t.Errorf("the loot after combat priced %.3f, want %.3f", got, want)
	}
	if got := at("precombat_main"); got >= 0 {
		t.Errorf("the loot before combat priced %.3f, want below zero: it costs the 3/3's attack", got)
	}
}

// ADR 0126 §7: Unexpected Windfall pays what the discarded card is
// worth, so the spare land goes late and the last land stays early.
func TestADiscardCostIsWhatTheCardIsWorth(t *testing.T) {
	pol := heuristic.New()
	windfall := spell(cardID(1), 0, "Unexpected Windfall", "{2}{R}{R}")
	windfall.Purpose = &protocol.PurposeView{Draws: 2, Tokens: 2}
	windfall.AdditionalCost = &protocol.AdditionalCostView{DiscardCards: 1}
	cast := func(discard string) legal.Move {
		m := castMove(t, 0, windfall.InstanceID, "Cast Unexpected Windfall discarding "+discard)
		m.Params = mustJSON(t, map[string]any{"instance_id": windfall.InstanceID, "from_zone": "hand", "discard_ids": []string{discard}})
		return m
	}
	spare, other := land(cardID(2), 0), land(cardID(3), 0)
	bolt := spell(cardID(4), 0, "Lightning Bolt", "{R}")

	t.Run("late, a spare land", func(t *testing.T) {
		v := newView([]protocol.PlayerView{newSeat(0, withHand(windfall, spare, other, bolt)), newSeat(1)},
			withBattlefield(manaLands(7, 0, 100)...), withTurn(9, 0, "precombat_main"))
		in := input(0, v, passMove(0), cast(spare.InstanceID), cast(bolt.InstanceID))
		d, _ := pol.Decide(context.Background(), in)
		if in.Moves[d.Index].Label != "Cast Unexpected Windfall discarding "+spare.InstanceID {
			t.Errorf("late, the heuristic did not pitch the spare land: %+v", d)
		}
	})
	t.Run("early, the last land", func(t *testing.T) {
		v := newView([]protocol.PlayerView{newSeat(0, withHand(windfall, spare)), newSeat(1)},
			withBattlefield(manaLands(4, 0, 100)...), withTurn(4, 0, "precombat_main"))
		in := input(0, v, passMove(0), cast(spare.InstanceID))
		if d, _ := pol.Decide(context.Background(), in); in.Moves[d.Index].Kind != legal.KindPass {
			t.Errorf("early, the heuristic pitched its last land: %+v", d)
		}
	})
	t.Run("baseline pays a flat card", func(t *testing.T) {
		v := newView([]protocol.PlayerView{newSeat(0, withHand(windfall, spare, other)), newSeat(1)},
			withBattlefield(manaLands(7, 0, 100)...), withTurn(9, 0, "precombat_main"))
		in := input(0, v, passMove(0), cast(spare.InstanceID))
		base := heuristic.NewWithConfig(heuristic.BaselineConfig())
		if got, want := rankValue(t, base, in, in.Moves[1].Label), 0.6*4-1.2-1.2; !nearly(got, want) {
			t.Errorf("baseline priced Windfall %.3f, want %.3f", got, want)
		}
	})
}
