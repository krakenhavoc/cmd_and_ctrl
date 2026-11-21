package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cursed Wombat — Creature — Nightmare Wombat {B}{G}, 2/3 (Modern Horizons 3):
//
//	"{2}{B}{G}: Adapt 2.
//	 Permanents you control have "Whenever one or more +1/+1 counters are put
//	 on this permanent, put an additional +1/+1 counter on it. This ability
//	 triggers only once each turn.""
//
// The granted ability is each permanent's own trigger (ADR 0093), limited to
// one per permanent per turn by its own label, so the additional counter's
// own placement does not trigger it again.
func init() {
	const (
		key   = "cursed-wombat/additional-counter"
		label = "Cursed Wombat — put an additional +1/+1 counter on it"
		text  = "Whenever one or more +1/+1 counters are put on this permanent, put an additional +1/+1 counter on it. This ability triggers only once each turn."
	)
	effects.Register(effects.Spec{
		OracleID:     "fd077d1a-f395-41c2-b1df-dd5e2f723abf",
		Name:         "Cursed Wombat",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}{G}: Adapt 2.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{2}{B}{G}"),
			Effect:  effects.AdaptTwo,
		}},
		Grants: []effects.AbilityGrant{{
			Key: key,
			Triggered: []game.TriggeredAbility{
				effects.On(game.EventCounterPlaced, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
					return effects.GapB50CountersPlacedOnThis(ev, source, lki, g) && !effects.B11TriggeredThisTurn(g, source.InstanceID, label)
				}, label, effects.PutCounterOnSelf),
			},
			Text: text,
		}},
		Static: []game.StaticAbility{effects.Labelled(effects.GrantAbilities(func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.Controller == source.Controller
		}, key), "Permanents you control have \""+text+"\"")},
	})
}
