package protocol

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// untap_opponent_mana_view_test.go — #6208: "{T}, Untap a tapped land an
// opponent controls" (game.TapOthersCost.Untap with OpponentControls) on
// the wire. It is the first untap cost a MANA ability may print, so
// `mana_abilities[i]` carries the tap-another component's label and
// options plus `tap_others_untap` for the picker's verb, and the answer
// rides `activate_mana_ability`'s `tap_ids` as before.

// benthicLandClause is creatureCostClause one card type over: "a tapped
// land an opponent controls", with the owner left to the engine.
func benthicLandClause() *game.TargetSpec {
	return &game.TargetSpec{
		Mode:  "permanent",
		Label: "a tapped land an opponent controls",
		Zones: []game.ZoneKind{game.ZoneBattlefield},
		CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return c.HasCardType("land")
		},
		Min: 1, Max: 1,
	}
}

// seatBenthicExplorers seats Benthic Explorers' mana ability: the untap
// cost with the owner turned over, and an output read off the land the
// payment untapped.
func seatBenthicExplorers(g *game.Game, owner uuid.UUID) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       "Benthic Explorers",
		TypeLine:   "Creature — Merfolk",
		Power:      2, Toughness: 3,
		Owner:      owner,
		Controller: owner,
		ManaAbilities: []game.ManaAbilityShape{{
			TapCost: true,
			TapOthers: &game.TapOthersCost{
				Count:            1,
				Filter:           benthicLandClause(),
				Label:            "a tapped land an opponent controls",
				Untap:            true,
				OpponentControls: true,
			},
			ProducedForPaid: func(g *game.Game, _, _ uuid.UUID, paid game.PaidCost) game.ManaOutput {
				if len(paid.UntappedOthers) != 1 {
					return game.ManaOutput{}
				}
				land, ok := g.LookupCardForEffect(paid.UntappedOthers[0].ID)
				if !ok {
					return game.ManaOutput{}
				}
				types := g.ProducibleManaLocked(land)
				if len(types) == 0 {
					return game.ManaOutput{}
				}
				return game.ManaOutput{Amount: game.ManaAmount{Options: types, N: 1, OneColor: true}}
			},
			DerivesFromOtherSources: true,
			Label:                   "Add one mana of any type that land could produce",
		}},
	})
	return id
}

func seatTappedLandFor(g *game.Game, controller uuid.UUID, name, typeLine string, tapped bool, keywords ...string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, Tapped: tapped,
		Owner: controller, Controller: controller, Keywords: keywords,
	})
	return id
}

// The row says it untaps, the options are the OPPONENT's tapped lands —
// a hexproof one among them, since paying a cost does not target
// (CR 601.2h) — and the viewer's own lands are never on the list.
func TestManaAbilityViewCarriesAnUntapOpponentCost(t *testing.T) {
	g := busyTable(t, 0)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	src := seatBenthicExplorers(g, me.ID)
	theirIsland := seatTappedLandFor(g, opp.ID, "Their Island", "Land — Island", true)
	theirHexproof := seatTappedLandFor(g, opp.ID, "Their Hexproof Forest", "Land — Forest", true, "hexproof")
	seatTappedLandFor(g, opp.ID, "Their Untapped Swamp", "Land — Swamp", false) // untapped: cannot pay
	mine := seatTappedLandFor(g, me.ID, "My Island", "Land — Island", true)     // mine: cannot pay
	g.BumpLayerVersionForTest()

	rows := vehicleView(t, g, src).ManaAbilities
	if len(rows) != 1 {
		t.Fatalf("mana ability rows = %d, want 1", len(rows))
	}
	row := rows[0]
	if !row.TapOthersUntap {
		t.Error("tap_others_untap is not set on the untap form")
	}
	if row.TapOthersLabel != "a tapped land an opponent controls" {
		t.Errorf("tap_others_label = %q, want the clause as printed", row.TapOthersLabel)
	}
	if row.TapOthersOptions == nil || row.TapOthersOptions.Min != 1 || row.TapOthersOptions.Max != 1 {
		t.Fatalf("tap_others_options = %+v, want a fixed one-card choice", row.TapOthersOptions)
	}
	shown := map[string]bool{}
	for _, id := range row.TapOthersOptions.Cards {
		shown[id] = true
	}
	for _, id := range []uuid.UUID{theirIsland, theirHexproof} {
		if !shown[id.String()] {
			t.Errorf("view omitted the opponent's payable land %s", id)
		}
	}
	if shown[mine.String()] || shown[src.String()] {
		t.Errorf("view offered a permanent the viewer controls: %v", row.TapOthersOptions.Cards)
	}
	if len(row.TapOthersOptions.Cards) != 2 {
		t.Errorf("tap_others_options = %v, want the opponent's two tapped lands", row.TapOthersOptions.Cards)
	}

	// The flag reaches the wire under its documented name, and no colour
	// is published up front: which colours the one slot offers is decided
	// by WHICH land the activator untaps, so the slot asks its own
	// mana_pick.
	raw, err := json.Marshal(row)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if m["tap_others_untap"] != true {
		t.Errorf("wire JSON has no tap_others_untap: %s", raw)
	}
	if _, ok := m["color_options"]; ok {
		t.Errorf("wire JSON publishes color_options before the land is named: %s", raw)
	}

	// Springleaf Drum's tap form never carries the flag.
	drum := vehicleView(t, g, seatTapOthersManaSource(g, me.ID)).ManaAbilities
	if len(drum) != 1 || drum[0].TapOthersUntap {
		t.Errorf("the tap form carries tap_others_untap = %v", drum[0].TapOthersUntap)
	}
}
