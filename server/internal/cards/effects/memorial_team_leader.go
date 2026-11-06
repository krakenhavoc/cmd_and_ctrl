package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Memorial Team Leader — Creature — Kavu Soldier {3}{R}, 4/3 (Edge of
// Eternities, #4461):
//
//	"During your turn, other creatures you control get +1/+0.
//	 Warp {1}{R} (You may cast this card from your hand for its warp cost.
//	 Exile this creature at the beginning of the next end step, then you
//	 may cast it from exile on a later turn.)"
//
// A layer 7c anthem (b16Anthem, CR 613.4c) over the other creatures you
// control, gated on the Leader's controller being the active player, the
// gate Alpharael, Dreaming Acolyte's deathtouch reads. A static ability
// applies to whatever its text says at each moment (CR 611.3a), so a
// creature that enters later in your turn gets the bonus too, and the
// bonus is gone on every other player's turn. Warp is the shared
// alternative cost: a warped Leader pumps the team for the turn it
// attacks and leaves at the end step.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "c71d3775-0c71-4e37-979d-35e319f357f0",
		Name:             "Memorial Team Leader",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Warp("{1}{R}")},
		Static: []game.StaticAbility{
			b16Anthem(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return OtherCreaturesYouControl(target, g, source) && isActivePlayer(g, source.Controller)
			}, 1, 0),
		},
	})
}
