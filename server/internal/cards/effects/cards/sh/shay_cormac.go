package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shay Cormac — Legendary Creature — Human Knight Rogue {W}{B}, 1/1
// (Assassin's Creed, #5948):
//
//	"{1}: Permanents your opponents control lose hexproof, indestructible,
//	 protection, shroud, and ward until end of turn.
//	 Whenever a creature an opponent controls becomes the target of a spell or
//	 ability you control, put a bounty counter on that creature.
//	 Whenever a creature with a bounty counter on it dies, put two +1/+1
//	 counters on Shay Cormac."
//
// The first ability is Shadowspear's strip with the rest of the list: one
// scoped record over the permanents your opponents control as it resolves
// (CR 611.2c locks the set, so a permanent that arrives later keeps its
// keywords). Hexproof, indestructible and shroud are exact tokens
// (RemoveKeywordsMod); protection and "hexproof from" are families of tokens
// (RemoveKeywordPrefixesMod, #5948); ward is a triggered ability, so losing it
// is the record telling the ward gate not to trigger (SuppressWardMod). It
// removes and does not add "can't have", so a later grant puts one back.
//
// The second is a counter on the creature as the targeting happens: one
// trigger per target slot (CR 115.3) naming a creature an opponent of Shay's
// controller controls and a spell or ability Shay's controller controls.
//
// The third is any creature with a bounty counter dying, judged as it last
// existed (CR 603.10a: the exit's last-known counters), whoever controlled it.
//
// No simplification.
func init() {
	const (
		bountyLabel = "Shay Cormac — put a bounty counter on that creature"
		deathLabel  = "Shay Cormac — put two +1/+1 counters on Shay Cormac"
	)
	effects.Register(effects.Spec{
		OracleID:     "95892392-2ed3-4c1a-bbaa-4083f526769c",
		Name:         "Shay Cormac",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}: Permanents your opponents control lose hexproof, indestructible, protection, shroud, and ward until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerRestrict},
			Cost:    effects.ManaCost("{1}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.UntilEndOfTurn(effects.NewContext(g, item), uuid.Nil, effects.OpponentControls(),
					"Shay Cormac — opponents' permanents lose hexproof, indestructible, protection, shroud, and ward",
					game.RemoveKeywordsMod("hexproof", "indestructible", "shroud"),
					game.RemoveKeywordPrefixesMod("protection from ", "hexproof from "),
					game.SuppressWardMod())
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventBecomesTarget || ev.CardID == uuid.Nil || ev.Actor != source.Controller || !effects.OnBattlefield(g, ev.CardID) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() && c.Controller != source.Controller
			}, bountyLabel, func(g *game.Game, item *game.StackItem) error {
				id := item.Trigger.Event.CardID
				if !effects.OnBattlefield(g, id) {
					return nil
				}
				return effects.AddCounter{Target: id, Kind: shayBountyCounter, N: 1}.Apply(effects.NewContext(g, item))
			}),
			effects.On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				if _, ok := effects.DiedCreature(ev, g); !ok {
					return false
				}
				return g.LastKnownCountersForEffect(ev.CardID)[shayBountyCounter] > 0
			}, deathLabel, effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 2})),
		},
	})
}

// shayBountyCounter is the bounty counter's kind. On-disk: never renamed.
const shayBountyCounter = "bounty"
