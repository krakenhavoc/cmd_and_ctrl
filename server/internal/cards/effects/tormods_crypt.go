package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "1573f7f9-672c-421a-b1ac-3d0d8aea59ca",
		Name:         "Tormod's Crypt",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice Tormod's Crypt: Exile target player's graveyard",
			Cost:    Plus(TapCost(), SacrificeThis()),
			Targets: TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
					return nil
				}
				return exileGraveyardForEffect(g, item, item.Targets[0].ID)
			},
		}},
	})
}
