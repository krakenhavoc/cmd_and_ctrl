package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Undercover Butler — Creature — Human Rogue {2}{U/B}, 2/3 (Ravnica: Clue
// Edition, S82):
//
//	"Whenever this creature attacks the player with the most life or tied
//	 for most life, it can't be blocked this turn."
//
// Preacher of the Schism's "attacks the player with the most life" trigger
// (an attack on a planeswalker or battle is not an attack on a player),
// read once as the attack is declared; the restriction is Artful Dodge's,
// laid on the Butler until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ff135dc7-31c8-41d6-a5d5-626217844d09",
		Name:         "Undercover Butler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.PreacherAttackedTheLifeLeader,
				"Undercover Butler — it can't be blocked this turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.RestrictUntilEOT{
						Target:       item.SourceCardID,
						Restrictions: game.CantBeBlocked,
						Label:        "Undercover Butler — can't be blocked",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
