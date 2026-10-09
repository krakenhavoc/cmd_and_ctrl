package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "11b7ee78-fca6-4b34-93e4-e60492583d50",
		Name:            "Agrus Kos, Spirit of Justice",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike", "vigilance"},
		Triggered: []game.TriggeredAbility{Targeting(
			WhenThisEntersOrAttacks("Agrus Kos — exile target creature if it's suspected, otherwise suspect it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if g.IsSuspected(t.ID) {
							if err := (ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
								return err
							}
							continue
						}
						if err := (Suspect{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
			UpToOneTargetCreature("up to one target creature"),
		)},
	})
}
