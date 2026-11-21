package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reaper's Scythe — Artifact — Equipment {2}{B}:
//
//	"Job select (When this Equipment enters, create a 1/1 colorless Hero
//	 creature token, then attach this to it.)
//	 At the beginning of your end step, put a soul counter on this
//	 Equipment for each player who lost life this turn.
//	 Equipped creature gets +1/+1 for each soul counter on this
//	 Equipment and is an Assassin in addition to its other types.
//	 Death Sickle — Equip {2}"
//
// Job select is living weapon with a Hero. The counters are counted as
// the end-step ability resolves: every seated player who lost life this
// turn (damage or loss, b18LifeLostThisTurn), yourself included. The
// bonus is read live from the Equipment's counters on every layer pass.
//
// No simplification.
func init() {
	const soul = "soul"
	effects.Register(effects.Spec{
		OracleID:     "b0f7c2a9-914a-48c1-871c-6f72946d61d8",
		Name:         "Reaper's Scythe",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttachedPer(1, 1, func(_ *game.Game, source *game.Card) int {
				return source.Counters[soul]
			}),
			effects.AttachedIsAlso("Assassin"),
		},
		Triggered: []game.TriggeredAbility{
			effects.JobSelect("Reaper's Scythe"),
			effects.AtYourEndStep("Reaper's Scythe — put a soul counter on it for each player who lost life this turn",
				func(g *game.Game, item *game.StackItem) error {
					n := 0
					for _, p := range g.Seats {
						if p != nil && !p.Eliminated && effects.B18LifeLostThisTurn(g, p.ID) > 0 {
							n++
						}
					}
					return effects.AddCounter{Target: item.SourceCardID, Kind: soul, N: n}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
