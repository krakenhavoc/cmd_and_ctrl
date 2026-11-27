package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Selesnya Charm — Instant {G}{W} (Commander 2013, S82 older sets):
//
//	"Choose one —
//	 • Target creature gets +2/+2 and gains trample until end of turn.
//	 • Exile target creature with power 5 or greater.
//	 • Create a 2/2 white Knight creature token with vigilance."
//
// The exile bullet's power restriction is part of the target (CR 115.1),
// judged at announce and again as the spell resolves (CR 608.2b), read
// through the layered view.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a1a49639-bcc4-4522-8862-c9fb13d40880",
		Name:         "Selesnya Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.Mode("Target creature gets +2/+2 and gains trample until end of turn.",
				effects.TargetCreature("target creature")), effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2, "trample"))),
			effects.ModeWithPurpose(effects.Mode("Exile target creature with power 5 or greater.",
				effects.TargetCreature("target creature with power 5 or greater", effects.PowerGE(5))), game.Purpose{Answers: game.AnswerRemove}),
			effects.ModeWithPurpose(effects.Mode("Create a 2/2 white Knight creature token with vigilance."),
				game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerMakesBlocker}),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			switch {
			case ctx.HasMode(0):
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				id := item.Targets[0].ID
				if err := (effects.BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Selesnya Charm — +2/+2 until end of turn"}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: "Selesnya Charm — trample until end of turn"}.Apply(ctx)
			case ctx.HasMode(1):
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				return effects.ExileTarget{Target: item.Targets[0].ID}.Apply(ctx)
			case ctx.HasMode(2):
				return effects.CreateToken{Template: effects.TokenCard("2/2 white Knight with vigilance"), N: 1}.Apply(ctx)
			}
			return nil
		},
	})
}
