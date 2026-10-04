package legal_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// cuts_test.go — ADR 0122 §6.2. Every count cap in this package now
// says what it cut, and a request naming one card or one prompt lifts
// the caps to legal.ExpandCeiling. Each test below takes one cap to a
// board where it bites, asks for the report, and then asks again for
// that card or prompt alone.

// cutsOf picks the cuts filed for one cap.
func cutsOf(cuts []legal.Cut, c legal.Cap) []legal.Cut {
	var out []legal.Cut
	for _, k := range cuts {
		if k.Cap == c {
			out = append(out, k)
		}
	}
	return out
}

// sameMoves fails unless the report's moves are exactly the moves the
// bot's path enumerates: the report is extra, never different.
func sameMoves(t *testing.T, g *game.Game, seat uuid.UUID, rep legal.Report) {
	t.Helper()
	plain, err := json.Marshal(legal.EnumerateFor(g, seat))
	if err != nil {
		t.Fatal(err)
	}
	reported, err := json.Marshal(rep.Moves)
	if err != nil {
		t.Fatal(err)
	}
	if string(plain) != string(reported) {
		t.Fatalf("EnumerateReport's moves differ from EnumerateFor's:\nfor:    %s\nreport: %s", plain, reported)
	}
}

func movesOfSource(moves []legal.Move, src uuid.UUID) []legal.Move {
	var out []legal.Move
	for _, m := range moves {
		if m.Source == src {
			out = append(out, m)
		}
	}
	return out
}

// Lightning Bolt with 24 legal targets: twelve offered, the other
// twelve reported exactly, and all 24 when the card is asked for alone.
func TestCutReportPerSourceIsExactAndASourceRequestLiftsIt(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	clearHand(active)
	bolt := handCard(active, game.Card{Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: oracleLightningBolt})
	battlefieldCard(g, active, basic("Mountain", "Mountain"))
	for i := 0; i < 20; i++ {
		battlefieldCard(g, opp, creature(fmt.Sprintf("Bear %d", i), "{1}{G}", 2, 2))
	}
	advanceTo(t, g, game.StepPrecombatMain)

	rep := legal.EnumerateReport(g, active.ID, legal.Options{})
	sameMoves(t, g, active.ID, rep)
	if n := len(movesOfSource(rep.Moves, bolt)); n != 12 {
		t.Fatalf("want the default 12 Bolt moves, got %d", n)
	}
	cuts := cutsOf(rep.Cuts, legal.CapPerSource)
	if len(cuts) != 1 {
		t.Fatalf("want one per_source cut, got %+v", rep.Cuts)
	}
	// Four players and twenty Bears.
	if c := cuts[0]; c.Source != bolt || c.Omitted != 12 || c.AtLeast {
		t.Errorf("cut %+v, want Bolt, 12 omitted, exact", c)
	}

	one := legal.EnumerateReport(g, active.ID, legal.Options{Source: bolt})
	if len(one.Moves) != 24 {
		t.Fatalf("a request for the Bolt alone: want all 24 targets, got %d: %v", len(one.Moves), labels(one.Moves))
	}
	for _, m := range one.Moves {
		if m.Source != bolt {
			t.Errorf("a request for one card returned another's move: %q", m.Label)
		}
	}
	if len(one.Cuts) != 0 {
		t.Errorf("nothing is cut under the ceiling, got %+v", one.Cuts)
	}
	dispatchAll(t, g, active.ID, one.Moves)
}

// The cleanup discard has no pending-choice ID; it is named
// cleanup_discard, in the report and in the request.
func TestCutReportNamesTheCleanupDiscard(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	for i := 0; i < 9; i++ {
		handCard(active, creature(fmt.Sprintf("Card %d", i), "{1}", 1, 1))
	}
	advanceTo(t, g, game.StepEnd)
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatal(err)
	}
	rep := legal.EnumerateReport(g, active.ID, legal.Options{})
	sameMoves(t, g, active.ID, rep)
	// C(9,2) = 36, twelve offered.
	cuts := cutsOf(rep.Cuts, legal.CapPerSource)
	if len(cuts) != 1 || cuts[0].Choice != legal.CleanupDiscardChoice || cuts[0].Omitted != 24 || cuts[0].AtLeast {
		t.Fatalf("want one exact cut of 24 on cleanup_discard, got %+v", rep.Cuts)
	}
	one := legal.EnumerateReport(g, active.ID, legal.Options{Choice: legal.CleanupDiscardChoice})
	if len(one.Moves) != 36 || len(one.Cuts) != 0 {
		t.Fatalf("want all 36 discards and no cut, got %d moves, cuts %+v", len(one.Moves), one.Cuts)
	}
}

