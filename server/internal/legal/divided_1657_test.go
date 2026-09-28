package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// divided_1657_test.go — #1657: the enumerator sizes a divided amount
// read off the board, or off the offer it claims, exactly as the gate
// will, so every offered division is one the engine accepts (#544).

const (
	oracleUreniSongUnending = "1c995c80-3301-409b-822b-9297aa260823"
	oracleAvacynsJudgment   = "f3ae58ed-8ef7-4e0a-945f-1f622157236b"
	oracleLathiel           = "d3c56fc4-3611-41b1-952e-4c5311b1510b"
)

// pickTargetAnswers decodes every answer the enumerator offers to the
// open pick_target prompt and checks it against the amount.
func pickTargetAnswers(t *testing.T, moves []legal.Move, total int) (maxTargets int) {
	t.Helper()
	for _, m := range moves {
		var p struct {
			Targets      []map[string]any `json:"targets"`
			Target       map[string]any   `json:"target"`
			Distribution map[string]int   `json:"distribution"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		n := len(p.Targets)
		if p.Target != nil {
			n = 1
		}
		if n > total {
			t.Errorf("%q: %d targets cannot divide %d", m.Label, n, total)
		}
		if n > maxTargets {
			maxTargets = n
		}
		if len(p.Targets) >= 2 {
			checkEvenDivision(t, m.Label, dividedMove{Targets: p.Targets, Distribution: p.Distribution}, total)
		}
	}
	return maxTargets
}

// widestDivided is the most targets any of the moves announces — so a
// test can tell "every offer is legal" from "only the empty offer was
// made", which an unsized division would also produce.
func widestDivided(t *testing.T, moves []legal.Move) int {
	t.Helper()
	w := 0
	for _, m := range moves {
		if n := len(decodeDivided(t, m).Targets); n > w {
			w = n
		}
	}
	return w
}

// Ureni's amount is the lands its controller has, not the mana it was
// paid with: two lands and three enemy creatures offer at most two
// targets, split 1/1, and every answer is accepted.
func TestUreniPickTargetAnswersSizeTheDivisionByLands(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(active)
	ureni := handCard(active, game.Card{
		Name: "Ureni, the Song Unending", TypeLine: "Legendary Creature — Spirit Dragon",
		OracleID: oracleUreniSongUnending, ManaCost: "{5}{G}{U}{R}", Power: 10, Toughness: 10,
	})
	lands(g, active, "Forest", "Forest", 2)
	for _, name := range []string{"A", "B", "C"} {
		battlefieldCard(g, opp, creature(name, "{1}{G}", 2, 2))
	}
	advanceTo(t, g, game.StepPrecombatMain)
	g.WithWriteLock(func() { _ = g.AddManaForEffect(active.ID, uuid.Nil, "{G}{U}{R}{C}{C}{C}{C}{C}") })
	if err := g.CastSpell(active.ID, ureni, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast Ureni: %v", err)
	}
	for i := 0; i < 8 && len(g.PendingChoices) == 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if len(g.PendingChoices) == 0 || g.PendingChoices[0].Kind != game.PendingChoicePickTarget {
		t.Fatal("expected Ureni's pick_target prompt")
	}
	moves := legal.EnumerateFor(g, active.ID)
	if len(moves) == 0 {
		t.Fatal("no answers to Ureni's prompt — the seat is stuck")
	}
	if got := pickTargetAnswers(t, moves, 2); got != 2 {
		t.Errorf("widest answer offered: %d targets, want 2 (two lands)", got)
	}
	dispatchAll(t, g, active.ID, moves)
}

// Lathiel's "up to": the enumerator announces the whole amount (a
// legal "up to" answer) and offers choosing nobody.
func TestLathielPickTargetAnswersAreAccepted(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	battlefieldCard(g, active, game.Card{
		Name: "Lathiel, the Bounteous Dawn", TypeLine: "Legendary Creature — Unicorn",
		OracleID: oracleLathiel, ManaCost: "{2}{G}{W}", Power: 2, Toughness: 2,
	})
	battlefieldCard(g, active, creature("Mine", "{1}{G}", 2, 2))
	battlefieldCard(g, opp, creature("Theirs", "{1}{G}", 2, 2))
	advanceTo(t, g, game.StepPrecombatMain)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, active.ID, 3) })
	advanceTo(t, g, game.StepEnd)
	for i := 0; i < 8 && len(g.PendingChoices) == 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if len(g.PendingChoices) == 0 || g.PendingChoices[0].Kind != game.PendingChoicePickTarget {
		t.Fatal("expected Lathiel's pick_target prompt")
	}
	moves := legal.EnumerateFor(g, active.ID)
	if len(moves) == 0 {
		t.Fatal("no answers to Lathiel's prompt")
	}
	none := false
	for _, m := range moves {
		var p struct {
			Targets []map[string]any `json:"targets"`
			Target  map[string]any   `json:"target"`
		}
		_ = json.Unmarshal(m.Params, &p)
		if p.Target == nil && len(p.Targets) == 0 {
			none = true
		}
	}
	if !none {
		t.Error("choosing no creature is a legal answer and is not offered")
	}
	pickTargetAnswers(t, moves, 3)
	dispatchAll(t, g, active.ID, moves)
}

// oracleLegalLandCannon is a test-only card: no catalogued ACTIVATED
// ability divides a ruled amount yet, so the activation enumerator's
// binding is proved on this.
const oracleLegalLandCannon = "test-1657-legal-land-cannon"

func init() {
	effects.Register(effects.Spec{
		OracleID: oracleLegalLandCannon,
		Name:     "Test Legal Land Cannon",
		Activated: []effects.ActivatedAbility{{
			Label: "{T}: This deals damage equal to the number of lands you control divided as you choose among any number of targets.",
			Cost:  effects.TapCost(),
			Targets: effects.TargetAny().WithCount(0, 0).
				Dividing(effects.DivideBy(effects.DivideLandsYouControl)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DealDividedDamage(effects.NewContext(g, item))
			},
		}},
	})
}

// An activation's ruled amount is sized as the activation gate sizes
// it: two lands, never a three-target activation.
func TestActivationMovesSizeARuledDivision(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	cannon := battlefieldCard(g, active, game.Card{
		Name: "Test Legal Land Cannon", TypeLine: "Artifact", OracleID: oracleLegalLandCannon,
	})
	lands(g, active, "Wastes", "Wastes", 2)
	for _, name := range []string{"A", "B", "C"} {
		battlefieldCard(g, opp, creature(name, "{1}{G}", 2, 2))
	}
	advanceTo(t, g, game.StepPrecombatMain)
	var moves []legal.Move
	for _, m := range legal.EnumerateFor(g, active.ID) {
		if m.Kind == legal.KindActivate && m.Source == cannon {
			moves = append(moves, m)
		}
	}
	if len(moves) == 0 {
		t.Fatal("the cannon's activation is not offered")
	}
	for _, m := range moves {
		d := decodeDivided(t, m)
		if len(d.Targets) > 2 {
			t.Errorf("%q: %d targets cannot divide 2", m.Label, len(d.Targets))
		}
		checkEvenDivision(t, m.Label, d, 2)
	}
	if w := widestDivided(t, moves); w != 2 {
		t.Errorf("widest activation: %d targets, want 2 (two lands)", w)
	}
	dispatchAll(t, g, active.ID, moves)
}

// Avacyn's Judgment: the printed cast divides 2, the madness cast from
// exile divides the X it announces — and every offer is accepted.
func TestAvacynsJudgmentMovesSizeTheDivisionByTheClaimedCost(t *testing.T) {
	t.Run("hand", func(t *testing.T) {
		g := newTable(t)
		active := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		clearHand(active)
		spell := handCard(active, game.Card{
			Name: "Avacyn's Judgment", TypeLine: "Sorcery",
			OracleID: oracleAvacynsJudgment, ManaCost: "{1}{R}",
		})
		lands(g, active, "Mountain", "Mountain", 2)
		battlefieldCard(g, opp, creature("A", "{1}{G}", 2, 2))
		battlefieldCard(g, opp, creature("B", "{1}{G}", 2, 2))
		battlefieldCard(g, opp, creature("C", "{1}{G}", 2, 2))
		advanceTo(t, g, game.StepPrecombatMain)
		moves := castMovesFor(legal.EnumerateFor(g, active.ID), spell)
		if len(moves) == 0 {
			t.Fatal("Avacyn's Judgment not offered")
		}
		for _, m := range moves {
			d := decodeDivided(t, m)
			if len(d.Targets) > 2 {
				t.Errorf("%q: %d targets cannot divide 2", m.Label, len(d.Targets))
			}
			checkEvenDivision(t, m.Label, d, 2)
		}
		if w := widestDivided(t, moves); w != 2 {
			t.Errorf("widest cast: %d targets, want 2", w)
		}
		dispatchAll(t, g, active.ID, moves)
	})

	t.Run("madness", func(t *testing.T) {
		g := newTable(t)
		active := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		clearHand(active)
		card := game.Card{
			InstanceID: uuid.New(), Name: "Avacyn's Judgment", TypeLine: "Sorcery",
			OracleID: oracleAvacynsJudgment, ManaCost: "{1}{R}",
			Owner: active.ID, Controller: active.ID,
		}
		lands(g, active, "Mountain", "Mountain", 4)
		battlefieldCard(g, opp, creature("A", "{1}{G}", 2, 2))
		battlefieldCard(g, opp, creature("B", "{1}{G}", 2, 2))
		advanceTo(t, g, game.StepPrecombatMain)
		g.WithWriteLock(func() {
			g.Exile.PushTop(card)
			g.GrantCastPermissionToCardsForEffect(game.CastPermission{
				Player: active.ID, Zone: game.ZoneExile, AltCostKey: game.AltCostKeyMadness,
				Cost: "{X}{R}", Timing: game.TimingFlash, CastOnly: true, Label: game.MadnessTriggerLabel,
			}, []game.Card{card})
		})
		moves := castMovesFor(legal.EnumerateFor(g, active.ID), card.InstanceID)
		if len(moves) == 0 {
			t.Fatal("the madness cast is not offered")
		}
		sawX := false
		for _, m := range moves {
			d := decodeDivided(t, m)
			if d.XValue > 0 {
				sawX = true
			}
			if len(d.Targets) > d.XValue {
				t.Errorf("%q: %d targets cannot divide X=%d", m.Label, len(d.Targets), d.XValue)
			}
			checkEvenDivision(t, m.Label, d, d.XValue)
		}
		if !sawX {
			t.Error("no madness cast announced an X above 0")
		}
		if w := widestDivided(t, moves); w != 2 {
			t.Errorf("widest madness cast: %d targets, want 2 (X=3 over two creatures)", w)
		}
		dispatchAll(t, g, active.ID, moves)
	})
}
