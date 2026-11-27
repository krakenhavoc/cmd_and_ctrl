package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kashi-Tribe Elite — Creature — Snake Warrior {1}{G}{G}, 2/3 (Saviors of
// Kamigawa, S82 OLD b583):
//
//	"Legendary Snakes you control have shroud.
//	 Whenever this creature deals combat damage to a creature, tap that
//	 creature and it doesn't untap during its controller's next untap step."
//
// The grant is Indomitable Archangel's layer-6 keyword grant over the
// controller's legendary Snake creatures (the subtype is a word of the text,
// CR 612.2); the Elite is not legendary itself. The trigger is Matsu-Tribe
// Birdstalker's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dac9bbfc-62e9-42a9-8ad6-2530ba47b990",
		Name:         "Kashi-Tribe Elite",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.Labelled(effects.KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller &&
					target.IsLegendary() && target.HasSubtype(source.TextSubtype("Snake"))
			}, "shroud"), "Legendary Snakes you control have shroud."),
		},
		Triggered: []game.TriggeredAbility{
			effects.B583WhenThisDealsCombatDamageToACreatureTapFreeze("Kashi-Tribe Elite — tap the creature it dealt combat damage to; it doesn't untap during its controller's next untap step"),
		},
	})
}
