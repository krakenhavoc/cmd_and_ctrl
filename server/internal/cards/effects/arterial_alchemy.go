package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	pump := PumpAttached(2, 0)
	pump.Label = "Equipped creature gets +2/+0."
	Register(Spec{
		OracleID:     "dfc982a2-9ae8-4206-a2d7-4dcfcde02451",
		Name:         "Arterial Alchemy",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{Key: arterialAlchemyPumpGrant, Static: []game.StaticAbility{pump}, Text: "Equipped creature gets +2/+0."},
			{Key: arterialAlchemyEquipGrant, Activated: []ActivatedAbility{EquipAbility("{2}")}, Text: "Equip {2}"},
		},
		Static: []game.StaticAbility{
			AreAlsoEquipment("Blood tokens you control are Equipment in addition to their other types.", vowB05BloodTokensYouControl),
			Labelled(GrantAbilities(vowB05BloodTokensYouControl, arterialAlchemyPumpGrant, arterialAlchemyEquipGrant),
				"Blood tokens you control have \"Equipped creature gets +2/+0\" and \"Equip {2}\"."),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Arterial Alchemy — create a Blood token for each opponent you have",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return CreateToken{Controller: item.Controller, Template: BloodToken(), N: len(ctx.Opponents())}.Apply(ctx)
				}),
		},
	})
}

// vowB05BloodTokensYouControl is "Blood tokens you control".
func vowB05BloodTokensYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.Controller == source.Controller && IsToken(*target) && target.HasSubtype("Blood")
}
