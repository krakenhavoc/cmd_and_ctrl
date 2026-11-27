package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Energy Field — Enchantment {1}{U} (Urza's Saga, S82 OLD b721):
//
//	"Prevent all damage that would be dealt to you by sources you don't
//	 control.
//	 When a card is put into your graveyard from anywhere, sacrifice this
//	 enchantment."
//
// The shield is a static prevention of damage to you from a source another
// player controls, combat and noncombat, read off the source as the damage
// would be dealt (CR 609.7b). The trigger watches every route into a graveyard (a move,
// a mill, a discard, a countered spell) for a card you own, never a token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fbf7a7f4-9915-4a53-84b3-e55e9f487323",
		Name:         "Energy Field",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.PreventAllDamageDealtTo(effects.PreventionStatic{
				To:     effects.ToYou,
				Damage: effects.AnyDamage,
				From:   effects.S82OldB721FromASourceYouDontControl(),
				Label:  "Energy Field — prevent all damage that would be dealt to you by sources you don't control",
			}),
		},
		Triggered: []game.TriggeredAbility{
			effects.OnAny(effects.S82OldB721CardPutIntoYourGraveyardEvents, effects.S82OldB721CardPutIntoYourGraveyard,
				"Energy Field — sacrifice this enchantment", effects.SacrificeThisIfStillOnBattlefield),
		},
	})
}
