package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tendrils of Agony — Sorcery {2}{B}{B}:
//
//	"Target player loses 2 life and you gain 2 life.
//	 Storm (When you cast this spell, copy it for each spell cast
//	 before it this turn. You may choose new targets for the copies.)"
//
// The storm kill. Ten copies is twenty life off one player and twenty
// onto you, which is why the count has to be exactly right in both
// directions and why the countered-spell reading matters: a storm turn
// is full of cantrips that were cast and are long gone from every
// zone, and every one of them counts (CR 702.40a, "each other spell
// that was cast").
//
// "Loses 2 life and you gain 2 life" is TWO separate life changes
// rather than a drain, which is what the printed text says and what
// anything watching life gain — or a "your life total can't change"
// effect on either side (ADR 0085) — reads. It is not damage, so
// nothing that prevents or redirects damage touches it.
//
// "You" is the copy's controller, not the original's, which falls out
// of the copies being controlled by the player who created them
// (CR 707.10b) and of the trigger's controller being the caster.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "78db3190-59bc-4033-95bc-32d8ed872b6a",
		Name:         "Tendrils of Agony",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		Triggered:    []game.TriggeredAbility{Storm()},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			if err := (GainLife{Player: item.Targets[0].ID, Amount: -2}).Apply(ctx); err != nil {
				return err
			}
			return GainLife{Player: item.Controller, Amount: 2}.Apply(ctx)
		},
	})
}
