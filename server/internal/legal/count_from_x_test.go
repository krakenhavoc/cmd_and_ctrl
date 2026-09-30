package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// count_from_x_test.go — #619, the enumerator's half of an X-defined
// target count (CR 601.2c, TargetSpec.CountFromX).
//
// The soak found this one as a REJECTED move, which is the enumerator
// breaking its own contract rather than playing badly:
//
//	WARN cast_spell rejected: X-defined target count mismatch
//	     card_name="Crackle with Power" x_value=0 targets_received=1
//
// The X and the targets were picked in two different places — the
// affordable X off the cost, the target list off the clause's printed
// count — and for a clause whose count IS X they have to be one
// decision.

const (
	oracleCrackleWithPower = "273f5483-b67e-4dd6-bba8-c0a047fa34d7"
	oracleWaterbenders     = "285046f6-b3c4-4eb7-8712-9dffebabc762"
)

// castXAndTargets decodes the pair this file is about.
func castXAndTargets(t *testing.T, m legal.Move) (x, targets int) {
	t.Helper()
	x, targets, _ = castXTargetsModes(t, m)
	return x, targets
}

// castXTargetsModes also returns the chosen modes, which decide WHICH
// clause a modal card's cast is answering.
func castXTargetsModes(t *testing.T, m legal.Move) (x, targets int, modes []int) {
	t.Helper()
	var p struct {
		Targets []json.RawMessage `json:"targets"`
		Modes   []int             `json:"modes"`
		XValue  int               `json:"x_value"`
	}
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatalf("params %s: %v", string(m.Params), err)
	}
	return p.XValue, len(p.Targets), p.Modes
}

// targetRefWire mirrors the unexported wire shape legal.wireTargets
// writes (legal.go's targetWire) — kind/id/slot/mode, the same field
// names the client and the engine both read. Slot and Mode are
// omitted at zero, which json.Unmarshal already treats as the default
// for an absent field, so a single-clause non-modal move's targets
// decode with Mode == Slot == 0 exactly as they should.
type targetRefWire struct {
	Kind string `json:"kind"`
	ID   string `json:"id,omitempty"`
	Slot int    `json:"slot,omitempty"`
	Mode int    `json:"mode,omitempty"`
}

// castXTargetsPerStep is castXTargetsModes plus each target's own
// (Mode, Slot) — the per-step counting a modal card with more than
// one announced clause needs (a "choose two" cycle whose every bullet
// targets announces one clause per chosen bullet, so len(Targets) is
// the WHOLE move's count, not any one bullet's).
func castXTargetsPerStep(t *testing.T, m legal.Move) (x int, refs []targetRefWire, modes []int) {
	t.Helper()
	var p struct {
		Targets []targetRefWire `json:"targets"`
		Modes   []int           `json:"modes"`
		XValue  int             `json:"x_value"`
	}
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatalf("params %s: %v", string(m.Params), err)
	}
	return p.XValue, p.Targets, p.Modes
}

// Crackle with Power is {X}{X}{X}{R}{R}: X=1 costs five mana, X=2
// costs eight. Off five Mountains the only cast is one target at X=1,
// and off eight both arities are offered — each announcing exactly as
// many as it targets. X=0 with a target, the move the engine refused,
// is offered at neither.
func TestCrackleWithPowerTiesXToItsTargetCount(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	crackle := handCard(active, game.Card{
		Name: "Crackle with Power", TypeLine: "Sorcery",
		ManaCost: "{X}{X}{X}{R}{R}", OracleID: oracleCrackleWithPower,
	})
	for i := 0; i < 5; i++ {
		battlefieldCard(g, active, basic("Mountain", "Mountain"))
	}
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	casts := castMovesFor(moves, crackle)
	if len(casts) == 0 {
		t.Fatalf("Crackle not offered off five Mountains: %v", labels(moves))
	}
	for _, m := range casts {
		x, n := castXAndTargets(t, m)
		if n != 1 || x != 1 {
			t.Errorf("five mana buys X=1 and one target; %q offered X=%d with %d targets", m.Label, x, n)
		}
	}
	dispatchAll(t, g, active.ID, moves)

	for i := 0; i < 3; i++ {
		battlefieldCard(g, active, basic("Mountain", "Mountain"))
	}
	moves = legal.EnumerateFor(g, active.ID)
	casts = castMovesFor(moves, crackle)
	arities := map[int]bool{}
	for _, m := range casts {
		x, n := castXAndTargets(t, m)
		if x != n {
			t.Errorf("%q announced X=%d for %d targets — the count IS X (CR 601.2c)", m.Label, x, n)
		}
		if n == 0 {
			t.Errorf("%q targets nothing: at X=0 Crackle deals five times nothing (#810)", m.Label)
		}
		arities[n] = true
	}
	if !arities[1] || !arities[2] {
		t.Errorf("eight mana pays for X=1 and X=2; arities offered = %v", arities)
	}
	dispatchAll(t, g, active.ID, moves)
}

