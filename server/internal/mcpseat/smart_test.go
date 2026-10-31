package mcpseat

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// TestSmartPassFollowsTheBrowsersAutopass pins pass_until "smart" (#3603)
// against client/src/lib/autopassDecision.ts at its default settings: the
// seat's own main phases and declarations stop, the other steps of its own
// turn pass even with an instant in hand, and a stack item stops only
// when the seat holds a response to it.
func TestSmartPassFollowsTheBrowsersAutopass(t *testing.T) {
	me, opp := uuid.New(), uuid.New()
	card := uuid.New()
	pass := legal.Move{Type: legal.TypePassPriority, Kind: legal.KindPass, Label: "Pass"}
	mana := legal.Move{Type: legal.TypeActivateManaAbility, Kind: legal.KindMana, Label: "Tap"}
	land := legal.Move{Type: "play_land", Kind: legal.KindLand, Label: "Play Forest"}
	instant := legal.Move{Type: legal.TypeCastSpell, Kind: legal.KindCast, Label: "Cast Deadly Rollick", Source: card}
	// #2877: a value instant (Opt), declared to answer nothing.
	opt := legal.Move{Type: legal.TypeCastSpell, Kind: legal.KindCast, Label: "Cast Opt", Source: card, ValueOnly: true}
	value := legal.Move{Type: legal.TypeActivateAbility, Kind: legal.KindActivate, Label: "Cycle", Source: card}
	targeted := legal.Move{Type: legal.TypeActivateAbility, Kind: legal.KindActivate, Label: "Zap", Source: card, HasTargets: true}
	outlet := legal.Move{Type: legal.TypeActivateAbility, Kind: legal.KindActivate, Label: "Sacrifice", Source: card, Interacts: true}
	counter := legal.Move{Type: legal.TypeCastSpell, Kind: legal.KindCast, Label: "Counter", Source: card, TargetsStack: true}
	crew := legal.Move{Type: legal.TypeActivateAbility, Kind: legal.KindActivate, Label: "Crew", Source: card, CombatInteracts: true}
	attack := legal.Move{Type: legal.TypeDeclareAttacker, Kind: legal.KindAttack, Label: "Attack"}
	block := legal.Move{Type: legal.TypeDeclareBlocker, Kind: legal.KindBlock, Label: "Block"}
	special := legal.Move{Type: "special_action", Kind: legal.KindSpecialAction, Label: "Foretell"}

	view := func(active int, step string, mut func(*protocol.GameView)) *protocol.GameView {
		v := &protocol.GameView{State: "active",
			Seats: []protocol.PlayerView{{ID: me.String(), Seat: 0}, {ID: opp.String(), Seat: 1}},
			Turn:  protocol.TurnView{ActiveSeat: active, Step: step}}
		if mut != nil {
			mut(v)
		}
		return v
	}
	oppSpell := func(v *protocol.GameView) {
		v.StackItems = []protocol.StackItemView{{ID: "s", Kind: "spell", Controller: opp.String()}}
	}
	mySpell := func(v *protocol.GameView) {
		v.StackItems = []protocol.StackItemView{{ID: "s", Kind: "spell", Controller: me.String()}}
	}
	myTrigger := func(v *protocol.GameView) {
		v.StackItems = []protocol.StackItemView{{ID: "t", Kind: "triggered", Controller: me.String(), Label: "Tovolar: it becomes night"}}
	}
	theirAnswerOnMine := func(v *protocol.GameView) {
		v.StackItems = []protocol.StackItemView{{ID: "s", Kind: "spell", Controller: me.String()}, {ID: "c", Kind: "spell", Controller: opp.String()}}
	}
	attackerInPlay := func(v *protocol.GameView) {
		v.Battlefield.Cards = []protocol.CardView{{InstanceID: "a", Controller: opp.String(), AttackingTarget: me.String()}}
	}
	myAttackerInPlay := func(v *protocol.GameView) {
		v.Battlefield.Cards = []protocol.CardView{{InstanceID: "a", Controller: me.String(), AttackingTarget: opp.String()}}
	}
	foreignSource := func(v *protocol.GameView) {
		v.Battlefield.Cards = []protocol.CardView{{InstanceID: card.String(), Controller: opp.String()}}
	}

	for _, tc := range []struct {
		name  string
		v     *protocol.GameView
		moves []legal.Move
		want  bool // passes without the model
	}{
		// The seat's own turn.
		{"own upkeep with an instant", view(0, "upkeep", nil), []legal.Move{pass, instant}, true},
		{"own draw with a sacrifice outlet", view(0, "draw", nil), []legal.Move{pass, outlet}, true},
		{"own main with only a pass", view(0, "precombat_main", nil), []legal.Move{pass}, false},
		{"own first main with a land", view(0, "precombat_main", nil), []legal.Move{pass, land}, false},
		{"own second main", view(0, "postcombat_main", nil), []legal.Move{pass, mana}, false},
		{"own begin combat with an instant", view(0, "begin_combat", nil), []legal.Move{pass, instant}, true},
		{"own begin combat with a manland", view(0, "begin_combat", nil), []legal.Move{pass, crew}, true},
		{"own declare attackers with nothing able to attack", view(0, "declare_attackers", nil), []legal.Move{pass, instant}, true},
		{"own declare attackers with an attacker to declare", view(0, "declare_attackers", nil), []legal.Move{pass, attack}, false},
		{"own attack declared, no response", view(0, "declare_attackers", myAttackerInPlay), []legal.Move{pass, mana, value}, true},
		{"own attack declared, an instant to cast", view(0, "declare_attackers", myAttackerInPlay), []legal.Move{pass, instant}, false},
		{"own attack declared, a sacrifice outlet", view(0, "declare_attackers", myAttackerInPlay), []legal.Move{pass, outlet}, false},
		{"own declare blockers after a block, no response", view(0, "declare_blockers", myAttackerInPlay), []legal.Move{pass}, true},
		{"own combat damage", view(0, "combat_damage", myAttackerInPlay), []legal.Move{pass, instant}, true},
		{"own end step with an instant", view(0, "end", nil), []legal.Move{pass, instant}, true},

		// Another player's turn.
		{"their upkeep with an instant", view(1, "upkeep", nil), []legal.Move{pass, instant}, true},
		{"their main with an instant", view(1, "precombat_main", nil), []legal.Move{pass, instant}, true},
		{"their attack, a block is owed", view(1, "declare_blockers", attackerInPlay), []legal.Move{block, pass}, false},
		{"their attack, nothing to answer with", view(1, "declare_attackers", attackerInPlay), []legal.Move{pass, mana}, true},
		{"their attack, an instant to answer", view(1, "declare_attackers", attackerInPlay), []legal.Move{pass, instant}, false},
		{"their attack, a pure value ability", view(1, "declare_attackers", attackerInPlay), []legal.Move{pass, value}, true},
		{"their attack, crew counts in combat", view(1, "declare_attackers", attackerInPlay), []legal.Move{pass, crew}, false},
		{"their end step, nothing to answer with", view(1, "end", nil), []legal.Move{pass, mana}, true},
		{"their end step, an instant", view(1, "end", nil), []legal.Move{pass, instant}, false},
		{"their end step, a pure value ability", view(1, "end", nil), []legal.Move{pass, value}, true},
		{"their end step, an ability of theirs anyone may use", view(1, "end", foreignSource), []legal.Move{pass, targeted}, true},

		// The stack.
		{"their spell, nothing to answer with", view(1, "precombat_main", oppSpell), []legal.Move{pass, mana}, true},
		{"their spell, a pure value ability", view(1, "precombat_main", oppSpell), []legal.Move{pass, value}, true},
		{"their spell, an instant", view(1, "precombat_main", oppSpell), []legal.Move{pass, instant}, false},
		{"their spell, a value instant", view(1, "precombat_main", oppSpell), []legal.Move{pass, opt}, true},
		{"their spell, a value instant and an instant", view(1, "precombat_main", oppSpell), []legal.Move{pass, opt, instant}, false},
		{"their end step, a value instant", view(1, "end", nil), []legal.Move{pass, opt}, false},
		{"their attack, a value instant", view(1, "declare_attackers", attackerInPlay), []legal.Move{pass, opt}, false},
		{"their spell, a counterspell", view(1, "precombat_main", oppSpell), []legal.Move{pass, counter}, false},
		{"their spell, a targeted ability", view(1, "precombat_main", oppSpell), []legal.Move{pass, targeted}, false},
		{"their spell, a sacrifice outlet", view(1, "precombat_main", oppSpell), []legal.Move{pass, outlet}, false},
		{"their spell, a special action", view(1, "precombat_main", oppSpell), []legal.Move{pass, special}, false},
		{"their spell on my turn, nothing to answer with", view(0, "precombat_main", oppSpell), []legal.Move{pass}, true},

		// #4186: holding an instant. The browser passes a stack that is
		// entirely the seat's own (autoPassOwnStack, on by default) and
		// an own upkeep (not a ticked step); it stops at the seat's own
		// main phases, where the instant is a play on a ticked step.
		{"my own spell, holding an instant", view(0, "precombat_main", mySpell), []legal.Move{pass, instant, mana}, true},
		{"my own spell, holding a counterspell", view(0, "precombat_main", mySpell), []legal.Move{pass, counter}, true},
		{"my own trigger at my upkeep, holding an instant", view(0, "upkeep", myTrigger), []legal.Move{pass, instant}, true},
		{"my own spell on their turn, holding an instant", view(1, "end", mySpell), []legal.Move{pass, instant}, true},
		{"their answer on top of my spell, holding an instant", view(0, "precombat_main", theirAnswerOnMine), []legal.Move{pass, instant}, false},
		{"my upkeep after my trigger resolved, holding an instant", view(0, "upkeep", nil), []legal.Move{pass, instant, mana}, true},
		{"own first main, an instant and mana abilities", view(0, "precombat_main", nil), []legal.Move{pass, instant, mana}, false},
		{"own second main, an instant and mana abilities", view(0, "postcombat_main", nil), []legal.Move{pass, instant, mana}, false},
		{"my own spell while a trigger still queues", view(0, "precombat_main", func(v *protocol.GameView) {
			mySpell(v)
			v.PendingTriggers = []protocol.StackItemView{{}}
		}), []legal.Move{pass, instant}, false},

		// What always stops it.
		{"a pending choice", view(1, "upkeep", func(v *protocol.GameView) {
			v.PendingChoices = []protocol.PendingChoiceView{{ID: "c", Chooser: me.String()}}
		}), []legal.Move{pass}, false},
		{"a trigger still queuing", view(1, "upkeep", func(v *protocol.GameView) {
			v.PendingTriggers = []protocol.StackItemView{{}}
		}), []legal.Move{pass}, false},
		{"the loop notice", view(1, "upkeep", func(v *protocol.GameView) {
			v.LoopNotice = &protocol.LoopNoticeView{Label: "x", Count: 3}
		}), []legal.Move{pass}, false},
		{"no pass offered", view(1, "upkeep", nil), []legal.Move{instant}, false},
	} {
		i, ok := smartPassApplies(tc.v, me.String(), tc.moves)
		if ok != tc.want {
			t.Errorf("%s: passes=%v, want %v", tc.name, ok, tc.want)
			continue
		}
		if ok && tc.moves[i].Kind != legal.KindPass {
			t.Errorf("%s: answered with %q, not the pass", tc.name, tc.moves[i].Label)
		}
	}
}

