package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thundering Djinn — Creature — Djinn {3}{U}{R}, 3/4 (Modern Horizons):
//
//	"Flying
//	 Whenever this creature attacks, it deals damage to any target equal to the
//	 number of cards you've drawn this turn."
//
// Flying is a printed keyword. The target is chosen as the trigger goes on the
// stack; the amount is your tally of cards drawn this turn read as it
// resolves (CR 608.2c), every draw counting however it came, as Fractal
// Anomaly's X does. The damage is the Djinn's, as it last existed if it has
// gone (CR 608.2h), so a lifelink or deathtouch Djinn keeps those on it. With
// none drawn nothing is dealt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "da0042a7-89a1-4500-b689-b0ea7cab5ed8",
		Name:            "Thundering Djinn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisAttacks("Thundering Djinn — it deals damage to any target equal to the number of cards you've drawn this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					n := g.TurnTallyFor(ctx.Controller()).CardsDrawn
					ref, ok := ctx.SourceRef()
					legal := ctx.LegalTargets()
					if n <= 0 || !ok || len(legal) == 0 {
						return nil
					}
					return effects.DealDamage{SourceObject: &ref, Target: legal[0].ID, Amount: n}.Apply(ctx)
				}),
				effects.TargetAny()),
		},
	})
}
