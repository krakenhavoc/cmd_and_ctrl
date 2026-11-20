package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grinding Station — Artifact {2} (EDHREC rank 2170):
//
//	"{T}, Sacrifice an artifact: Target player mills three cards.
//	 Whenever an artifact enters, you may untap this artifact."
//
// The artifact-combo mill outlet. The activated ability is a tap
// plus "sacrifice an artifact" (b10SacrificeAnArtifact — the Station
// itself qualifies, as printed) targeting a player; the trigger
// watches EVERY artifact entering, under anyone's control, the
// Station's own entry included, and asks before untapping. A Station
// untapped by the trigger can be activated again in the same
// resolution window, which is the printed loop.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0fcd476f-4db8-4293-9388-1678a0043c9e",
		Name:         "Grinding Station",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice an artifact: Target player mills three cards",
			Cost:    effects.Plus(effects.TapCost(), effects.B10SacrificeAnArtifact()),
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetPlayer {
						continue
					}
					return effects.MillCards{Player: t.ID, N: 3}.Apply(ctx)
				}
				return nil
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventETB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B20ArtifactEntered(ev, g)
			}, "Grinding Station — untap", func(g *game.Game, item *game.StackItem) error {
				if !effects.B15OnBattlefield(g, item.SourceCardID) {
					return nil
				}
				return effects.UntapTarget{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			}), "Grinding Station — untap it?"),
		},
	})
}
