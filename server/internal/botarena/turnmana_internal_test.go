package botarena

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// turnmana_internal_test.go pins ADR 0136 §8's two tallies on
// hand-built observer feeds: what counts as mana left, which pass a
// turn is read at, and when a plan's next cast counts as missed.

func forest(id string, me string, tapped bool) protocol.CardView {
	return protocol.CardView{InstanceID: id, Name: "Forest", TypeLine: "Basic Land — Forest", Controller: me, Tapped: tapped,
		ManaAbilities: []protocol.ManaAbilityView{{TapCost: true, Produced: "{G}"}}}
}

func TestManaLeftCountsWhatTheSeatCanStillMake(t *testing.T) {
	me, them := uuid.NewString(), uuid.NewString()
	v := protocol.GameView{
		Seats: []protocol.PlayerView{{ID: me, ManaPool: []string{"G"}}, {ID: them, ManaPool: []string{"R", "R"}}},
		Battlefield: protocol.ZoneView{Cards: []protocol.CardView{
			forest("a", me, false),
			forest("b", me, true), // tapped
			forest("c", them, false),
			{InstanceID: "d", Name: "Sol Ring", TypeLine: "Artifact", Controller: me,
				ManaAbilities: []protocol.ManaAbilityView{{TapCost: true, Produced: "{C}{C}"}}},
			// A filter nets what it makes less what it takes.
			{InstanceID: "e", Name: "Azorius Signet", TypeLine: "Artifact", Controller: me,
				ManaAbilities: []protocol.ManaAbilityView{{TapCost: true, ManaCost: "{1}", Produced: "{W}{U}"}}},
			// CR 302.6: a sick dork adds nothing this turn.
			{InstanceID: "f", Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid", Controller: me, SummoningSick: true,
				ManaAbilities: []protocol.ManaAbilityView{{TapCost: true, Produced: "{G}"}}},
			// One-shot sources are not repeatable.
			{InstanceID: "g", Name: "Treasure", TypeLine: "Token Artifact — Treasure", Controller: me,
				ManaAbilities: []protocol.ManaAbilityView{{TapCost: true, SacrificeCost: true, Produced: "{W|U|B|R|G}"}}},
			// A row the view greys out.
			{InstanceID: "h", Name: "Cabal Coffers", TypeLine: "Land", Controller: me,
				ManaAbilities: []protocol.ManaAbilityView{{TapCost: true, ConditionUnmet: true}}},
			// A land with no rows on the view counts one.
			{InstanceID: "i", Name: "Mystery Land", TypeLine: "Land", Controller: me},
			// A counted choice: three of one colour.
			{InstanceID: "j", Name: "Gilded Lotus", TypeLine: "Artifact", Controller: me,
				ManaAbilities: []protocol.ManaAbilityView{{TapCost: true, Produced: "{W3|U3|B3|R3|G3}"}}},
		}},
	}
	// pool 1 + Forest 1 + Sol Ring 2 + Signet 1 + Mystery Land 1 + Gilded Lotus 3.
	if got, want := manaLeft(&v, me), 9; got != want {
		t.Errorf("manaLeft = %d, want %d", got, want)
	}
}

// turnFeed builds windows for one seat at a two-seat table.
type turnFeed struct {
	me, them uuid.UUID
	rock     uuid.UUID
	spell    uuid.UUID
	other    uuid.UUID
}

func newTurnFeed() turnFeed {
	return turnFeed{me: uuid.New(), them: uuid.New(), rock: uuid.New(), spell: uuid.New(), other: uuid.New()}
}

// view is the seat's view in turn seq on step, active seat active
// (0 is the seat, 1 the opponent), with lands untapped Forests and the
// stack holding a spell when stacked.
func (f turnFeed) view(seq, active int, step string, lands int, stacked bool) protocol.GameView {
	me := f.me.String()
	v := protocol.GameView{
		Seats: []protocol.PlayerView{{ID: me}, {ID: f.them.String()}},
		Turn:  protocol.TurnView{Seq: seq, ActiveSeat: active, Step: step, PhaseID: phaseOf(step)},
	}
	for i := 0; i < lands; i++ {
		v.Battlefield.Cards = append(v.Battlefield.Cards, forest(uuid.NewString(), me, false))
	}
	if stacked {
		v.Stack.Cards = []protocol.CardView{{InstanceID: f.rock.String(), Name: "Arcane Signet"}}
	}
	return v
}