// Waterbender's Restoration's X is announced by its waterbend cost,
// not by its mana cost ({U}{U}, no {X} slot). The enumerator cannot
// price that, so the only announcement it could make is X=0 — which
// buys no targets and a spell that does nothing — and any larger one
// would be a cast the engine refuses for an unpaid cost. So it offers
// none, where before #619 it offered one target at X=0 and had it
// rejected.
func TestAnXFromANonManaCostIsNotEnumerated(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	blink := handCard(active, game.Card{
		Name: "Waterbender's Restoration", TypeLine: "Instant — Lesson",
		ManaCost: "{U}{U}", OracleID: oracleWaterbenders,
	})
	battlefieldCard(g, active, creature("Blinkee", "{1}{G}", 2, 2))
	for i := 0; i < 6; i++ {
		battlefieldCard(g, active, basic("Island", "Island"))
	}
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	if casts := castMovesFor(moves, blink); len(casts) != 0 {
		t.Errorf("offered %d casts of a spell whose X the enumerator cannot pay for: %v",
			len(casts), labels(casts))
	}
	dispatchAll(t, g, active.ID, moves)
}

// The catalog-wide statement, which is what makes this a rule rather
// than two cards: for EVERY registered clause whose count is X, every
// cast the enumerator offers announces exactly as many targets as it
// picks, and the engine accepts all of them.
//
// Each card is probed at a stand-in printed cost of "{X}{R}" rather
// than its real one, because a Spec carries an oracle ID and a name,
// not a mana cost (that comes from the Scryfall dump, which these
// tests do not load). The rule under test is about the SHAPE of the
// announcement, so a cost the seat can obviously afford is the probe
// that exercises the most arities. A spec with its own tap cost is
// skipped and named: its X is that cost's, not the mana cost's, and
// TestAnXFromANonManaCostIsNotEnumerated is where that case lives.
func TestEveryXDefinedTargetCountIsSound(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	clearHand(active)
	// A board with one of everything an X-defined clause might want.
	battlefieldCard(g, active, creature("Mine", "{1}{G}", 2, 2))
	battlefieldCard(g, opp, creature("Theirs", "{1}{G}", 2, 2))
	battlefieldCard(g, opp, game.Card{Name: "Their Rock", TypeLine: "Artifact"})
	battlefieldCard(g, opp, game.Card{Name: "Their Aura", TypeLine: "Enchantment"})
	for i := 0; i < 12; i++ {
		battlefieldCard(g, active, basic("Mountain", "Mountain"))
	}
	// Kozilek's Command's fourth bullet is CountFromX over "cards
	// from graveyards" with no other predicate (Min == Max == X):
	// the enumerator only ever tries the single MAX AFFORDABLE X
	// (affordableXFrom's doc — X never ranges), which for the
	// {X}{R} stand-in cost below and 12 Mountains is 11. Seed
	// exactly that many graveyard cards, so the clause is fillable
	// with exactly one combination (all of them) rather than an
	// unfillable "pick 11 of 1" or a combinatorial "pick 11 of 20".
	for i := 0; i < 11; i++ {
		graveyardCard(opp, creature("Their Dead Creature", "{1}{G}", 2, 2))
	}
	advanceTo(t, g, game.StepPrecombatMain)

	probed := 0
	for _, spec := range effects.All() {
		if !specCountsTargetsFromX(spec) {
			continue
		}
		if spec.TapCost != nil {
			t.Logf("skipping %s: its X is a tap cost's, not the mana cost's", spec.Name)
			continue
		}
		probed++
		id := handCard(active, game.Card{
			Name: spec.Name, TypeLine: "Sorcery", ManaCost: "{X}{R}", OracleID: spec.OracleID,
		})
		// Kozilek's Command: "choose two" of four bullets, two of
		// them plain "target player" (16 target-pair combinations on
		// their own). The default per-source budget (12) is spent
		// entirely on the FIRST mode-set the recursive builder tries
		// (modes 0 and 1), so the graveyard bullet's own X-defined
		// clause never gets a combination to appear in. Raise the
		// budget for this sweep alone — it costs nothing else here,
		// the board is small and fixed, and it is what lets every
		// mode-set past the first get a turn.
		moves := legal.EnumerateForWithOptions(g, active.ID, legal.Options{MaxExpansionPerSource: 200})
		tied := 0
		for _, m := range castMovesFor(moves, id) {
			x, refs, modes := castXTargetsPerStep(t, m)
			// Ask the engine which clauses this announcement is
			// actually answering, rather than re-deriving it here.
			// Only some of a modal card's are X-counted: Heliod's
			// Intervention destroys X artifacts under one mode and
			// gives a player twice X life under the other, where X is
			// free and the one target is a player.
			steps := game.AnnouncedClauses(
				game.TargetSpecFor(spec.OracleID), game.ModeSpecFor(spec.OracleID), modes)
			// Per-step counting (#764): a "choose two" cycle whose
			// every bullet targets (Kozilek's Command) announces one
			// clause per CHOSEN bullet, so the move's whole target
			// list is the wrong count for any one bullet — only the
			// refs whose (Mode, Slot) match the X-counted step's own
			// answer it, exactly as ModeTargets narrows to one
			// occurrence at resolution.
			for _, st := range steps {
				if !st.Clause.CountFromX {
					continue
				}
				n := 0
				for _, r := range refs {
					if r.Mode == st.Mode && r.Slot == st.Slot {
						n++
					}
				}
				tied++
				if x != n {
					t.Errorf("%s: %q announced X=%d for %d targets on its X-defined step",
						spec.Name, m.Label, x, n)
				}
				if n == 0 {
					t.Errorf("%s: %q targets nothing on its X-defined step — an X-defined clause "+
						"at X=0 does nothing (#810)", spec.Name, m.Label)
				}
			}
		}
		if tied == 0 {
			t.Errorf("%s: no cast of its X-defined clause was offered on a board that can pay for one", spec.Name)
		}
		dispatchAll(t, g, active.ID, moves)
		removeFromHand(active, id)
	}
	if probed == 0 {
		t.Fatal("no CountFromX clause in the catalog — the sweep is reading the wrong field")
	}
	t.Logf("probed %d X-defined target clauses", probed)
}

// specCountsTargetsFromX reports whether a Spec's card-level clause or
// any of its modes ties the target count to X.
func specCountsTargetsFromX(spec effects.Spec) bool {
	if spec.Targets != nil && spec.Targets.CountFromX {
		return true
	}
	if spec.Modes == nil {
		return false
	}
	for _, o := range spec.Modes.Options {
		if o.Targets != nil && o.Targets.CountFromX {
			return true
		}
	}
	return false
}

func removeFromHand(p *game.Player, id uuid.UUID) {
	for i, c := range p.Hand.Cards {
		if c.InstanceID == id {
			p.Hand.Cards = append(p.Hand.Cards[:i], p.Hand.Cards[i+1:]...)
			return
		}
	}
}
