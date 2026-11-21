package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Battle Menu — Instant {1}{W} (Final Fantasy, #4462):
//
//	"Choose one —
//	 • Attack — Create a 2/2 white Knight creature token.
//	 • Ability — Target creature gets +0/+4 until end of turn.
//	 • Magic — Destroy target creature with power 4 or greater.
//	 • Item — You gain 4 life."
//
// A plain choose-one; each targeted mode's clause is judged at announce
// and again as the spell resolves (CR 608.2b). "Power 4 or greater" is
// the creature's power then, after every pump and counter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8af57b80-bbc6-47ce-83f6-34ae8e2f5bd5",
		Name:         "Battle Menu",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.Mode("Attack — Create a 2/2 white Knight creature token."),
			effects.ModeWithPurpose(effects.Mode("Ability — Target creature gets +0/+4 until end of turn.", effects.TargetCreature("target creature")),
				effects.ForTargets(effects.BenefitUntilEOT(0, 0, 4))),
			effects.Mode("Magic — Destroy target creature with power 4 or greater.",
				effects.TargetCreature("target creature with power 4 or greater", effects.PowerGE(4))),
			effects.Mode("Item — You gain 4 life."),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			switch {
			case ctx.HasMode(0):
				return effects.CreateToken{Template: effects.TokenCard("2/2 white Knight"), N: 1}.Apply(ctx)
			case ctx.HasMode(1):
				if t, ok := battleMenuTarget(ctx); ok {
					return effects.BoostUntilEOT{Target: t, Toughness: 4, Label: "Battle Menu — +0/+4"}.Apply(ctx)
				}
			case ctx.HasMode(2):
				if t, ok := battleMenuTarget(ctx); ok {
					return effects.DestroyTarget{Target: t}.Apply(ctx)
				}
			case ctx.HasMode(3):
				return effects.GainLife{Amount: 4}.Apply(ctx)
			}
			return nil
		},
	})
}

// battleMenuTarget is the chosen mode's creature (the one occurrence's
// target group), if it is still legal.
func battleMenuTarget(ctx *effects.Context) (uuid.UUID, bool) {
	for _, t := range ctx.ModeTargets(0) {
		if t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
			return t.ID, true
		}
	}
	return uuid.Nil, false
}
