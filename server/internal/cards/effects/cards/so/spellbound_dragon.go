package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spellbound Dragon — Creature — Dragon {3}{U}{R}, 3/5 (Commander 2017, S82
// older sets):
//
//	"Flying
//	 Whenever this creature attacks, draw a card, then discard a card. This
//	 creature gets +X/+0 until end of turn, where X is the discarded card's
//	 mana value."
//
// Malcolm's chain: the draw settles first (Dredge or a CR 616 order can pause
// it), then the discard is asked, and the pump follows the discard as a keyed
// continuation reading the card that was discarded (CR 608.2c). X is the
// discarded card's mana value (CR 202.3e: 0 for a land), 0 when nothing was
// discarded.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "86b8a9ef-7cdf-47e9-accb-16fe1eab718f",
		Name:            "Spellbound Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Spellbound Dragon — draw a card, then discard a card; it gets +X/+0 until end of turn, where X is the discarded card's mana value",
				func(g *game.Game, item *game.StackItem) error {
					return g.DrawNThenForEffect(item.Controller, 1, game.DrawThen{
						Ref: spellboundDragonDiscard, Player: item.Controller, Source: item.SourceCardID,
					})
				}),
		},
	})
}

// spellboundDragonDiscard asks for the discard, then hands it to the pump.
// An on-disk key.
var spellboundDragonDiscard = game.RegisterDrawThen("spellbound-dragon/discard", func(g *game.Game, d game.DrawThen) error {
	return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
		Player:   d.Player,
		Source:   d.Source,
		N:        1,
		Question: "Spellbound Dragon — discard a card",
	}, spellboundDragonPump, nil)
})

// spellboundDragonPump is +X/+0 until end of turn for the Dragon, X the mana
// value of the card just discarded. An on-disk key.
var spellboundDragonPump = game.RegisterDiscardThen("choose-cards/spellbound-dragon-pump", func(g *game.Game, r game.DiscardsDone) error {
	discarded := r.Discarded.Cards()
	if len(discarded) == 0 || !effects.OnBattlefield(g, r.Source) {
		return nil
	}
	card, ok := g.LookupCardForEffect(discarded[0])
	if !ok {
		return nil
	}
	x, ok := g.ManaValueForEffect(card)
	if !ok || x <= 0 {
		return nil
	}
	return effects.BoostUntilEOT{
		Target: r.Source,
		Power:  x,
		Label:  "Spellbound Dragon — +X/+0 until end of turn",
	}.Apply(effects.KeyedContext(g, r.Player, r.Source))
})
