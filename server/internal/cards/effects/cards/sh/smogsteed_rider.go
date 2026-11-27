package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Smogsteed Rider — Creature — Human Wizard {2}{B}{B}, 2/2 (Guildpact):
//
//	"Whenever this creature attacks, each other attacking creature gains
//	 fear until end of turn. (They can't be blocked except by artifact
//	 creatures and/or black creatures.)"
//
// The set of attackers is locked as the ability resolves (CR 611.2c), the
// Rider itself left out, whoever controls the others (in a multiplayer
// attack an opponent's attacker is "other" too). Fear is an engine keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a23cb875-2ef8-4457-81eb-e66b67a3c477",
		Name:         "Smogsteed Rider",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Smogsteed Rider — each other attacking creature gains fear until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					self := item.SourceCardID
					return effects.GrantKeywordUntilEOT{
						Match:    effects.And(effects.AttackingCreature(), effects.NotSelf(self)),
						Keywords: []string{"fear"},
						Label:    "Smogsteed Rider — fear",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
