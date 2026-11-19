package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "ff135dc7-31c8-41d6-a5d5-626217844d09",
		Name:         "Undercover Butler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, preacherAttackedTheLifeLeader,
				"Undercover Butler — it can't be blocked this turn",
				func(g *game.Game, item *game.StackItem) error {
					return RestrictUntilEOT{
						Target:       item.SourceCardID,
						Restrictions: game.CantBeBlocked,
						Label:        "Undercover Butler — can't be blocked",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
