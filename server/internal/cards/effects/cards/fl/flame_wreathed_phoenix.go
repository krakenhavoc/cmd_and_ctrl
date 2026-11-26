package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Flame-Wreathed Phoenix — {2}{R}{R} Creature — Phoenix 3/3 (Born of the Gods):
//
//	"Flying
//	 Tribute 2 (As this creature enters, an opponent of your choice may put two
//	 +1/+1 counters on it.)
//	 When this creature enters, if tribute wasn't paid, it gains haste and "When this
//	 creature dies, return it to its owner's hand.""
//
// No stated duration, so the grant lasts as long as the creature stays on the
// battlefield (CR 611.2a); a creature that left and came back is a new object
// without it (CR 400.7).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "24b9204f-ecc8-472e-9296-f999adf1ab1b",
		Name:            "Flame-Wreathed Phoenix",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Grants: []effects.AbilityGrant{{
			Key: flameWreathedPhoenixGrant,
			Triggered: []game.TriggeredAbility{
				effects.WhenThisDies("Flame-Wreathed Phoenix — return it to its owner's hand", effects.ReturnSelfFromGraveyardToHand),
			},
			Text: "When this creature dies, return it to its owner's hand.",
		}},
		Replacements: effects.Tribute("Flame-Wreathed Phoenix", 2),
		Triggered: []game.TriggeredAbility{
			effects.WhenEntersIfTributeWasntPaid("Flame-Wreathed Phoenix — it gains haste and \"When this creature dies, return it to its owner's hand\"",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.B09SourceStillOnBattlefield(g, item) {
						return nil
					}
					return effects.GrantAbilitiesFor{
						Target:   item.SourceCardID,
						Keys:     []string{flameWreathedPhoenixGrant},
						Also:     []game.Mod{game.AddKeywordsMod("haste")},
						Duration: game.IndefiniteDuration(),
						Label:    "Flame-Wreathed Phoenix — gains haste and a death trigger",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// flameWreathedPhoenixGrant is the granted death trigger's key. An on-disk identity:
// never renamed, never reused.
const flameWreathedPhoenixGrant = "flame-wreathed-phoenix/dies-return"