// An X past MaxX is a max_x cut, and the move states its X as an open
// range the server accepts.
func TestCutReportMaxXAndTheOpenX(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	sooth := battlefieldCard(g, active, game.Card{
		Name: "Soothsaying", TypeLine: "Enchantment", ManaCost: "{U}", OracleID: oracleSoothsaying,
	})
	advanceTo(t, g, game.StepPrecombatMain)
	mana(g, active, 25)

	rep := legal.EnumerateReport(g, active.ID, legal.Options{})
	sameMoves(t, g, active.ID, rep)
	var xMove *legal.Move
	for _, m := range activationsOf(rep.Moves, sooth) {
		if m.Value != nil {
			m := m
			xMove = &m
		}
	}
	if xMove == nil {
		t.Fatalf("Soothsaying's {X} activation carries no open X: %v", labels(rep.Moves))
	}
	if v := xMove.Value; v.Kind != legal.ValueX || v.Min == nil || v.Max == nil || *v.Min != 1 || *v.Max != 20 {
		t.Errorf("want x from 1 to 20 (the cap), got %+v", v)
	}
	if x := xValueOf(t, *xMove); x != 20 {
		t.Errorf("the move announces X=%d, want the largest offered, 20", x)
	}
	cuts := cutsOf(rep.Cuts, legal.CapMaxX)
	if len(cuts) != 1 || cuts[0].Source != sooth || !cuts[0].AtLeast {
		t.Fatalf("want one at-least max_x cut on Soothsaying, got %+v", rep.Cuts)
	}

	one := legal.EnumerateReport(g, active.ID, legal.Options{Source: sooth})
	if len(cutsOf(one.Cuts, legal.CapMaxX)) != 0 {
		t.Errorf("the ceiling is above 25 mana, yet max_x is still cut: %+v", one.Cuts)
	}
	found := false
	for _, m := range one.Moves {
		if m.Value != nil && m.Value.Kind == legal.ValueX {
			found = true
			if *m.Value.Max != 25 {
				t.Errorf("expanded: want X up to all 25 mana, got %d", *m.Value.Max)
			}
		}
	}
	if !found {
		t.Fatalf("expanded: no open X: %v", labels(one.Moves))
	}
	dispatchAll(t, g, active.ID, one.Moves)
}

