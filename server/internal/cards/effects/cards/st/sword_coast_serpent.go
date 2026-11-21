package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sword Coast Serpent // Capsizing Wave — an Adventure creature (Commander
// Legends: Battle for Baldur's Gate, #5948):
//
//	Sword Coast Serpent {5}{U}{U}, Creature — Serpent Dragon, 6/6
//	  "This creature can't be blocked as long as you've cast a noncreature
//	   spell this turn."
//	Capsizing Wave {1}{U}, Instant — Adventure
//	  "Return target creature to its owner's hand."
//
// Bast's conditional restriction: a layer 6 static on this creature alone,
// read continuously off the controller's cast tally (the tally is bumped as a
// spell becomes cast, CR 601.2i, so the creature is unblockable from the
// moment the spell is cast; a creature spell never counts). Capsizing Wave is
// Petty Theft's bounce over any creature.
//
// No simplification.
func init() {
	const oracleID = "d60e14b4-f750-40cc-85b5-17c87ae8a584"
	effects.Register(effects.Spec{
		OracleID:     oracleID,
		Name:         "Sword Coast Serpent",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label: "This creature can't be blocked as long as you've cast a noncreature spell this turn.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.SelfOnly(target, g, source) && g.CastTallyFor(source.Controller).Noncreature > 0
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Restrictions |= game.CantBeBlocked
			},
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     oracleID + "#1",
		Name:         "Capsizing Wave",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.BounceTheTarget,
	})
}
