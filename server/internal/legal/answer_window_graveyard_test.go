package legal_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// answer_window_graveyard_test.go — ADR 0009's amendment of 2026-10-30
// (#4734), owner decision Q5 (a), slice 9: `graveyard` is split out of
// `restrict`, as `survive` was split from `protect`. These tests run the
// untargeted graveyard-exile rows of the real catalog (Soul-Guide Lantern,
// Scavenger Grounds, Silent Gravestone) beside Ranger-Captain of Eos,
// which keeps `restrict`, through the windows where the two kinds part:
//
//   - an opponent's reanimation trigger on a card in a graveyard matches
//     the exile and not the restriction: the trigger is an ability, and
//     restrict never matches one;
//   - an opponent's upkeep with nothing on the stack matches the
//     restriction and not the exile: the exile answers a card in a
//     graveyard, and none is on its way there;
//   - an opponent's spell matches both.
//
// TestAnswersNowByKind pins the kinds one at a time on declared relics;
// this pins the catalog's re-declarations to the same windows.

const (
	oracleSoulGuideLantern = "1b5e6560-ff2e-4475-96cb-63f64c8a86db"
	oracleScavengerGrounds = "5ece7d03-9ee7-4953-a06e-9d8e41874903"
	oracleSilentGravestone = "803815c5-12be-48b3-a101-f416c1ef9b7b"
)

// graveyardRows are the catalog rows under test: the card, and a piece of
// the activation's label that picks the row out of the card's others
// (Soul-Guide Lantern also has a draw row, declared value).
var graveyardRows = []struct{ name, label string }{
	{"Soul-Guide Lantern", "Exile each opponent's graveyard"},
	{"Scavenger Grounds", "Exile all graveyards"},
	{"Silent Gravestone", "all cards from all graveyards"},
}

func TestAnswersNowGraveyardRowsOfTheCatalog(t *testing.T) {
	reanimation := func(t *testing.T, b *kindBoard) {
		src := battlefieldCard(b.g, b.active, creature("Trigger Source", "{1}{G}{G}", 2, 1))
		card := graveyardCard(b.active, creature("Fallen Bear", "{1}{G}", 2, 2))
		pushAbilityAW(b.g, b.active, src, game.StackItemTriggered, "Trigger Source — return target card from your graveyard to your hand", nil,
			game.TargetRef{Kind: game.TargetCard, ID: card})
	}
	cases := []struct {
		name   string
		offset int
		step   game.Step
		push   func(t *testing.T, b *kindBoard)
		// graveyard is the verdict of the three exile rows, restrict
		// Ranger-Captain's.
		graveyard, restrict bool
	}{
		{"an opponent's reanimation trigger on a card in a graveyard", 1, game.StepPrecombatMain, reanimation, true, false},
		{"an opponent's creature spell", 1, game.StepPrecombatMain, func(t *testing.T, b *kindBoard) {
			pushSpellAW(b.g, b.active, creature("Opposing Bear", "{1}{G}", 2, 2))
		}, true, true},
		{"an opponent's untargeted ability", 1, game.StepPrecombatMain, func(t *testing.T, b *kindBoard) {
			src := battlefieldCard(b.g, b.active, creature("Trigger Source", "{G}", 1, 1))
			pushAbilityAW(b.g, b.active, src, game.StackItemTriggered, "Trigger Source — something", nil)
		}, false, false},
		{"an opponent's ability on a creature of another player's", 1, game.StepPrecombatMain, func(t *testing.T, b *kindBoard) {
			src := battlefieldCard(b.g, b.active, creature("Trigger Source", "{G}", 1, 1))
			pushAbilityAW(b.g, b.active, src, game.StackItemTriggered, "Trigger Source — damage", nil,
				game.TargetRef{Kind: game.TargetCard, ID: b.otherBear})
		}, false, false},
		{"an opponent's upkeep, nothing on the stack", 3, game.StepUpkeep, nil, false, true},
		{"an opponent's end step, nothing on the stack", 1, game.StepEnd, nil, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b := newKindBoard(t, tc.offset, false)
			for range 6 {
				battlefieldCard(b.g, b.seat, basic("Island", "Island"))
			}
			rows := map[string]uuid.UUID{
				"Soul-Guide Lantern": battlefieldCard(b.g, b.seat, game.Card{Name: "Soul-Guide Lantern", TypeLine: "Artifact", ManaCost: "{1}", OracleID: oracleSoulGuideLantern}),
				"Scavenger Grounds":  battlefieldCard(b.g, b.seat, game.Card{Name: "Scavenger Grounds", TypeLine: "Land — Desert", OracleID: oracleScavengerGrounds}),
				"Silent Gravestone":  battlefieldCard(b.g, b.seat, game.Card{Name: "Silent Gravestone", TypeLine: "Artifact", ManaCost: "{3}", OracleID: oracleSilentGravestone}),
			}
			captain := battlefieldCard(b.g, b.seat, game.Card{
				Name: "Ranger-Captain of Eos", TypeLine: "Creature — Human Soldier Ranger", OracleID: oracleRangerCaptainAW,
				ManaCost: "{1}{W}{W}", Power: 3, Toughness: 3,
			})
			advanceTo(t, b.g, tc.step)
			if tc.push != nil {
				tc.push(t, b)
			}
			givePriority(b.g, b.seat)
			moves := legal.EnumerateFor(b.g, b.seat.ID)
			for _, r := range graveyardRows {
				wantRowVerdict(t, moves, r.name, rows[r.name], r.label, tc.graveyard)
			}
			wantRowVerdict(t, moves, "Ranger-Captain of Eos", captain, "Your opponents can't cast noncreature spells", tc.restrict)
		})
	}
}

