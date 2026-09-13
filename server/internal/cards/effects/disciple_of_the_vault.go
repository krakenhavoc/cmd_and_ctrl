package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Disciple of the Vault — Creature — Human Cleric {B}, 1/1 (EDHREC
// rank 2452):
//
//	"Whenever an artifact is put into a graveyard from the
//	 battlefield, you may have target opponent lose 1 life."
//
// The artifact-aristocrats one-drop. One optional, targeted trigger
// per artifact that leaves the battlefield for a graveyard — anyone's
// artifact, a Treasure cracked for mana, a Clue sacrificed to draw,
// an artifact creature that died in combat; a bounce or an exile
// does not count. "You may" is the S19 yes/no prompt, "target
// opponent" the S20 pick, and the loss is life loss rather than
// damage, so nothing prevents it. A board wipe that takes five
// artifacts fires it five times, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c8625113-0ce4-4454-83a1-25c31b8bfb9a",
		Name:         "Disciple of the Vault",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b23ArtifactPutIntoGraveyardFromBattlefield(ev, g)
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Disciple of the Vault: have target opponent lose 1 life?"},
			Targets:        TargetPlayer("target opponent", Opponent()),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Disciple of the Vault — target opponent loses 1 life",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							if t.Kind != game.TargetPlayer {
								continue
							}
							if err := g.ChangePlayerLifeForEffect(ctx.Source(), t.ID, -1); err != nil {
								return err
							}
						}
						return nil
					})
			},
		}},
	})
}