// A card-name prompt: five suggestions, the rest of the ranked names
// reported, every answer marked as an open card name.
func TestCutReportCardNamesAndTheOpenName(t *testing.T) {
	g := newTable(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	for i := 0; i < 7; i++ {
		battlefieldCard(g, opp, game.Card{
			Name: fmt.Sprintf("Bombardment %d", i), TypeLine: "Enchantment",
			ManaCost: "{1}{R}", OracleID: oracleGoblinBombardment,
		})
	}
	var choice uuid.UUID
	g.WithWriteLock(func() {
		choice = g.QueueCardNameChoiceForEffect(me.ID, uuid.New(), "Pithing Needle — choose a card name")
	})

	rep := legal.EnumerateReport(g, me.ID, legal.Options{})
	sameMoves(t, g, me.ID, rep)
	if len(rep.Moves) != 5 {
		t.Fatalf("want the five suggestions, got %d: %v", len(rep.Moves), labels(rep.Moves))
	}
	for _, m := range rep.Moves {
		if m.Value == nil || m.Value.Kind != legal.ValueCardName || m.Value.Min != nil {
			t.Errorf("%q does not say its name is open: %+v", m.Label, m.Value)
		}
	}
	cuts := cutsOf(rep.Cuts, legal.CapCardNames)
	if len(cuts) != 1 || cuts[0].Choice != choice.String() || cuts[0].Omitted != 2 || cuts[0].AtLeast {
		t.Fatalf("want an exact card_names cut of 2 on the prompt, got %+v", rep.Cuts)
	}
	one := legal.EnumerateReport(g, me.ID, legal.Options{Choice: choice.String()})
	if len(one.Moves) != 7 || len(one.Cuts) != 0 {
		t.Fatalf("the prompt alone: want all seven names and no cut, got %d, %+v", len(one.Moves), one.Cuts)
	}
	dispatchAll(t, g, me.ID, one.Moves)
}

// A creature-type prompt names five tribes off the board; the rest of
// the vocabulary is the cut, and the prompt alone lists all of it.
func TestCutReportCreatureTypesListsTheVocabulary(t *testing.T) {
	g := newTable(t)
	me := g.Seats[0]
	for _, tp := range []string{"Angel", "Bird", "Cat", "Dragon", "Elf", "Faerie", "Goblin"} {
		battlefieldCard(g, me, typedCreature("A "+tp, tp))
	}
	queueCreatureTypeChoice(g, me.ID)

	rep := legal.EnumerateReport(g, me.ID, legal.Options{})
	sameMoves(t, g, me.ID, rep)
	cuts := cutsOf(rep.Cuts, legal.CapCreatureTypes)
	if len(cuts) != 1 || cuts[0].Omitted != len(game.AllCreatureTypes)-5 || cuts[0].AtLeast {
		t.Fatalf("want an exact creature_types cut of %d, got %+v", len(game.AllCreatureTypes)-5, rep.Cuts)
	}
	choice := cuts[0].Choice
	one := legal.EnumerateReport(g, me.ID, legal.Options{Choice: choice})
	if len(one.Moves) != len(game.AllCreatureTypes) || len(one.Cuts) != 0 {
		t.Fatalf("the prompt alone: want the whole vocabulary (%d) and no cut, got %d, %+v",
			len(game.AllCreatureTypes), len(one.Moves), one.Cuts)
	}
	// The board's tribes still come first.
	if !strings.HasSuffix(one.Moves[0].Label, ": Angel") {
		t.Errorf("first answer %q, want the board's tribes first", one.Moves[0].Label)
	}
	dispatchAll(t, g, me.ID, one.Moves[:8])
}

// Vicious Betrayal with six creatures: counts 0 to 3 offered, 4 to 6
// reported, all of them when the card is asked for alone.
func TestCutReportVariableCounts(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 5, "Swamp")
	for i := 0; i < 6; i++ {
		battlefieldCard(g, seat, creature("Bear", "{2}", 2, 2))
	}
	vb := handCard(seat, game.Card{Name: "Vicious Betrayal", TypeLine: "Sorcery", ManaCost: "{3}{B}{B}", OracleID: oracleViciousBetrayal})

	rep := legal.EnumerateReport(g, seat.ID, legal.Options{})
	sameMoves(t, g, seat.ID, rep)
	cuts := cutsOf(rep.Cuts, legal.CapVariableCounts)
	if len(cuts) != 1 || cuts[0].Source != vb || cuts[0].Omitted != 3 || cuts[0].AtLeast {
		t.Fatalf("want an exact variable_counts cut of 3 on Vicious Betrayal, got %+v", rep.Cuts)
	}
	one := legal.EnumerateReport(g, seat.ID, legal.Options{Source: vb})
	got := sacCounts(varSacMovesOf(t, one.Moves, vb))
	for n := 0; n <= 6; n++ {
		if !got[n] {
			t.Errorf("expanded: count %d not offered (offered %v)", n, got)
		}
	}
	dispatchAll(t, g, seat.ID, one.Moves)
}

// Multikicker past three payments is a repeats cut.
func TestCutReportRepeats(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 12, "Forest")
	wolfbriar := handCard(seat, game.Card{
		Name: "Wolfbriar Elemental", TypeLine: "Creature — Elemental", ManaCost: "{2}{G}{G}",
		Power: 4, Toughness: 4, OracleID: "2f8872fe-84dc-4cda-a253-e7503a5c96a3",
	})
	rep := legal.EnumerateReport(g, seat.ID, legal.Options{})
	sameMoves(t, g, seat.ID, rep)
	cuts := cutsOf(rep.Cuts, legal.CapRepeats)
	if len(cuts) != 1 || cuts[0].Source != wolfbriar || cuts[0].Omitted < 1 || cuts[0].AtLeast {
		t.Fatalf("want an exact repeats cut on Wolfbriar, got %+v", rep.Cuts)
	}
	// Twelve Forests pay {2}{G}{G} and eight kicks.
	one := legal.EnumerateReport(g, seat.ID, legal.Options{Source: wolfbriar})
	if !hasLabel(one.Moves, "Cast Wolfbriar Elemental (Multikicker {G} ×8)") {
		t.Errorf("expanded: the eight-kick cast is missing: %v", labels(one.Moves))
	}
	if hasLabel(one.Moves, "Cast Wolfbriar Elemental (Multikicker {G} ×9)") {
		t.Errorf("expanded: a nine-kick cast twelve Forests cannot pay: %v", labels(one.Moves))
	}
	dispatchAll(t, g, seat.ID, one.Moves)
}