// wantRowVerdict asserts the answers_now of every activation from
// `source` whose label contains `label`, and that there is one.
func wantRowVerdict(t *testing.T, moves []legal.Move, name string, source uuid.UUID, label string, want bool) {
	t.Helper()
	n := 0
	for _, m := range movesFrom(moves, source, legal.KindActivate) {
		if !strings.Contains(m.Label, label) {
			continue
		}
		n++
		if verdictString(m.AnswersNow) != verdictString(&want) {
			t.Errorf("%s: answers_now %s, want %v (%s)", name, verdictString(m.AnswersNow), want, m.Label)
		}
	}
	if n == 0 {
		t.Errorf("%s offered no activation labelled %q: %v", name, label, labels(moves))
	}
}

// TestAnswersNowGraveyardKeepsTheOldFlags: the split changes answers_now
// and nothing else. A graveyard row is a stack-tier answer that no threat
// gates, as restrict was, so it still sets interacts on an opponent's item
// that matches nothing, for every reader that does not know answers_now.
func TestAnswersNowGraveyardKeepsTheOldFlags(t *testing.T) {
	b := newKindBoard(t, 1, false)
	row := battlefieldCard(b.g, b.seat, game.Card{Name: "Soul-Guide Lantern", TypeLine: "Artifact", ManaCost: "{1}", OracleID: oracleSoulGuideLantern})
	advanceTo(t, b.g, game.StepPrecombatMain)
	src := battlefieldCard(b.g, b.active, creature("Trigger Source", "{G}", 1, 1))
	pushAbilityAW(b.g, b.active, src, game.StackItemTriggered, "Trigger Source — something", nil)
	givePriority(b.g, b.seat)
	n := 0
	for _, m := range movesFrom(legal.EnumerateFor(b.g, b.seat.ID), row, legal.KindActivate) {
		if !strings.Contains(m.Label, "Exile each opponent's graveyard") {
			continue
		}
		n++
		if !m.Interacts || m.CombatInteracts {
			t.Errorf("interacts=%v combat_interacts=%v, want true/false (%s)", m.Interacts, m.CombatInteracts, m.Label)
		}
		if m.AnswersNow == nil || *m.AnswersNow {
			t.Errorf("answers_now %s, want false (%s)", verdictString(m.AnswersNow), m.Label)
		}
	}
	if n == 0 {
		t.Fatal("Soul-Guide Lantern offered no exile activation")
	}
}

// TestAnswersNowGraveyardUnion: a row that declares graveyard beside
// another kind matches in either one's window and in neither's does not
// (ADR 0142 sweep ruling 4). Here survive: the exile answers a graveyard
// target, the indestructibility answers removal on a creature.
func TestAnswersNowGraveyardUnion(t *testing.T) {
	both := game.AnswerSurvive | game.AnswerGraveyard
	cases := []struct {
		name   string
		offset int
		step   game.Step
		push   func(t *testing.T, b *kindBoard)
		want   bool
	}{
		{"a reanimation trigger on a graveyard card (graveyard)", 1, game.StepPrecombatMain, func(t *testing.T, b *kindBoard) {
			src := battlefieldCard(b.g, b.active, creature("Trigger Source", "{1}{G}{G}", 2, 1))
			card := graveyardCard(b.active, creature("Fallen Bear", "{1}{G}", 2, 2))
			pushAbilityAW(b.g, b.active, src, game.StackItemTriggered, "Trigger Source — return target card from your graveyard to your hand", nil,
				game.TargetRef{Kind: game.TargetCard, ID: card})
		}, true},
		{"an ability aimed at your creature (survive)", 1, game.StepPrecombatMain, func(t *testing.T, b *kindBoard) {
			pushAbilityAW(b.g, b.active, b.otherBear, game.StackItemTriggered, "Other Bear — damage", nil,
				game.TargetRef{Kind: game.TargetCard, ID: b.bear})
		}, true},
		{"an untargeted ability (neither)", 1, game.StepPrecombatMain, func(t *testing.T, b *kindBoard) {
			pushAbilityAW(b.g, b.active, b.otherBear, game.StackItemTriggered, "Other Bear — something", nil)
		}, false},
		{"an opponent's upkeep (neither: that is restrict's)", 3, game.StepUpkeep, nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b := newKindBoard(t, tc.offset, false)
			row := battlefieldCard(b.g, b.seat, declaredRelic(both))
			advanceTo(t, b.g, tc.step)
			if tc.push != nil {
				tc.push(t, b)
			}
			givePriority(b.g, b.seat)
			wantVerdict(t, legal.EnumerateFor(b.g, b.seat.ID), "survive|graveyard", row, legal.KindActivate, verdict(tc.want))
		})
	}
}
