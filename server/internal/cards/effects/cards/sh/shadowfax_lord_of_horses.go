package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const shadowfaxLabel = "Shadowfax, Lord of Horses — you may put a creature card with lesser power from your hand onto the battlefield tapped and attacking"

// Shadowfax, Lord of Horses — Legendary Creature — Horse {3}{R}{W}, 4/4 (The
// Lord of the Rings: Tales of Middle-earth Commander):
//
//	"Horses you control have haste.
//	 Whenever Shadowfax attacks, you may put a creature card with lesser
//	 power from your hand onto the battlefield tapped and attacking."
//
// The first line is a tribal keyword grant. The attack trigger is Minion of
// the Mighty's hand-to-battlefield put (HandAttackerPutter): "lesser power"
// is a creature card whose power is below Shadowfax's as the trigger
// resolves (his last-known power if he has left). The entering card was
// never declared as an attacker, so it triggers no "whenever attacks"
// ability, and it is tapped and attacking a defender its controller picks.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aa630f3d-9443-46fe-a54f-8b3c9fcb872e",
		Name:         "Shadowfax, Lord of Horses",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{"Horse"}, YoursOnly: true}, "haste"),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.ThisAttacked, shadowfaxLabel, func(g *game.Game, item *game.StackItem) error {
				power := effects.B29PowerAsItLastStood(g, item.SourceCardID)
				shadowfaxPutter.Offer(g, item, effects.HandCardsMatching(g, item.Controller, effects.And(effects.Creature(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
					return c.Power < power
				})), shadowfaxLabel)
				return nil
			}),
		},
	})
}

var shadowfaxPutter = effects.NewHandAttackerPutter("shadowfax-lord-of-horses", "Shadowfax, Lord of Horses",
	effects.EnteredTail("shadowfax-lord-of-horses/entered", func(*effects.Context, uuid.UUID, game.TailOutcome) error { return nil }))
