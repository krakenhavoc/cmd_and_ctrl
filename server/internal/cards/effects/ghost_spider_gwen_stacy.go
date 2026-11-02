package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ghost-Spider, Gwen Stacy — {3}{R}{R} Legendary Creature — Spider Human
// Hero, 4/4 (Marvel's Spider-Man Eternal, #4460):
//
//	"Menace
//	 Whenever Ghost-Spider attacks, she deals X damage to defending player,
//	 where X is the number of attacking creatures."
//
// X is counted as the ability resolves (CR 608.2h), over every attacking
// creature of any controller. The defending player is the one behind the
// attack target the trigger's event recorded (b17DefendingPlayer), so it
// still names a player when she left combat first; the damage then comes
// from her last-known self.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1d95b5c0-d193-4a03-9954-23861ac1c762",
		Name:            "Ghost-Spider, Gwen Stacy",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Ghost-Spider — deal damage to defending player equal to the number of attacking creatures",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					defender := b17DefendingPlayer(g, item.Trigger.Event)
					if g.PlayerByIDForEffect(defender) == nil {
						return nil
					}
					x := len(b26AttackingCreatures(g))
					if x <= 0 {
						return nil
					}
					return DealDamage{Source: item.SourceCardID, Target: defender, Amount: x}.Apply(NewContext(g, item))
				}),
		},
	})
}