// TestSmartPassReadsAnswersNow pins #4734 (ADR 0009's amendment of
// 2026-10-30): the seat reads the server's answers_now first, ahead of
// interacts, has_targets, targets_stack, combat_interacts and value_only,
// and a move without it is classed as before. Rule 7's new clause holds a
// matched move with the stack empty in any window.
func TestSmartPassReadsAnswersNow(t *testing.T) {
	me, opp := uuid.New(), uuid.New()
	card := uuid.New()
	yes, no := true, false
	pass := legal.Move{Type: legal.TypePassPriority, Kind: legal.KindPass, Label: "Pass"}
	// A cast that declares a protective answer, matched or not.
	castYes := legal.Move{Type: legal.TypeCastSpell, Kind: legal.KindCast, Label: "Cast Flawless Maneuver", Source: card, AnswersNow: &yes}
	castNo := legal.Move{Type: legal.TypeCastSpell, Kind: legal.KindCast, Label: "Cast Flawless Maneuver", Source: card, AnswersNow: &no}
	// A row that sets every old flag, so a false verdict has to beat all of them.
	rowNo := legal.Move{Type: legal.TypeActivateAbility, Kind: legal.KindActivate, Label: "Ranger-Captain", Source: card,
		Interacts: true, HasTargets: true, CombatInteracts: true, AnswersNow: &no}
	rowYes := legal.Move{Type: legal.TypeActivateAbility, Kind: legal.KindActivate, Label: "Ranger-Captain", Source: card, AnswersNow: &yes}
	counterNo := legal.Move{Type: legal.TypeCastSpell, Kind: legal.KindCast, Label: "Counter", Source: card, TargetsStack: true, AnswersNow: &no}
	manaYes := legal.Move{Type: legal.TypeActivateManaAbility, Kind: legal.KindMana, Label: "Sac for mana", Source: card, AnswersNow: &yes}
	manaNo := legal.Move{Type: legal.TypeActivateManaAbility, Kind: legal.KindMana, Label: "Sac for mana", Source: card, Interacts: true, AnswersNow: &no}
	// The same flags without a verdict: today's classes.
	rowOld := legal.Move{Type: legal.TypeActivateAbility, Kind: legal.KindActivate, Label: "Old row", Source: card, Interacts: true}
	manaOld := legal.Move{Type: legal.TypeActivateManaAbility, Kind: legal.KindMana, Label: "Old outlet", Source: card, Interacts: true}
	crewOld := legal.Move{Type: legal.TypeActivateAbility, Kind: legal.KindActivate, Label: "Crew", Source: card, CombatInteracts: true}
	special := legal.Move{Type: "special_action", Kind: legal.KindSpecialAction, Label: "Foretell"}
	// A value cast has no verdict (declared exactly value).
	opt := legal.Move{Type: legal.TypeCastSpell, Kind: legal.KindCast, Label: "Cast Opt", Source: card, ValueOnly: true}

	view := func(active int, step string, mut func(*protocol.GameView)) *protocol.GameView {
		v := &protocol.GameView{State: "active",
			Seats: []protocol.PlayerView{{ID: me.String(), Seat: 0}, {ID: opp.String(), Seat: 1}},
			Turn:  protocol.TurnView{ActiveSeat: active, Step: step}}
		if mut != nil {
			mut(v)
		}
		return v
	}
	oppSpell := func(v *protocol.GameView) {
		v.StackItems = []protocol.StackItemView{{ID: "s", Kind: "spell", Controller: opp.String()}}
	}
	oppAbility := func(v *protocol.GameView) {
		v.StackItems = []protocol.StackItemView{{ID: "t", Kind: "triggered", Controller: opp.String(), Label: "Smaug: deals damage"}}
	}
	attackOnMe := func(v *protocol.GameView) {
		v.Battlefield.Cards = []protocol.CardView{{InstanceID: "a", Controller: opp.String(), AttackingTarget: me.String()}}
	}
	myAttack := func(v *protocol.GameView) {
		v.Battlefield.Cards = []protocol.CardView{{InstanceID: "a", Controller: me.String(), AttackingTarget: opp.String()}}
	}
	foreignSource := func(v *protocol.GameView) {
		v.Battlefield.Cards = []protocol.CardView{{InstanceID: card.String(), Controller: opp.String()}}
	}

	for _, tc := range []struct {
		name  string
		v     *protocol.GameView
		moves []legal.Move
		want  bool // passes without the model
	}{
		// With something of an opponent's on the stack: true holds, false
		// passes, whatever the old flags say.
		{"their spell, a matched cast", view(1, "precombat_main", oppSpell), []legal.Move{pass, castYes}, false},
		{"their spell, an unmatched cast", view(1, "precombat_main", oppSpell), []legal.Move{pass, castNo}, true},
		{"their spell, a matched row", view(1, "precombat_main", oppSpell), []legal.Move{pass, rowYes}, false},
		{"their spell, an unmatched row with every old flag", view(1, "precombat_main", oppSpell), []legal.Move{pass, rowNo}, true},
		{"their spell, an unmatched counter", view(1, "precombat_main", oppSpell), []legal.Move{pass, counterNo}, true},
		{"their spell, a matched mana move", view(1, "precombat_main", oppSpell), []legal.Move{pass, manaYes}, false},
		{"their spell, an unmatched mana move that interacts", view(1, "precombat_main", oppSpell), []legal.Move{pass, manaNo}, true},
		{"their ability, a matched cast", view(1, "precombat_main", oppAbility), []legal.Move{pass, castYes}, false},
		{"their ability, an unmatched cast and an unmatched row", view(1, "precombat_main", oppAbility), []legal.Move{pass, castNo, rowNo}, true},
		{"their ability, an unmatched row beside a special action", view(1, "precombat_main", oppAbility), []legal.Move{pass, rowNo, special}, false},
		{"their spell, an unmatched cast beside a matched row", view(1, "precombat_main", oppSpell), []legal.Move{pass, castNo, rowYes}, false},
		{"their spell on my turn, a matched cast", view(0, "precombat_main", oppSpell), []legal.Move{pass, castYes}, false},
		{"their spell on my turn, an unmatched cast", view(0, "precombat_main", oppSpell), []legal.Move{pass, castNo}, true},

		// No verdict: classed as before.
		{"their spell, an old row", view(1, "precombat_main", oppSpell), []legal.Move{pass, rowOld}, false},
		{"their spell, an old sacrifice outlet that makes mana", view(1, "precombat_main", oppSpell), []legal.Move{pass, manaOld}, false},
		{"their spell, a value cast", view(1, "precombat_main", oppSpell), []legal.Move{pass, opt}, true},

		// The empty stack, key windows: the verdict decides.
		{"their attack, a matched cast", view(1, "declare_attackers", attackOnMe), []legal.Move{pass, castYes}, false},
		{"their attack, an unmatched cast", view(1, "declare_attackers", attackOnMe), []legal.Move{pass, castNo}, true},
		{"their attack, an unmatched row that crews", view(1, "declare_attackers", attackOnMe), []legal.Move{pass, rowNo}, true},
		{"their attack, an old crew", view(1, "declare_attackers", attackOnMe), []legal.Move{pass, crewOld}, false},
		{"my attack, a matched cast", view(0, "declare_blockers", myAttack), []legal.Move{pass, castYes}, false},
		{"their end step, a matched row", view(1, "end", nil), []legal.Move{pass, rowYes}, false},
		{"their end step, an unmatched row", view(1, "end", nil), []legal.Move{pass, rowNo}, true},
		{"their end step, an unmatched cast", view(1, "end", nil), []legal.Move{pass, castNo}, true},
		{"their end step, an old row", view(1, "end", nil), []legal.Move{pass, rowOld}, false},

		// Rule 7's new clause: a matched move holds with the stack empty
		// in any other window, and an unmatched one does not.
		{"their upkeep, a matched row", view(1, "upkeep", nil), []legal.Move{pass, rowYes}, false},
		{"their upkeep, a matched cast", view(1, "upkeep", nil), []legal.Move{pass, castYes}, false},
		{"their upkeep, a matched mana move", view(1, "upkeep", nil), []legal.Move{pass, manaYes}, false},
		{"their upkeep, an unmatched row", view(1, "upkeep", nil), []legal.Move{pass, rowNo}, true},
		{"their upkeep, an old row", view(1, "upkeep", nil), []legal.Move{pass, rowOld}, true},
		{"their draw step, a matched row", view(1, "draw", nil), []legal.Move{pass, rowYes}, false},
		{"their main phase, a matched cast", view(1, "precombat_main", nil), []legal.Move{pass, castYes}, false},
		{"their begin combat, a matched cast", view(1, "begin_combat", nil), []legal.Move{pass, castYes}, false},
		{"first-strike damage step, a matched cast", view(1, "first_strike_damage", attackOnMe), []legal.Move{pass, castYes}, false},
		{"first-strike damage step, an unmatched cast", view(1, "first_strike_damage", attackOnMe), []legal.Move{pass, castNo}, true},
		{"first-strike damage step, an old row", view(1, "first_strike_damage", attackOnMe), []legal.Move{pass, rowOld}, true},
		{"regular damage step, a matched cast", view(1, "combat_damage", attackOnMe), []legal.Move{pass, castYes}, false},
		{"regular damage step, an unmatched cast", view(1, "combat_damage", attackOnMe), []legal.Move{pass, castNo}, true},
		{"declare blockers with no attack, a matched cast", view(1, "declare_blockers", nil), []legal.Move{pass, castYes}, false},
		{"my begin combat, a matched cast", view(0, "begin_combat", nil), []legal.Move{pass, castYes}, false},
		{"my begin combat, an old crew", view(0, "begin_combat", nil), []legal.Move{pass, crewOld}, true},
		{"my upkeep, an unmatched cast", view(0, "upkeep", nil), []legal.Move{pass, castNo}, true},

		// A permanent another player controls is never a reason to stop
		// (ADR 0106 §1), matched or not.
		{"their upkeep, a matched row of theirs anyone may use", view(1, "upkeep", foreignSource), []legal.Move{pass, rowYes}, true},
		{"their spell, a matched row of theirs anyone may use", view(1, "precombat_main", func(v *protocol.GameView) {
			oppSpell(v)
			foreignSource(v)
		}), []legal.Move{pass, rowYes}, true},

		// The seat's own main phases stop whatever the verdict, and so does
		// everything that always stops it.
		{"my first main, an unmatched cast", view(0, "precombat_main", nil), []legal.Move{pass, castNo}, false},
		{"my second main, an unmatched row", view(0, "postcombat_main", nil), []legal.Move{pass, rowNo}, false},
		{"my own spell, a matched cast", view(0, "precombat_main", func(v *protocol.GameView) {
			v.StackItems = []protocol.StackItemView{{ID: "s", Kind: "spell", Controller: me.String()}}
		}), []legal.Move{pass, castYes}, true},
		{"a matched cast with a choice owed", view(1, "upkeep", func(v *protocol.GameView) {
			v.PendingChoices = []protocol.PendingChoiceView{{ID: "c", Chooser: me.String()}}
		}), []legal.Move{pass, castNo}, false},
	} {
		i, ok := smartPassApplies(tc.v, me.String(), tc.moves)
		if ok != tc.want {
			t.Errorf("%s: passes=%v, want %v", tc.name, ok, tc.want)
			continue
		}
		if ok && tc.moves[i].Kind != legal.KindPass {
			t.Errorf("%s: answered with %q, not the pass", tc.name, tc.moves[i].Label)
		}
	}
}

