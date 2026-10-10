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
// type (it is not an Equipment).
//
// The metalcraft clause GRANTS an equip ability (#2562, ADR 0093
// amendment 2026-10-10); it does not rewrite the Equipment's own. Each
// Equipment you control has its printed equip and equip {0}, and either
// may be activated (CR 702.6d). The granted row is the Equipment's own
// activated ability (CR 702.6a), so it attaches that Equipment, at
// sorcery speed. The grant is a layer-6 static whose predicate counts
// your artifacts as the layer pass reaches layer 6, so it switches on
// the moment the third artifact arrives and off when one leaves.
//
// No simplification.
const puresteelPaladinEquipGrant = "puresteel-paladin/equip-zero"

func init() {
	Register(Spec{
		OracleID:     "74a62c7b-4753-4af2-b7a1-9a4ae8988801",
		Name:         "Puresteel Paladin",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{Key: puresteelPaladinEquipGrant, Activated: []ActivatedAbility{EquipAbility("{0}")}, Text: "Equip {0}"},
		},
		Static: []game.StaticAbility{
			GrantAbilities(puresteelPaladinMetalcraft, puresteelPaladinEquipGrant),
		},
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

// puresteelPaladinMetalcraft is "Equipment you control … as long as you
// control three or more artifacts".
func puresteelPaladinMetalcraft(target *game.Card, g *game.Game, source *game.Card) bool {
	return equipmentYouControl(target, g, source) &&
		permanentsYouControl(g, source, func(c *game.Card) bool { return c.IsArtifact() }) >= 3
}
