package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "0fcd476f-4db8-4293-9388-1678a0043c9e",
		Name:         "Grinding Station",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice an artifact: Target player mills three cards",
			Cost:    Plus(TapCost(), b10SacrificeAnArtifact()),
			Targets: TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetPlayer {
						continue
					}
					return MillCards{Player: t.ID, N: 3}.Apply(ctx)
				}
				return nil
			},
		}},
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventETB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b20ArtifactEntered(ev, g)
			}, "Grinding Station — untap", func(g *game.Game, item *game.StackItem) error {
				if !b15OnBattlefield(g, item.SourceCardID) {
					return nil
				}
				return UntapTarget{Target: item.SourceCardID}.Apply(NewContext(g, item))
			}), "Grinding Station — untap it?"),
		},
	})
}