// agreementFrame is one frame of the shared fixture, as
// internal/legal's smart_pass_agreement_test.go writes it.
type agreementFrame struct {
	Name    string            `json:"name"`
	Viewer  string            `json:"viewer"`
	View    protocol.GameView `json:"view"`
	Verdict string            `json:"verdict"`
	Why     string            `json:"why"`
}

// TestSmartPassAgreesWithTheSharedFixture is ADR 0009's agreement test
// for the seat (#4734): every frame of internal/legal's
// smart_pass_agreement.json, a real wire frame with its legal moves, has
// a stop or pass verdict declared by hand at the default settings, and
// the seat must reach it. client/src/lib/smartPassAgreement.test.ts
// asserts the browser's. Either side drifting fails a test, and neither
// is the other's oracle: when the seat disagrees, work out which side is
// wrong before touching either. The fixture lives in internal/legal
// because this package may not import internal/game (imports_test.go);
// regenerate it there with
//
//	go test ./internal/legal/ -run TestSmartPassAgreement -update
func TestSmartPassAgreesWithTheSharedFixture(t *testing.T) {
	body, err := os.ReadFile("../legal/testdata/smart_pass_agreement.json")
	if err != nil {
		t.Fatalf("read the shared fixture: %v", err)
	}
	var frames []agreementFrame
	if err := json.Unmarshal(body, &frames); err != nil {
		t.Fatalf("unmarshal the shared fixture: %v", err)
	}
	if len(frames) == 0 {
		t.Fatal("the shared fixture is empty")
	}
	stops, passes := 0, 0
	for _, f := range frames {
		t.Run(f.Name, func(t *testing.T) {
			var want bool
			switch f.Verdict {
			case "stop":
				stops++
			case "pass":
				want = true
				passes++
			default:
				t.Fatalf("verdict %q is neither stop nor pass", f.Verdict)
			}
			if len(f.View.LegalMoves) == 0 {
				t.Fatal("the frame carries no legal moves")
			}
			i, ok := smartPassApplies(&f.View, f.Viewer, f.View.LegalMoves)
			if ok != want {
				t.Fatalf("the seat's smart pass: passes=%v, the fixture declares %q (%s)", ok, f.Verdict, f.Why)
			}
			if ok && f.View.LegalMoves[i].Kind != legal.KindPass {
				t.Errorf("answered with %q, not the pass", f.View.LegalMoves[i].Label)
			}
		})
	}
	if stops == 0 || passes == 0 {
		t.Errorf("the fixture should hold both verdicts: %d stops, %d passes", stops, passes)
	}
}

