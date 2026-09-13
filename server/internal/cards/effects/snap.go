package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Snap — Instant {1}{U} (EDHREC rank 273):
//
//	"Return target creature to its owner's hand. Untap up to two
//	 lands."
//
// A free bounce spell: two mana in, two lands back. The bounce is
// Unsummon's; the refund runs after it, in printed order.
//
// Sandbox simplification, declared: "untap up to two lands" prints
// no "target" and no "you control" — printed, it is a resolution-time
// choice among every land at the table. The engine has no
// resolution-time pick-a-permanent prompt for a spell, so this
// untaps the first two TAPPED lands the caster controls, in
// battlefield order. That is the outcome a player chooses in every
// real game, it is one the printed card allows, and it is never
// stronger — only less controllable (you cannot pick WHICH two of
// your tapped lands, nor untap an opponent's). With fewer than two
// tapped lands it untaps what there is.
func init() {
	Register(Spec{
		OracleID:     "ac914d98-221e-426c-8a50-342896b15f9e",
		Name:         "Snap",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"You can't choose which lands untap — it automatically untaps the first two tapped lands you control and can never untap an opponent's lands."},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			if err := (BounceToHand{Target: item.Targets[0].ID}).Apply(ctx); err != nil {
				return err
			}
			return b02UntapLandsYouControl(ctx, ctx.Controller(), 2)
		},
	})
}
