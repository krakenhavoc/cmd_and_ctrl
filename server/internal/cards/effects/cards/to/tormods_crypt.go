package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tormod's Crypt — Artifact {0} (EDHREC rank 1831):
//
//	"{T}, Sacrifice this artifact: Exile target player's graveyard."
//
// The free graveyard hoser. One activated ability, two cost
// components (tap and sacrifice itself, merged by Plus), one targeted
// player, and the whole graveyard leaves as Bojuka Bog's body does —
// the IDs snapshotted before the first exile. A target player with
// an empty graveyard is still a legal target, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1573f7f9-672c-421a-b1ac-3d0d8aea59ca",
		Name:         "Tormod's Crypt",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice Tormod's Crypt: Exile target player's graveyard",
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
					return nil
				}
				return effects.ExileGraveyardForEffect(g, item, item.Targets[0].ID)
			},
		}},
	})
}