// TestSmartPassDoesNotGoThroughMyTurnOrStack keeps the older mode as it
// was: it still stops for every stack item and for the seat's own turn.
func TestSmartPassDoesNotGoThroughMyTurnOrStack(t *testing.T) {
	me := uuid.New()
	v := &protocol.GameView{State: "active",
		Seats: []protocol.PlayerView{{ID: me.String(), Seat: 0}, {ID: uuid.NewString(), Seat: 1}},
		Turn:  protocol.TurnView{ActiveSeat: 0, Step: "upkeep"}}
	pass := legal.Move{Type: legal.TypePassPriority, Kind: legal.KindPass}
	def, _ := ParseAbsorb("none")
	if _, ok := decideAutomatically(v, me, []legal.Move{pass, {Kind: legal.KindCast}}, def, passOthers); ok {
		t.Error("my_turn_or_stack passed an idle step of the seat's own turn")
	}
	if ans, ok := decideAutomatically(v, me, []legal.Move{pass, {Kind: legal.KindCast}}, def, passSmart); !ok || ans.rule != rulePassUntil {
		t.Errorf("smart on an idle own step: %+v %v", ans, ok)
	}
}

func TestParsePassUntil(t *testing.T) {
	for in, want := range map[string]passMode{"": passNone, "none": passNone, "my_turn_or_stack": passOthers, "smart": passSmart} {
		if got, ok := parsePassUntil(in); !ok || got != want {
			t.Errorf("parsePassUntil(%q) = %v %v, want %v", in, got, ok, want)
		}
	}
	if _, ok := parsePassUntil("always"); ok {
		t.Error("an unknown pass_until was accepted")
	}
}

