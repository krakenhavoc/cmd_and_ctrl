package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grim Feast — Enchantment {1}{B}{G} (Mirage, #5633):
//
//	"At the beginning of your upkeep, this enchantment deals 1 damage to you.
//	 Whenever a creature is put into an opponent's graveyard from the
//	 battlefield, you gain life equal to its toughness."
//
// The toughness is the creature's last-known toughness on the battlefield
// (CR 603.10a). A token counts: it is put into the graveyard for a moment
// (CR 111.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "331233ba-0ddc-4cb2-b3b7-1b782e03529f",
		Name:         "Grim Feast",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Grim Feast — this enchantment deals 1 damage to you", func(g *game.Game, item *game.StackItem) error {
				return effects.DealDamage{Source: item.SourceCardID, Target: item.Controller, Amount: 1}.Apply(effects.NewContext(g, item))
			}),
			effects.On(game.EventLTB, effects.OldB033CreatureToAnOpponentsGraveyard,
				"Grim Feast — you gain life equal to its toughness",
				effects.Do(effects.GainLifeEqualToSubject{Stat: effects.SubjectToughness})),
		},
	})
}
