package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stone Drake — Creature — Drake {3}{U}{B}, 4/4 (Mystery Booster 2):
//
//	"Flying
//	 When Stone Drake enters the battlefield, choose one —
//	 • Distract — Tap target land. Draw a card.
//	 • Enthrall — Target player discards a card for each spell you've cast
//	   this turn."
//
// A modal ETB trigger: the mode and its target are chosen as it goes on the
// stack. Enthrall counts the spells its controller cast this turn as it
// resolves (CR 608.2h); the discard is the target player's own choice.
//
// No simplification.
func init() {
	drake := WhenThisEnters("Stone Drake — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	drake.Modes = ChooseOne(
		ModeDoing("Distract — Tap target land. Draw a card.",
			TargetPermanent("target land", Land()),
			func(item *game.StackItem, ctx *Context, occ int) error {
				if t, ok := ModeTarget(ctx, occ); ok {
					if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
				}
				return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
			}),
		ModeDoing("Enthrall — Target player discards a card for each spell you've cast this turn.",
			TargetPlayer("target player"),
			func(item *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetPlayer {
					return nil
				}
				n := ctx.Game.SpellsCastThisTurn[item.Controller].Total
				if n <= 0 {
					return nil
				}
				ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: t.ID, Source: item.SourceCardID, N: n})
				return nil
			}),
	)
	Register(Spec{
		OracleID:        "7373b50b-a707-40fb-967a-550c31e6ea59",
		Name:            "Stone Drake",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{drake},
	})
}
