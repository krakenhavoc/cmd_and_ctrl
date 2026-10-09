package heuristic_test

import (
	"context"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// plan_test.go is ADR 0136 PR 4: the turn plan, on boards shaped like
// the review game's evidence windows (game 8a9f18d7, seq 133 and 402)
// and the cantrip position. Each test fails with Config.PlanTurnMana
// off, which decides one move at a time.

func planForest(id string) protocol.CardView {
	c := land(id, 0)
	c.Name, c.TypeLine = "Forest", "Basic Land — Forest"
	c.ManaAbilities = []protocol.ManaAbilityView{{TapCost: true, Produced: "{G}"}}
	return c
}

func planIsland(id string) protocol.CardView {
	c := land(id, 0)
	c.Name, c.TypeLine = "Island", "Basic Land — Island"
	c.ManaAbilities = []protocol.ManaAbilityView{{TapCost: true, Produced: "{U}"}}
	return c
}

func planSwamp(id string) protocol.CardView {
	c := land(id, 0)
	c.Name, c.TypeLine = "Swamp", "Basic Land — Swamp"
	c.ManaAbilities = []protocol.ManaAbilityView{{TapCost: true, Produced: "{B}"}}
	return c
}

// stampedCast is castMove with the total mana the enumerator stamps on
// a cast (legal.MoveCost.Mana, ADR 0136 owner answer 1).
func stampedCast(t *testing.T, id, label, mana string) legal.Move {
	m := castMove(t, 0, id, label)
	m.Cost = &legal.MoveCost{Mana: mana}
	return m
}

func decideBoth(t *testing.T, in aiseat.Input) (on, off aiseat.Decision, plan []aiseat.PlanMember) {
	t.Helper()
	d, tr, err := heuristic.New().DecideTraced(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	cfg := heuristic.DefaultConfig()
	cfg.PlanTurnMana = false
	o, otr, err := heuristic.NewWithConfig(cfg).DecideTraced(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	if len(otr.Plan) != 0 {
		t.Errorf("with the plan off the trace carries a plan: %+v", otr.Plan)
	}
	return d, o, tr.Plan
}

func moveLabel(in aiseat.Input, d aiseat.Decision) string {
	if d.Index < 0 || d.Index >= len(in.Moves) {
		return "(none)"
	}
	return in.Moves[d.Index].Label
}

// TestPlanCastsTheRockFirst is seq 133 (`rock-before-the-two-drop`):
// three lands, Arcane Signet and a two-mana dork in hand. One move at a
// time the dork wins and the Signet waits a turn; the plan casts the
// Signet, then the dork on the Signet's mana.
func TestPlanCastsTheRockFirst(t *testing.T) {
	signet := protocol.CardView{InstanceID: cardID(10), Name: "Arcane Signet", Owner: seatID(0).String(), Controller: seatID(0).String(),
		TypeLine: "Artifact", ManaCost: "{2}", KnownByYou: true,
		ManaAbilities: []protocol.ManaAbilityView{{TapCost: true, Produced: "{W|U|B|R|G}"}}}
	dork := creature(cardID(11), 0, "Ornithopter of Paradise", 2, 1, keywords("flying"))
	dork.TypeLine, dork.ManaCost = "Artifact Creature — Thopter", "{2}"
	dork.ManaAbilities = []protocol.ManaAbilityView{{TapCost: true, Produced: "{W|U|B|R|G}"}}
	big := creature(cardID(12), 0, "Avenger of Zendikar", 5, 5)
	big.ManaCost = "{5}{G}{G}"
	orchard := land(cardID(3), 0)
	orchard.Name, orchard.TypeLine = "Exotic Orchard", "Land"
	orchard.ManaAbilities = []protocol.ManaAbilityView{{TapCost: true, ColorOptions: [][]string{{"G", "U"}}}}
	field := land(cardID(2), 0)
	field.Name, field.TypeLine = "Field of the Dead", "Land"
	field.ManaAbilities = []protocol.ManaAbilityView{{TapCost: true, Produced: "{C}"}}
	v := newView(
		[]protocol.PlayerView{newSeat(0, withHand(signet, dork, big)), newSeat(1)},
		withTurn(3, 0, "precombat_main"),
		withBattlefield(planIsland(cardID(1)), field, orchard),
	)
	in := input(0, v,
		passMove(0),
		stampedCast(t, cardID(11), "Cast Ornithopter of Paradise", "{2}"),
		stampedCast(t, cardID(10), "Cast Arcane Signet", "{2}"),
	)
	on, off, plan := decideBoth(t, in)
	if got := moveLabel(in, off); got != "Cast Ornithopter of Paradise" {
		t.Fatalf("with the plan off the bot casts %q; this board no longer shows the problem", got)
	}
	if got := moveLabel(in, on); got != "Cast Arcane Signet" {
		t.Errorf("the plan casts %q (%s), want the Signet first", got, on.Reason)
	}
	if len(plan) != 2 || plan[0].Label != "Cast Arcane Signet" || plan[1].Label != "Cast Ornithopter of Paradise" {
		t.Errorf("trace plan = %+v, want Signet then Ornithopter", plan)
	}
	if !strings.HasPrefix(on.Reason, "plan: Arcane Signet → Ornithopter of Paradise (+") {
		t.Errorf("reason = %q", on.Reason)
	}

	// Two lands: the Signet leaves one mana and the dork needs two.
	v.Battlefield.Cards = v.Battlefield.Cards[:2]
	in.View = v
	if on, _, plan := decideBoth(t, in); moveLabel(in, on) != "Cast Ornithopter of Paradise" || len(plan) != 0 {
		t.Errorf("with two lands the bot casts %q with plan %+v, want the dork alone", moveLabel(in, on), plan)
	}
}

// TestPlanTakesTwoSpellsOverTheTaxedCommander is seq 402
// (`two-spells-over-the-taxed-commander`): eight lands, the commander
// with {2} of tax, and a four-drop and a draw spell in hand. The
// commander's seven mana buys more as the two spells, and the draw
// spell goes first (§4).
func TestPlanTakesTwoSpellsOverTheTaxedCommander(t *testing.T) {
	in := taxedCommanderBoard(t, nil, 5, 3)
	on, off, plan := decideBoth(t, in)
	if got := moveLabel(in, off); got != "Cast Tatyova, Benthic Druid from the command zone" {
		t.Fatalf("with the plan off the bot casts %q; this board no longer shows the problem", got)
	}
	// Oracle declares no purpose here, so it is not ordered as ramp
	// (Config.PlanLandDropsAsRamp reads the purpose, never the name):
	// the draw goes first.
	if got := moveLabel(in, on); got != "Cast Harmonize" {
		t.Errorf("the plan casts %q (%s), want Harmonize first", got, on.Reason)
	}
	if len(plan) != 2 || plan[0].Label != "Cast Harmonize" || plan[1].Label != "Cast Oracle of Mul Daya" {
		t.Errorf("trace plan = %+v, want Harmonize then Oracle", plan)
	}

	// Seven lands: the two spells need eight, and the commander is
	// still the one cast.
	in = taxedCommanderBoard(t, nil, 5, 2)
	if on, _, _ := decideBoth(t, in); moveLabel(in, on) != "Cast Tatyova, Benthic Druid from the command zone" {
		t.Errorf("with seven lands the bot casts %q (%s), want the commander", moveLabel(in, on), on.Reason)
	}
}

// TestPlanCastsTheExtraLandDropBeforeTheDraw is seq 402 as the catalog
// declares Oracle of Mul Daya today (purpose.extra_land_drops 1): the
// owner's decision of 2026-10-08 (ADR 0136 §4, amendment) orders a
// permanent that grants extra land drops with the mana members, so
// Oracle is cast first and a land Harmonize draws can still be played
// this turn. It fails with Config.PlanLandDropsAsRamp off.
//
// Six Forests and two Islands: Oracle's {3}{G} leaves {G}{G} for
// Harmonize however the auto-tapper pays it.
func TestPlanCastsTheExtraLandDropBeforeTheDraw(t *testing.T) {
	in := taxedCommanderBoard(t, &protocol.PurposeView{ExtraLandDrops: 1}, 6, 2)
	d, tr, err := heuristic.New().DecideTraced(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	if got := moveLabel(in, d); got != "Cast Oracle of Mul Daya" {
		t.Errorf("the plan casts %q (%s), want Oracle of Mul Daya first", got, d.Reason)
	}
	if len(tr.Plan) != 2 || tr.Plan[0].Label != "Cast Oracle of Mul Daya" || tr.Plan[1].Label != "Cast Harmonize" {
		t.Errorf("trace plan = %+v, want Oracle then Harmonize", tr.Plan)
	}

	cfg := heuristic.DefaultConfig()
	cfg.PlanLandDropsAsRamp = false
	off, otr, err := heuristic.NewWithConfig(cfg).DecideTraced(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	if got := moveLabel(in, off); got != "Cast Harmonize" {
		t.Errorf("with PlanLandDropsAsRamp off the plan casts %q (%s), want Harmonize first", got, off.Reason)
	}
	if len(otr.Plan) != 2 {
		t.Errorf("with PlanLandDropsAsRamp off the plan is %+v, want the same two members", otr.Plan)
	}

	// Five Forests and three Islands: the auto-tapper may pay Oracle's
	// generic with Forests and leave Harmonize one {G} short, so the
	// amended order cannot be promised. The set is still the best one,
	// so it falls back to the order before the amendment: the draw
	// first, then Oracle, rather than the commander alone.
	in = taxedCommanderBoard(t, &protocol.PurposeView{ExtraLandDrops: 1}, 5, 3)
	d, tr, err = heuristic.New().DecideTraced(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	if got := moveLabel(in, d); got != "Cast Harmonize" || len(tr.Plan) != 2 {
		t.Errorf("on five Forests the plan casts %q (%s) with plan %+v, want Harmonize then Oracle", got, d.Reason, tr.Plan)
	}
}

// taxedCommanderBoard is seq 402's board: `forests` Forests and
// `islands` Islands, Tatyova in the command zone with {2} of tax, and
// Oracle of Mul Daya (with oraclePurpose) and Harmonize in hand.
func taxedCommanderBoard(t *testing.T, oraclePurpose *protocol.PurposeView, forests, islands int) aiseat.Input {
	t.Helper()
	cmdr := creature(cardID(999), 0, "Tatyova, Benthic Druid", 3, 3, commander())
	cmdr.ManaCost = "{3}{G}{U}"
	cmdr.TypeLine = "Legendary Creature — Merfolk Druid"
	cmdr.AbilityRows = []protocol.AbilityRowView{{Kind: "triggered", Label: "Landfall"}}
	oracle := creature(cardID(20), 0, "Oracle of Mul Daya", 2, 2)
	oracle.ManaCost = "{3}{G}"
	oracle.AbilityRows = []protocol.AbilityRowView{{Kind: "static", Label: "You may play an additional land on each of your turns."}, {Kind: "static", Label: "Play with the top card of your library revealed."}}
	oracle.Purpose = oraclePurpose
	harmonize := protocol.CardView{InstanceID: cardID(21), Name: "Harmonize", Owner: seatID(0).String(), Controller: seatID(0).String(),
		TypeLine: "Sorcery", ManaCost: "{2}{G}{G}", KnownByYou: true, Purpose: &protocol.PurposeView{Draws: 3}}
	me := newSeat(0, withHand(oracle, harmonize), withCommanderCasts(1))
	me.Command = protocol.ZoneView{Kind: "command", Count: 1, Cards: []protocol.CardView{cmdr}}
	var bf []protocol.CardView
	for i := 0; i < forests; i++ {
		bf = append(bf, planForest(cardID(100+i)))
	}
	for i := 0; i < islands; i++ {
		bf = append(bf, planIsland(cardID(200+i)))
	}
	v := newView([]protocol.PlayerView{me, newSeat(1)}, withTurn(8, 0, "precombat_main"), withBattlefield(bf...))
	tatyova := stampedCast(t, cardID(999), "Cast Tatyova, Benthic Druid from the command zone", "{5}{G}{U}")
	tatyova.Params = mustJSON(t, map[string]any{"instance_id": cardID(999), "from_zone": "command"})
	return input(0, v,
		passMove(0),
		stampedCast(t, cardID(20), "Cast Oracle of Mul Daya", "{3}{G}"),
		stampedCast(t, cardID(21), "Cast Harmonize", "{2}{G}{G}"),
		tatyova,
	)
}

// TestPlanDrawsBeforeThePermanent is `cantrip-before-the-permanent`
// (owner answer 10): five Swamps in the second main phase, Night's
// Whisper with its declared purpose, and two permanents. The plan
// casts Night's Whisper and a permanent with the five mana, the draw
// first.
func TestPlanDrawsBeforeThePermanent(t *testing.T) {
	whisper := protocol.CardView{InstanceID: cardID(30), Name: "Night's Whisper", Owner: seatID(0).String(), Controller: seatID(0).String(),
		TypeLine: "Sorcery", ManaCost: "{1}{B}", KnownByYou: true, Purpose: &protocol.PurposeView{Draws: 2}}
	bastion := protocol.CardView{InstanceID: cardID(31), Name: "Bastion of Remembrance", Owner: seatID(0).String(), Controller: seatID(0).String(),
		TypeLine: "Enchantment", ManaCost: "{2}{B}", KnownByYou: true,
		AbilityRows: []protocol.AbilityRowView{{Kind: "triggered", Label: "When this enters, create a 1/1 Human Soldier."}, {Kind: "triggered", Label: "Whenever a creature you control dies, drain 1."}}}
	blood := protocol.CardView{InstanceID: cardID(32), Name: "Exquisite Blood", Owner: seatID(0).String(), Controller: seatID(0).String(),
		TypeLine: "Enchantment", ManaCost: "{4}{B}", KnownByYou: true,
		AbilityRows: []protocol.AbilityRowView{{Kind: "triggered", Label: "Whenever an opponent loses life, you gain that much life."}}}
	var lands []protocol.CardView
	for i := 0; i < 5; i++ {
		lands = append(lands, planSwamp(cardID(100+i)))
	}
	v := newView([]protocol.PlayerView{newSeat(0, withHand(whisper, bastion, blood)), newSeat(1)},
		withTurn(6, 0, "postcombat_main"), withBattlefield(lands...))
	in := input(0, v,
		passMove(0),
		stampedCast(t, cardID(32), "Cast Exquisite Blood", "{4}{B}"),
		stampedCast(t, cardID(31), "Cast Bastion of Remembrance", "{2}{B}"),
		stampedCast(t, cardID(30), "Cast Night's Whisper", "{1}{B}"),
	)
	on, off, plan := decideBoth(t, in)
	if moveLabel(in, off) == "Cast Night's Whisper" {
		t.Fatalf("with the plan off the bot already casts Night's Whisper; this board no longer shows the problem")
	}
	if got := moveLabel(in, on); got != "Cast Night's Whisper" {
		t.Errorf("the plan casts %q (%s), want Night's Whisper first", got, on.Reason)
	}
	if len(plan) != 2 || plan[0].Label != "Cast Night's Whisper" {
		t.Errorf("trace plan = %+v, want Night's Whisper and one permanent", plan)
	}
}

// TestPlanLeavesTheLandDropAlone is §1: the land drop is played first,
// and the next window plans the rest.
func TestPlanLeavesTheLandDropAlone(t *testing.T) {
	a := creature(cardID(40), 0, "Bear", 2, 2)
	a.ManaCost = "{1}{G}"
	b := creature(cardID(41), 0, "Other Bear", 2, 2)
	b.ManaCost = "{1}{G}"
	l := planForest(cardID(42))
	v := newView([]protocol.PlayerView{newSeat(0, withHand(a, b, l)), newSeat(1)},
		withTurn(4, 0, "precombat_main"),
		withBattlefield(planForest(cardID(1)), planForest(cardID(2)), planForest(cardID(3)), planForest(cardID(4))))
	in := input(0, v,
		passMove(0),
		landMove(t, 0, cardID(42)),
		stampedCast(t, cardID(40), "Cast Bear", "{1}{G}"),
		stampedCast(t, cardID(41), "Cast Other Bear", "{1}{G}"),
	)
	d, tr, err := heuristic.New().DecideTraced(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	if in.Moves[d.Index].Kind != legal.KindLand || len(tr.Plan) != 0 {
		t.Errorf("took %q with plan %+v, want the land drop and no plan", in.Moves[d.Index].Label, tr.Plan)
	}
}

// TestPlanOnlyInTheSeatsOwnMainPhase is §1: nowhere else.
func TestPlanOnlyInTheSeatsOwnMainPhase(t *testing.T) {
	a := spell(cardID(50), 0, "Bolt A", "{R}")
	a.TargetMode = "any"
	b := spell(cardID(51), 0, "Bolt B", "{R}")
	b.TargetMode = "any"
	v := newView([]protocol.PlayerView{newSeat(0, withHand(a, b)), newSeat(1, withLife(3))},
		withTurn(4, 1, "end"),
		withBattlefield(land(cardID(1), 0), land(cardID(2), 0)))
	in := input(0, v,
		passMove(0),
		castMove(t, 0, cardID(50), "Cast Bolt A", playerTarget(1)),
		castMove(t, 0, cardID(51), "Cast Bolt B", playerTarget(1)),
	)
	if _, tr, err := heuristic.New().DecideTraced(context.Background(), in); err != nil || len(tr.Plan) != 0 {
		t.Errorf("a plan outside the seat's own main phase: %+v (%v)", tr.Plan, err)
	}
}
