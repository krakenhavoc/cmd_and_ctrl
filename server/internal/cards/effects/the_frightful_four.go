package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Frightful Four — Legendary Creature — Villain {3}{B}, 4/4 (Marvel
// Super Heroes Commander, #4455):
//
//	"Menace
//	 Whenever an opponent casts their first noncreature spell each turn,
//	 that player loses life equal to that spell's mana value."
//
// Esper Sentinel's trigger condition (AnOpponentCastTheirFirstNoncreature-
// SpellThisTurn, read off the per-turn cast tally). The amount is the
// spell's mana value read as the trigger resolves (triggeringSpellManaValue,
// CR 202.3e: an announced X counts while the spell is on the stack; a spell
// that has since been countered keeps the value it had). The loss is life
// loss by the caster, not damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6d3469dd-d39e-463b-9b41-2f6215a41feb",
		Name:            "The Frightful Four",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, AnOpponentCastTheirFirstNoncreatureSpellThisTurn,
				"The Frightful Four — that player loses life equal to that spell's mana value",
				func(g *game.Game, item *game.StackItem) error {
					caster := item.Trigger.Event.Actor
					if p := g.PlayerByIDForEffect(caster); p == nil || p.Eliminated {
						return nil
					}
					n := triggeringSpellManaValue(g, item)
					if n <= 0 {
						return nil
					}
					return g.ChangePlayerLifeForEffect(item.SourceCardID, caster, -n)
				}),
		},
	})
}