func phaseOf(step string) int {
	if step == "postcombat_main" {
		return 4
	}
	return 2
}

// moves are a pass, then a cast per source given.
func (f turnFeed) moves(sources ...uuid.UUID) []legal.Move {
	out := []legal.Move{{Kind: legal.KindPass, Label: "Pass"}}
	for _, s := range sources {
		out = append(out, legal.Move{Kind: legal.KindCast, Source: s, Label: "Cast " + s.String()[:4]})
	}
	return out
}

func (f turnFeed) event(seat uuid.UUID, v protocol.GameView, moves []legal.Move, index int, applied bool) aiseat.DecisionEvent {
	return aiseat.DecisionEvent{Seat: seat, Input: aiseat.Input{View: v, Seat: seat, Moves: moves}, Index: index, Applied: applied}
}

func TestStrandedManaIsReadAtTheLastMainPhasePass(t *testing.T) {
	f := newTurnFeed()
	w := newTurnManaWatch()
	pass := func(seq, active int, step string, lands int, stacked bool, casts ...uuid.UUID) {
		w.Observe(f.event(f.me, f.view(seq, active, step, lands, stacked), f.moves(casts...), 0, true))
	}

	// Turn 1: five open in the first main phase, one at the last pass.
	// The second pass is the one the turn is read at: not stranded.
	pass(1, 0, "precombat_main", 5, false, f.spell)
	pass(1, 0, "postcombat_main", 1, false, f.spell)
	// Turn 2: three open and a cast on offer at the last pass: stranded.
	pass(2, 0, "postcombat_main", 3, false, f.spell)
	// A pass with a spell on the stack is not the end of a phase.
	pass(2, 0, "postcombat_main", 7, true, f.spell)
	// Turn 3: three open and nothing castable: idle, not stranded.
	pass(3, 0, "postcombat_main", 3, false)
	// The opponent's turn, and a step that is not a main phase, are
	// not the seat's own main phase.
	pass(4, 1, "postcombat_main", 6, false, f.spell)
	pass(5, 0, "end", 6, false, f.spell)
	// A cast is not a pass; a rejected pass is no pass.
	w.Observe(f.event(f.me, f.view(6, 0, "postcombat_main", 6, false), f.moves(f.spell), 1, true))
	w.Observe(f.event(f.me, f.view(6, 0, "postcombat_main", 6, false), f.moves(f.spell), 0, false))

	got := w.forSeat(f.me)
	want := TurnMana{Turns: 3, Stranded: 1, Idle: 2, Unspent: 1 + 3 + 3}
	if got != want {
		t.Errorf("forSeat = %+v, want %+v", got, want)
	}
	if got.StrandedShare() != 1.0/3 || got.MeanUnspent() != 7.0/3 {
		t.Errorf("share %v, mean %v", got.StrandedShare(), got.MeanUnspent())
	}
	if (w.forSeat(f.them) != TurnMana{}) {
		t.Errorf("the opponent, who was never observed, has numbers: %+v", w.forSeat(f.them))
	}
}