func TestWaitForDecisionRejectsAnUnknownPassUntil(t *testing.T) {
	f := newFakeServer(t)
	s := newTestSeat(t, nil)
	joinFake(t, f, s)
	r, _ := s.WaitForDecision(context.Background(), WaitInput{PassUntil: "always"})
	if !r.IsError {
		t.Fatalf("pass_until \"always\" was accepted: %s", r.Text)
	}
}

// TestSmartPassPassesIdleStepsOfTheSeatsOwnTurnAndKeepsThroughTheTurn is
// the round trip: smart passes an idle own step with an instant in hand,
// survives the start of the seat's own turn (my_turn_or_stack does not),
// and still gives the seat its main phase.
//
// Nothing here waits on timing. The table starts in an opponent's end step
// holding an instant, which is a real decision (no Layer A pass goes out,
// so the fake server has seen no action), the seat is switched to smart,
// and each later step is awaited by the observable it produces: exactly
// one action for the idle upkeep, then the decision window of the snapshot
// just sent for the main phase. The earlier version opened with a forced
// window whose Layer A pass could reach the fake server after the count
// was read, which made the final count comparison race.
func TestSmartPassPassesIdleStepsOfTheSeatsOwnTurnAndKeepsThroughTheTurn(t *testing.T) {
	f := newFakeServer(t)
	s := newTestSeat(t, nil)
	joinFake(t, f, s)

	theirs := activeView(f)
	theirs.Turn.Step = "end"
	f.setState(theirs, twoChoices(f), false)
	waitFor(t, "the opponent's end step to open a decision", func() bool {
		s.mu.Lock()
		defer s.mu.Unlock()
		return s.win != nil && s.win.state == winDecision && s.view != nil && s.view.Turn.Step == "end"
	})
	if n := f.actionCount(); n != 0 {
		t.Fatalf("%d actions before smart was switched on", n)
	}
	s.mu.Lock()
	s.passMode = passSmart
	s.mu.Unlock()

	// The seat's own turn starts at upkeep.
	mine := activeView(f)
	mine.Turn.ActiveSeat = 0
	mine.Turn.Step = "upkeep"
	f.setState(mine, twoChoices(f), false)
	waitFor(t, "the upkeep pass", func() bool { return f.actionCount() == 1 })
	if a := f.lastAction(); a.Type != legal.TypePassPriority {
		t.Fatalf("smart sent %+v at the seat's own upkeep", a)
	}
	s.mu.Lock()
	still := s.passMode
	s.mu.Unlock()
	if still != passSmart {
		t.Errorf("the start of the seat's own turn cleared smart (mode %v)", still)
	}

	main := mine
	main.Turn.Step = "precombat_main"
	f.setState(main, twoChoices(f), false)
	f.mu.Lock()
	seq := f.seq
	f.mu.Unlock()
	waitFor(t, "the main-phase snapshot's decision window", func() bool {
		s.mu.Lock()
		defer s.mu.Unlock()
		return s.win != nil && s.win.state == winDecision && s.win.key.seq == seq
	})
	if n := f.actionCount(); n != 1 {
		t.Errorf("smart passed the seat's own main phase (%d actions, want the upkeep pass alone)", n)
	}
}
