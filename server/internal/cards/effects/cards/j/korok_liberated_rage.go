package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Korok, Liberated Rage — Legendary Artifact Creature — Klingon Borg
// {3}{W}, 2/2 (Star Trek Commander, #4974):
//
//	"Battlefront — Whenever Korok attacks, creatures you control get
//	 +X/+X until end of turn, where X is the number of opponents you
//	 attacked this turn."
//
// Battlefront is an ability word. X is counted as the ability resolves
// (OpponentsYouAttackedThisTurn): each opponent still in the game that a
// creature of yours was declared attacking this turn counts once,
// whichever combat it was; an attack on a planeswalker or a battle is
// not an attack on its controller (CR 508.3e's distinction). The set of
// creatures is locked as it resolves (CR 611.2c), so a creature put onto
// the battlefield afterwards gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2f0bad22-a3f5-4b2b-96bf-ff1473e2262d",
		Name:         "Korok, Liberated Rage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Korok — battlefront: creatures you control get +X/+X, where X is the number of opponents you attacked this turn",
				func(g *game.Game, item *game.StackItem) error {
					x := effects.OpponentsYouAttackedThisTurn(g, item.Controller)
					return effects.BoostUntilEOT{
						Match:     effects.And(effects.Creature(), effects.YouControl()),
						Power:     x,
						Toughness: x,
						Label:     "Korok, Liberated Rage — battlefront",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