func TestPlanMissesCountOnlyTheModelsOwnMisses(t *testing.T) {
	f := newTurnFeed()
	w := newTurnManaWatch()
	main := func(step string, stacked bool, sources ...uuid.UUID) (protocol.GameView, []legal.Move) {
		return f.view(1, 0, step, 4, stacked), f.moves(sources...)
	}
	plan := []aiseat.PlanMember{{Index: 1, Label: "Cast the rock"}, {Index: 2, Label: "Cast the spell"}}
	// planFirst is a window whose plan is rock then spell, in which the
	// seat casts the rock (move 1).
	planFirst := func(step string) {
		v, m := main(step, false, f.rock, f.spell)
		ev := f.event(f.me, v, m, 1, true)
		ev.Trace.Plan = plan
		w.Observe(ev)
	}
	// next is the seat's next window on step, offering the given casts.
	next := func(step string, sources ...uuid.UUID) {
		v, m := main(step, false, sources...)
		w.Observe(f.event(f.me, v, m, 0, true))
	}

	// 1. The rock resolves and the spell is offered: checked, no miss.
	//    The window with the rock on the stack is not where it is checked.
	planFirst("precombat_main")
	v, m := main("precombat_main", true, f.spell)
	w.Observe(f.event(f.me, v, m, 0, true))
	next("precombat_main", f.spell)

	// 2. The spell is not offered after the rock: a miss.
	planFirst("precombat_main")
	next("precombat_main", f.other)

	// 3. The opponent acts in between (not a pass): not checked.
	planFirst("precombat_main")
	ov := f.view(1, 0, "precombat_main", 0, true)
	w.Observe(f.event(f.them, ov, f.moves(f.other), 1, true))
	next("precombat_main", f.other)

	// 4. The opponent only passes in between: checked, and a miss.
	planFirst("precombat_main")
	w.Observe(f.event(f.them, ov, f.moves(f.other), 0, true))
	next("precombat_main", f.other)

	// 5. The next window is in a later phase: the plan is gone, not
	//    missed.
	planFirst("precombat_main")
	next("postcombat_main", f.other)

	// 6. A plan whose first move was not the one made sets nothing up.
	{
		v, m := main("postcombat_main", false, f.rock, f.spell)
		ev := f.event(f.me, v, m, 2, true)
		ev.Trace.Plan = plan
		w.Observe(ev)
		next("postcombat_main", f.other)
	}

	// 7. A held member is skipped: the one after it is expected.
	{
		v, m := main("postcombat_main", false, f.rock, f.spell, f.other)
		ev := f.event(f.me, v, m, 1, true)
		ev.Trace.Plan = []aiseat.PlanMember{{Index: 1}, {Index: 2, Held: true}, {Index: 3}}
		w.Observe(ev)
		next("postcombat_main", f.other)
	}

	// 8. A plan of one is not a plan.
	{
		v, m := main("postcombat_main", false, f.rock)
		ev := f.event(f.me, v, m, 1, true)
		ev.Trace.Plan = plan[:1]
		w.Observe(ev)
	}

	// 9. The rock's resolution asks the seat something (a search, a
	//    "may"): that prompt offers no pass and no cast, and is not where
	//    the plan is checked. The window after it is: checked, no miss.
	planFirst("precombat_main")
	{
		v, _ := main("precombat_main", false)
		w.Observe(f.event(f.me, v, []legal.Move{{Kind: legal.KindChoice, Label: "take a Forest"}}, 0, true))
	}
	next("precombat_main", f.spell)

	// 10. With the rock still on the stack, the seat casts the spell
	//     itself, at instant speed: the plan's next cast was payable and
	//     made. Checked, no miss, though the next main-phase window no
	//     longer offers the spell.
	planFirst("precombat_main")
	{
		v, m := main("precombat_main", true, f.spell)
		w.Observe(f.event(f.me, v, m, 1, true))
	}
	next("precombat_main", f.other)

	got := w.forSeat(f.me)
	if got.Planned != 9 || got.Checked != 6 || got.Misses != 2 {
		t.Errorf("planned %d, checked %d, misses %d; want 9, 6, 2", got.Planned, got.Checked, got.Misses)
	}
	if got.MissShare() != 2.0/9 {
		t.Errorf("MissShare = %v, want 2/9", got.MissShare())
	}
}

func TestTurnManaSectionIsInTheReport(t *testing.T) {
	tm := TurnMana{Turns: 10, Stranded: 4, Idle: 5, Unspent: 25}
	s := Summary{
		PerPolicy:     map[string]*PolicyTotals{"heuristic": {Policy: "heuristic", Games: 1, TurnMana: tm}},
		PerContestant: []*PolicyTotals{{Policy: "heuristic", Deck: "simic-ramp", Games: 1, TurnMana: tm}},
	}
	md := s.Markdown()
	for _, want := range []string{
		"### Turn mana (ADR 0136 §8)",
		"| heuristic | all | 10 | 4 | 40.0% | 5 | 2.50 | 0 | 0 | 0 | 0.0% |",
		"| heuristic | simic-ramp | 10 | 4 | 40.0% |",
	} {
		if !strings.Contains(md, want) {
			t.Errorf("Markdown() is missing %q:\n%s", want, md)
		}
	}
}
