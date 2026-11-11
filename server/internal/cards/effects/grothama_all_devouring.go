package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "34fd52c1-3d81-4005-b826-eeefbd7fd874",
		Name:         "Grothama, All-Devouring",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: grothamaFightGrant,
			Triggered: []game.TriggeredAbility{
				Optional(WheneverThisAttacks("Grothama — you may have this creature fight Grothama, All-Devouring",
					cmdr2S52GrothamaFight), "Have this creature fight Grothama, All-Devouring?"),
			},
			Text: "Whenever this creature attacks, you may have it fight Grothama, All-Devouring.",
		}},
		Static: []game.StaticAbility{
			GrantAbilities(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.InstanceID != source.InstanceID
			}, grothamaFightGrant),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisLeaves("Grothama — each player draws cards equal to the damage dealt to it this turn by sources they controlled", cmdr2S52GrothamaDraw),
		},
	})
}
