package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Agrus Kos, Spirit of Justice — Legendary Creature — Spirit Detective
// {2}{R}{W}:
//
//	"Double strike, vigilance
//	 Whenever Agrus Kos enters or attacks, choose up to one target
//	 creature. If it's suspected, exile it. Otherwise, suspect it. (A
//	 suspected creature has menace and can't block.)"
//
// One ability with two trigger conditions (WhenThisEntersOrAttacks, as
// Sun Titan). The branch is made on resolution, against the creature as
// it stands then: a target suspected in response is exiled, not
// suspected a second time.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "11b7ee78-fca6-4b34-93e4-e60492583d50",
		Name:            "Agrus Kos, Spirit of Justice",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike", "vigilance"},
		Triggered: []game.TriggeredAbility{effects.Targeting(
			effects.WhenThisEntersOrAttacks("Agrus Kos — exile target creature if it's suspected, otherwise suspect it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if g.IsSuspected(t.ID) {
							if err := (effects.ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
								return err
							}
							continue
						}
						if err := (effects.Suspect{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
			effects.UpToOneTargetCreature("up to one target creature"),
		)},
	})
}
