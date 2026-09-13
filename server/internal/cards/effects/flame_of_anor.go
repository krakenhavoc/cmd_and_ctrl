package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flame of Anor — Instant {1}{U}{R} (EDHREC rank 1741):
//
//	"Choose one. If you control a Wizard as you cast this spell, you
//	 may choose two instead.
//	 • Target player draws two cards.
//	 • Destroy target artifact.
//	 • Flame of Anor deals 5 damage to target creature."
//
// The Wizard deck's charm. Three targeted options on a "choose one",
// each one primitive — the draw's target is a player, the other two
// are permanents — resolved in printed order.
//
// Sandbox simplification, declared: the Wizard bonus is not
// modelled, so the spell is always "choose one". All three options
// target, and a mode spec with Max > 1 may carry at most one
// targeted option (per-mode target slots are the Cryptic Command
// seam; Register panics on more). Without a Wizard the printed card
// IS this card; with one it is weaker than printed, never stronger.
func init() {
	Register(Spec{
		OracleID:     "ecd49d85-9c8c-4cc9-9a83-072ecd433677",
		Name:         "Flame of Anor",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"You always choose one mode — controlling a Wizard doesn't let you choose two."},
		Modes: ChooseOne(
			Mode("Target player draws two cards.", TargetPlayer("target player")),
			Mode("Destroy target artifact.", TargetPermanent("target artifact", Artifact())),
			Mode("Flame of Anor deals 5 damage to target creature.", TargetCreature("target creature")),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				switch {
				case ctx.HasMode(0) && t.Kind == game.TargetPlayer:
					if err := (DrawCards{Player: t.ID, N: 2}).Apply(ctx); err != nil {
						return err
					}
				case ctx.HasMode(1) && t.Kind == game.TargetCard:
					if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
				case ctx.HasMode(2) && t.Kind == game.TargetCard:
					if err := (DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 5}).Apply(ctx); err != nil {
						return err
					}
				}
			}
			return nil
		},
	})
}
