package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const grothamaFightGrant = "grothama/fight-on-attack"

// Grothama, All-Devouring — Legendary Creature — Wurm {3}{G}{G}, 10/8:
// "Other creatures have 'Whenever this creature attacks, you may have it
// fight Grothama, All-Devouring.' When Grothama leaves the battlefield, each
// player draws cards equal to the amount of damage dealt to Grothama this
// turn by sources they controlled."
//
// Every other creature, whoever controls it, carries the granted attack
// trigger; it names Grothama through the grant's source (StackItem.GrantedBy).
// The damage is read from the turn's damage record, which carries each
// event's amount and the source's controller.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "34fd52c1-3d81-4005-b826-eeefbd7fd874",
		Name:         "Grothama, All-Devouring",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: grothamaFightGrant,
			Triggered: []game.TriggeredAbility{
				effects.Optional(effects.WheneverThisAttacks("Grothama — you may have this creature fight Grothama, All-Devouring",
					effects.Cmdr2S52GrothamaFight), "Have this creature fight Grothama, All-Devouring?"),
			},
			Text: "Whenever this creature attacks, you may have it fight Grothama, All-Devouring.",
		}},
		Static: []game.StaticAbility{
			effects.GrantAbilities(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.InstanceID != source.InstanceID
			}, grothamaFightGrant),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisLeaves("Grothama — each player draws cards equal to the damage dealt to it this turn by sources they controlled", effects.Cmdr2S52GrothamaDraw),
		},
	})
}
