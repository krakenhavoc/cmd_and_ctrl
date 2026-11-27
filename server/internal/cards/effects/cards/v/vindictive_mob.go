package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vindictive Mob — Creature — Human Berserker {4}{B}{B}, 5/5 (Ravnica: City of
// Guilds, S82 OLD batch 576, #5600):
//
//	"When this creature enters, sacrifice a creature.
//	 This creature can't be blocked by Saprolings."
//
// The trigger makes its controller sacrifice a creature of their choice as it
// resolves; the Mob itself may be the one. The block rule is River Darter's
// kind: a creature with the Saproling type (through the layered view, so a
// changeling is one) can't block the Mob (CR 509.1b).
//
// No simplification.
func init() {
	const label = "Vindictive Mob — sacrifice a creature"
	effects.Register(effects.Spec{
		OracleID:     "af55096c-af6b-479f-8257-144e596e89d8",
		Name:         "Vindictive Mob",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.Subtype("Saproling"), "Saprolings"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				g.PlayerSacrificesForEffect(item.SourceCardID, item.Controller, effects.SacrificeSpec("a creature", effects.Creature()), label)
				return nil
			}),
		},
	})
}
