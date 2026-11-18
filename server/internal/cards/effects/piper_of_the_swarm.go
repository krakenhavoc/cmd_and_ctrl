package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Piper of the Swarm — {1}{B} Creature — Human Warlock 1/3:
//
//	"Rats you control have menace.
//	 {1}{B}, {T}: Create a 1/1 black Rat creature token.
//	 {2}{B}{B}, {T}, Sacrifice three Rats: Gain control of target
//	 creature."
//
// A layer-6 keyword grant to the Rats its controller controls (the
// Piper itself is a Human Warlock, so it is not one), a token maker
// and Rat King, Verminister's "Sacrifice three Rats" cost. The theft
// has no duration, so it lasts until the game ends or something else
// changes control (CR 611.2a). Because the three Rats are a cost, they
// are gone before the ability resolves: it cannot be answered by
// killing the Rats in response, and it fizzles, with the Rats still
// spent, if its target is gone (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a6c5c5ea-0bd0-4d5a-bd3f-6217f14c9939",
		Name:         "Piper of the Swarm",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && target.HasSubtype(source.TextSubtype("Rat"))
			}, "menace"),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "{1}{B}, {T}: Create a 1/1 black Rat creature token.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{1}{B}"), TapCost()),
				Effect:  Do(CreateToken{Template: TokenCard("1/1 black Rat"), N: 1}),
			},
			{
				Label:   "{2}{B}{B}, {T}, Sacrifice three Rats: Gain control of target creature.",
				Cost:    Plus(ManaCost("{2}{B}{B}"), TapCost(), SacrificeN(3, "three Rats", Subtype("Rat"))),
				Targets: TargetCreature("target creature"),
				Purpose: game.Purpose{Answers: game.AnswerRemove | game.AnswerSacOutlet},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						return GainControl{
							Target:   t.ID,
							Duration: game.IndefiniteDuration(),
							Label:    "Piper of the Swarm — gain control",
						}.Apply(ctx)
					}
					return nil
				},
			},
		},
	})
}
