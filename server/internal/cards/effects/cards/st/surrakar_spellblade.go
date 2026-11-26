package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Surrakar Spellblade — Creature — Surrakar {1}{U}{U}, 2/1 (Rise of the
// Eldrazi, S82 OLD batch 482, #5581):
//
//	"Whenever you cast an instant or sorcery spell, you may put a charge
//	 counter on this creature.
//	 Whenever this creature deals combat damage to a player, you may draw X
//	 cards, where X is the number of charge counters on it."
//
// Two "you may" triggers (CR 603.5). X is read as the draw resolves: the
// Spellblade's charge counters now, or as it last existed if it has left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e87813a3-893d-4416-9263-2ea2e3140490",
		Name:         "Surrakar Spellblade",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.InstantOrSorceryCastByYou(ev, source, g)
			}, "Surrakar Spellblade — put a charge counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterCharge, N: 1})),
				"Surrakar Spellblade — put a charge counter on it?"),
			effects.OptionalDrawThatMany(
				effects.WheneverThisDealsCombatDamageToAPlayer("Surrakar Spellblade — draw X cards, where X is the number of charge counters on it",
					effects.Do(effects.Instruction(func(ctx *effects.Context) error {
						return effects.DrawCards{N: surrakarSpellbladeCharge(ctx)}.Apply(ctx)
					}))),
				"Surrakar Spellblade — draw cards equal to its charge counters?",
				func(g *game.Game, item *game.StackItem) int {
					return surrakarSpellbladeCharge(effects.NewContext(g, item))
				}),
		},
	})
}

// surrakarSpellbladeCharge is the charge counters on the Spellblade, as it
// is now or as it last existed (CR 608.2h).
func surrakarSpellbladeCharge(ctx *effects.Context) int {
	self, ok := ctx.SourcePermanent()
	if !ok {
		return 0
	}
	return self.Counters[game.CounterCharge]
}
