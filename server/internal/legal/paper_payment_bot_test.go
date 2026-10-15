package legal_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// #3855: the bot never takes the paper payment path. Every cast and
// activation the enumerator offers is strict and auto-tapped — it pays
// from the board or is not offered — so it never waives a payment, never
// sets force_cast, never sends paper_payment, and a dispatch of it is
// never refused for want of a declaration. Pinned on a board that reads
// everything: a Vexing Bauble on the table (every cast is read) and a
// converge spell and a Protective Sphere (its activation reads its mana).
func TestBotMovesNeverWaiveAPaymentSomethingReads(t *testing.T) {
	const (
		oraclePainfulTruths    = "58a2d15f-1b4b-4303-b524-8cc0d1b3e7d2"
		oracleVexingBauble     = "4514777d-0589-4631-978b-ff244167c176"
		oracleProtectiveSphere = "2037c4bf-bf0d-46ab-8b4b-61d9a50c431e"
	)
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	basicLands(g, active, 2, "Swamp")
	basicLands(g, active, 1, "Island")
	handCard(active, game.Card{Name: "Painful Truths", TypeLine: "Sorcery", ManaCost: "{2}{B}", OracleID: oraclePainfulTruths})
	handCard(active, game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	battlefieldCard(g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)], game.Card{Name: "Vexing Bauble", TypeLine: "Artifact", OracleID: oracleVexingBauble})
	battlefieldCard(g, active, game.Card{Name: "Protective Sphere", TypeLine: "Enchantment", OracleID: oracleProtectiveSphere})
	advanceTo(t, g, game.StepPrecombatMain)

	checked := 0
	for _, m := range legal.EnumerateFor(g, active.ID) {
		if m.Type != string(actions.TypeCastSpell) && m.Type != string(actions.TypeActivateAbility) {
			continue
		}
		var p struct {
			Strict       bool            `json:"strict"`
			ForceCast    bool            `json:"force_cast"`
			PaperPayment json.RawMessage `json:"paper_payment"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("%q: params: %v", m.Label, err)
		}
		if p.ForceCast || len(p.PaperPayment) > 0 {
			t.Errorf("%q waives or declares a payment: %s", m.Label, m.Params)
		}
		if m.Type == string(actions.TypeCastSpell) && !p.Strict {
			t.Errorf("%q is not a strict cast: %s", m.Label, m.Params)
		}
		err := actions.Dispatch(g.Clone(), actions.Action{
			Type: actions.Type(m.Type), Player: m.Player, Caller: active.ID, Params: m.Params,
		})
		if errors.Is(err, game.ErrPaperPaymentRequired) {
			t.Errorf("%q was refused for a paper payment declaration", m.Label)
		}
		checked++
	}
	if checked == 0 {
		t.Fatal("no cast or activation was offered; the test checks nothing")
	}
}