// Escape over six fuel cards: three of the fifteen payments offered,
// twelve reported.
func TestCutReportCostPayments(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 7, "Forest")
	typhon := graveyardCard(seat, game.Card{
		Name: "Voracious Typhon", TypeLine: "Creature — Hydra", ManaCost: "{4}{G}",
		Power: 4, Toughness: 4, OracleID: oracleVoraciousTyphon,
	})
	for i := 0; i < 6; i++ {
		graveyardCard(seat, game.Card{Name: fmt.Sprintf("Fuel %d", i), TypeLine: "Instant", ManaCost: "{1}"})
	}
	rep := legal.EnumerateReport(g, seat.ID, legal.Options{})
	sameMoves(t, g, seat.ID, rep)
	cuts := cutsOf(rep.Cuts, legal.CapCostPayments)
	if len(cuts) != 1 || cuts[0].Source != typhon || cuts[0].Omitted != 12 || cuts[0].AtLeast {
		t.Fatalf("want an exact cost_payments cut of 12 on the Typhon, got %+v", rep.Cuts)
	}
	one := legal.EnumerateReport(g, seat.ID, legal.Options{Source: typhon})
	escapes := 0
	for _, p := range castPayloadsOf(t, one.Moves, typhon) {
		if p.AlternativeCost == "escape" {
			escapes++
		}
	}
	if escapes != 15 {
		t.Errorf("expanded: want all 15 escape payments, got %d", escapes)
	}
}

// A menace attacker facing seven blockers has 21 pairs; twelve are
// offered, and the attacker alone lists every one.
func TestCutReportBlockGroups(t *testing.T) {
	g, attacker, _, defender := menaceTable(t, 7)
	rep := legal.EnumerateReport(g, defender.ID, legal.Options{})
	sameMoves(t, g, defender.ID, rep)
	cuts := cutsOf(rep.Cuts, legal.CapPerSource)
	if len(cuts) != 1 || cuts[0].Source != attacker || cuts[0].Omitted < 1 || !cuts[0].AtLeast {
		t.Fatalf("want an at-least per_source cut on the attacker, got %+v", rep.Cuts)
	}
	one := legal.EnumerateReport(g, defender.ID, legal.Options{Source: attacker})
	groups := 0
	for _, m := range one.Moves {
		if m.Type == legal.TypeDeclareBlockers {
			groups++
		}
	}
	if groups != 21 {
		t.Errorf("expanded: want all 21 pairs, got %d: %v", groups, labels(one.Moves))
	}
	if len(one.Cuts) != 0 {
		t.Errorf("expanded: nothing left to cut, got %+v", one.Cuts)
	}
}

// A request names cards and prompts only. The pass, which belongs to
// neither, is not in any of them; an unknown card is an empty answer.
func TestASourceRequestHoldsNothingElse(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	bolt := handCard(active, game.Card{Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: oracleLightningBolt})
	battlefieldCard(g, active, basic("Mountain", "Mountain"))
	advanceTo(t, g, game.StepPrecombatMain)

	for _, m := range legal.EnumerateReport(g, active.ID, legal.Options{Source: bolt}).Moves {
		if m.Source != bolt {
			t.Errorf("request for the Bolt returned %q", m.Label)
		}
	}
	if rep := legal.EnumerateReport(g, active.ID, legal.Options{Source: uuid.New()}); len(rep.Moves) != 0 || len(rep.Cuts) != 0 {
		t.Errorf("an unknown card answered %v / %+v", labels(rep.Moves), rep.Cuts)
	}
}
