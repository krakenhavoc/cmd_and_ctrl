package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arterial Alchemy — Enchantment {2}{R} (Innistrad: Crimson Vow Commander,
// #5495):
//
//	"When this enchantment enters, create a Blood token for each opponent you
//	 have.
//	 Blood tokens you control are Equipment in addition to their other types
//	 and have "Equipped creature gets +2/+0" and equip {2}."
//
// The entry trigger counts the controller's opponents as it resolves
// (Chittering Witch's len(ctx.Opponents())). The rest is Gemcutter Buccaneer's
// shape over Blood tokens: a layer-4 "Equipment in addition to their other
// types" (CR 205.1b) and a layer-6 grant of a +2/+0 static and an equip {2}
// ability (CR 702.6a). A Blood token keeps its own {1}, {T}, discard,
// sacrifice ability. When the enchantment leaves, the tokens stop being
// Equipment and the state-based action unattaches them (CR 704.5n).
//
// No simplification.
const (
	arterialAlchemyPumpGrant  = "arterial-alchemy/blood-pump"
	arterialAlchemyEquipGrant = "arterial-alchemy/blood-equip"
)

func init() {
	pump := effects.PumpAttached(2, 0)
	pump.Label = "Equipped creature gets +2/+0."
	effects.Register(effects.Spec{
		OracleID:     "dfc982a2-9ae8-4206-a2d7-4dcfcde02451",
		Name:         "Arterial Alchemy",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{
			{Key: arterialAlchemyPumpGrant, Static: []game.StaticAbility{pump}, Text: "Equipped creature gets +2/+0."},
			{Key: arterialAlchemyEquipGrant, Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")}, Text: "Equip {2}"},
		},
		Static: []game.StaticAbility{
			effects.AreAlsoEquipment("Blood tokens you control are Equipment in addition to their other types.", vowB05BloodTokensYouControl),
			effects.Labelled(effects.GrantAbilities(vowB05BloodTokensYouControl, arterialAlchemyPumpGrant, arterialAlchemyEquipGrant),
				"Blood tokens you control have \"Equipped creature gets +2/+0\" and \"Equip {2}\"."),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Arterial Alchemy — create a Blood token for each opponent you have",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.CreateToken{Controller: item.Controller, Template: effects.BloodToken(), N: len(ctx.Opponents())}.Apply(ctx)
				}),
		},
	})
}

// vowB05BloodTokensYouControl is "Blood tokens you control".
func vowB05BloodTokensYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.Controller == source.Controller && effects.IsToken(*target) && target.HasSubtype("Blood")
}
