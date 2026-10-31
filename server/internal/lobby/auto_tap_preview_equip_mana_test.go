package lobby

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// auto_tap_preview_equip_mana_test.go — #4600. The ?ability= preview
// builds the spend context the activation pays under, equip bit
// included, so mana restricted to "Equipment spells or equip abilities"
// (Ronin, Shadow Stalker) covers an equip ability's cost in the preview
// and does not cover another ability of the same Equipment. The engine
// half is game/equip_mana_test.go; the fixture is previewFixture.

const shadowspearPreviewOracle = "8b27326f-e7b8-4a4d-b589-df459246d19a"

func TestAutoTapPreviewAbilityReadsTheEquipBitOfTheRow(t *testing.T) {
	f := newPreviewFixture(t)
	f.mainPhaseForAlice()
	spear := f.spawn(f.alice, game.ZoneBattlefield, game.Card{
		Name: "Shadowspear", TypeLine: "Legendary Artifact — Equipment", OracleID: shadowspearPreviewOracle,
	}, 1)[0]
	golem := f.spawn(f.alice, game.ZoneBattlefield, game.Card{
		Name: "Golem", TypeLine: "Artifact Creature — Golem", Power: 2, Toughness: 2,
	}, 1)[0]
	restriction := game.ManaRestrictAnyOf(
		[]string{game.ManaRestrictCast, game.ManaRestrictSubtype("Equipment")},
		[]string{game.ManaRestrictEquip},
	)
	seat := f.game().PlayerByIDForEffect(f.alice)
	for range 2 {
		seat.ManaPool.AddMana(game.ManaToken{Color: "R", Restrictions: []string{restriction}})
	}

	// Row 1 is the equip {2}: the floating pair covers it, tapping nothing.
	equip := f.previewWith(spear, "&ability=1&targets=card:"+golem.String())
	if !equip.OK || len(equip.Plan) != 0 {
		t.Errorf("equip {2} over equip-only mana: ok=%v plan=%v missing=%v, want ok with nothing to tap",
			equip.OK, equip.Plan, equip.Missing)
	}
	// Row 0 is "{1}: ...", not an equip ability: the same mana is no use.
	other := f.previewAbility(spear, "")
	if other.OK || len(other.Missing) != 1 {
		t.Errorf("the Equipment's other ability over equip-only mana: ok=%v missing=%v, want one symbol missing",
			other.OK, other.Missing)
	}
}
