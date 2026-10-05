package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Puresteel Paladin — Creature — Human Knight {W}{W}, 2/2:
//
//	"Whenever an Equipment you control enters, you may draw a card.
//	 Metalcraft — Equipment you control have equip {0} as long as you
//	 control three or more artifacts."
//
// The draw is an ordinary "you may" entry trigger over any Equipment
// that enters under your control, the Paladin's own entry excluded by
// type (it is not an Equipment). The metalcraft clause rewrites the
// equip cost of OTHER permanents, and no seam lets one permanent set
// another's equip cost to a fixed value (Leonin Shikari only
// discounts), so it is left out: Equipment keeps its printed equip
// cost, which is weaker than printed.
func init() {
	Register(Spec{
		OracleID:     "74a62c7b-4753-4af2-b7a1-9a4ae8988801",
		Name:         "Puresteel Paladin",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The metalcraft clause isn't implemented — your Equipment never gets equip {0}, whatever you control."},
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventETB, puresteelPaladinEquipmentEntered,
				"Puresteel Paladin — draw a card", Do(DrawCards{N: 1})),
				"Draw a card?"),
		},
	})
}

func puresteelPaladinEquipmentEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	return ok && hasSubtype(c, "Equipment")
}
