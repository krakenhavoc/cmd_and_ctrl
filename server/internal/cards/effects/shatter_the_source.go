package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shatter the Source — Instant {5}{R} (March of the Machine):
//
//	"Convoke. Choose one —
//	 • Shatter the Source deals 6 damage to target creature, planeswalker,
//	   or battle.
//	 • Destroy target artifact."
//
// Abrade's two modes with a wider first target and 6 damage. The creature,
// planeswalker or battle filter is Or(Creature(), Planeswalker(), Battle()).
func init() {
	Register(Spec{
		OracleID:     "d8b3bcf2-1be1-46ed-9531-ca4ce98e3b3c",
		Name:         "Shatter the Source",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Modes: ChooseOne(
			ModeWithPurpose(Mode("Shatter the Source deals 6 damage to target creature, planeswalker, or battle.",
				TargetPermanent("target creature, planeswalker, or battle", Or(Creature(), Planeswalker(), Battle()))),
				ForTargets(DamageToTarget(0, 6))),
			Mode("Destroy target artifact.",
				TargetPermanent("target artifact", Artifact())),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			switch {
			case ctx.HasMode(0):
				return DealDamage{Source: ctx.Source(), Target: target, Amount: 6}.Apply(ctx)
			case ctx.HasMode(1):
				return DestroyTarget{Target: target}.Apply(ctx)
			}
			return nil
		},
	})
}
