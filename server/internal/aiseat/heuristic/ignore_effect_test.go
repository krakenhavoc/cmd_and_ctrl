package heuristic

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// ignore_effect_test.go — #1986: the bot sheds a permanent to ignore
// Damping Engine only for something it would do this turn.

func ignoreEngineTable(step string, hand ...protocol.CardView) (*Policy, *state, legal.Move) {
	me, engine := uuid.New(), uuid.New()
	var v protocol.GameView
	v.Seats = []protocol.PlayerView{{
		ID:                  me.String(),
		LandDropsPerTurn:    1,
		LandsPlayedThisTurn: 0,
		Hand:                protocol.ZoneView{Cards: hand},
	}}
	v.Battlefield.Cards = []protocol.CardView{{
		InstanceID: engine.String(), Name: "Damping Engine", TypeLine: "Artifact",
		Owner: me.String(), Controller: me.String(),
	}}
	v.Turn.ActiveSeat = 0
	v.Turn.Step = step
	raw, _ := json.Marshal(map[string]any{
		"card_id": engine.String(), "kind": "ignore_effect", "sacrifice_id": uuid.NewString(),
	})
	move := legal.Move{Type: legal.TypeSpecialAction, Kind: legal.KindSpecialAction, Source: engine, Params: raw}
	p := New()
	return p, p.newState(aiseat.Input{Seat: me, View: v}), move
}

func TestTheBotIgnoresTheEngineForALandItHolds(t *testing.T) {
	land := protocol.CardView{
		InstanceID: uuid.NewString(), Name: "Forest", TypeLine: "Basic Land — Forest",
		CantCast: "A player who controls more permanents than each other player can't play lands. — Damping Engine",
	}
	p, st, move := ignoreEngineTable("precombat_main", land)
	got, reason := p.payoffOf(st, move)
	if got != p.cfg.LandValue {
		t.Errorf("ignoring for a land = %v (%s), want a land drop's %v", got, reason, p.cfg.LandValue)
	}
}

func TestTheBotDoesNotIgnoreTheEngineForNothing(t *testing.T) {
	refusedLand := protocol.CardView{
		InstanceID: uuid.NewString(), Name: "Forest", TypeLine: "Basic Land — Forest",
		CantCast: "can't play lands. — Damping Engine",
	}
	freeLand := protocol.CardView{InstanceID: uuid.NewString(), Name: "Island", TypeLine: "Basic Land — Island"}
	for name, tc := range map[string]struct {
		step string
		hand []protocol.CardView
	}{
		"empty hand":                {"precombat_main", nil},
		"a card nothing refuses":    {"precombat_main", []protocol.CardView{freeLand}},
		"not a main phase":          {"upkeep", []protocol.CardView{refusedLand}},
		"refused by another effect": {"precombat_main", []protocol.CardView{{InstanceID: uuid.NewString(), Name: "Forest", TypeLine: "Basic Land — Forest", CantCast: "Players can't play lands. — Territorial Dispute"}}},
	} {
		p, st, move := ignoreEngineTable(tc.step, tc.hand...)
		if got, reason := p.payoffOf(st, move); got >= 0 {
			t.Errorf("%s: ignoring = %v (%s), want below passing", name, got, reason)
		}
	}
	// No land drop left: the land can't be played this turn anyway.
	p, st, move := ignoreEngineTable("precombat_main", refusedLand)
	st.seat.LandsPlayedThisTurn = 1
	if got, reason := p.payoffOf(st, move); got >= 0 {
		t.Errorf("no land drop left: ignoring = %v (%s), want below passing", got, reason)
	}
}

func TestTheBotIgnoresTheEngineForASpellItCanCast(t *testing.T) {
	spell := protocol.CardView{
		InstanceID: uuid.NewString(), Name: "Grizzly Bears", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
		CantCast: "can't cast artifact, creature, or enchantment spells. — Damping Engine",
	}
	p, st, move := ignoreEngineTable("precombat_main", spell)
	st.myMana = 1
	if got, _ := p.payoffOf(st, move); got >= 0 {
		t.Errorf("a two-mana spell with one mana: ignoring = %v, want below passing", got)
	}
	st.myMana = 2
	if got, reason := p.payoffOf(st, move); got <= 0 {
		t.Errorf("a castable refused spell: ignoring = %v (%s), want positive", got, reason)
	}
}
