package effects

import (
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
	Register(Spec{
		OracleID:     "b0f7c2a9-914a-48c1-871c-6f72946d61d8",
		Name:         "Reaper's Scythe",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttachedPer(1, 1, func(_ *game.Game, source *game.Card) int {
				return source.Counters[soul]
			}),
			AttachedIsAlso("Assassin"),
		},
		Triggered: []game.TriggeredAbility{
			JobSelect("Reaper's Scythe"),
			AtYourEndStep("Reaper's Scythe — put a soul counter on it for each player who lost life this turn",
				func(g *game.Game, item *game.StackItem) error {
					n := 0
					for _, p := range g.Seats {
						if p != nil && !p.Eliminated && b18LifeLostThisTurn(g, p.ID) > 0 {
							n++
						}
					}
					return AddCounter{Target: item.SourceCardID, Kind: soul, N: n}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}
