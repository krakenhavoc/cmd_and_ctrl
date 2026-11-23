package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rootwire Amalgam — Artifact Creature — Golem {5}, 5/5 (The Brothers'
// War, #5656): "Prototype {1}{G} — 2/3", "{3}{G}{G}, Sacrifice this
// creature: Create an X/X colorless Golem artifact creature token, where X
// is three times this creature's power. It gains haste until end of turn.
// Activate only as a sorcery."
//
// Prototype is the card's second face (CR 718, game/prototype.go), built
// by the deck importer from the inset; nothing here spells it out. The
// Amalgam is sacrificed as a cost, so X is its last-known power (CR
// 608.2h): 6 for a prototyped Amalgam, 15 for one cast normally, counters
// and continuous effects included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7644148f-3b26-447d-9f52-28c94c111b69",
		Name:         "Rootwire Amalgam",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Activated: []effects.ActivatedAbility{{
			Label:        "{3}{G}{G}, Sacrifice this creature: Create an X/X colorless Golem artifact creature token, where X is three times this creature's power. It gains haste until end of turn. Activate only as a sorcery.",
			Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
			Cost:         effects.Plus(effects.ManaCost("{3}{G}{G}"), effects.SacrificeThis()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				x := 0
				if info, ok := ctx.SourcePermanent(); ok {
					x = 3 * info.Power
				}
				return effects.S82BroB21XGolem(ctx, x, true)
			},
		}},
	})
}
