package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Minion of the Mighty — Creature — Kobold {R}, 0/1:
//
//	"Menace
//	 Pack tactics — Whenever this creature attacks, if you attacked with
//	 creatures with total power 6 or greater this combat, you may put a
//	 Dragon creature card from your hand onto the battlefield tapped and
//	 attacking."
//
// The intervening "if" (CR 603.4) is read when the attack is declared and
// again as the ability resolves, over the creatures declared as attackers
// this combat (the shared pack-tactics reading). The Dragon is a creature
// card with the Dragon creature type, and you choose what it attacks as it
// enters, which need not be what Minion attacks (CR 508.4). It was never
// declared as an attacker, so nothing that watches for attacks sees it.
//
// No simplification.
func init() {
	const label = "Minion of the Mighty — pack tactics: you may put a Dragon creature card from your hand onto the battlefield tapped and attacking"
	effects.Register(effects.Spec{
		OracleID:        "058062c7-071b-4b14-903d-ae8d2876b37f",
		Name:            "Minion of the Mighty",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.CardID == source.InstanceID && ev.Actor == source.Controller &&
					effects.PackTacticsAttackingPower(g, source.Controller) >= 6
			}, label, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if effects.PackTacticsTotalPower(g, item.Controller, ctx.Trigger().Event.Batch) < 6 {
					return nil
				}
				minionPutter.Offer(g, item, effects.HandCardsMatching(g, item.Controller, effects.And(effects.Creature(), effects.CmdrB5Dragon())), label)
				return nil
			}),
		},
	})
}

var minionPutter = effects.NewHandAttackerPutter("minion-of-the-mighty", "Minion of the Mighty",
	effects.EnteredTail("minion-of-the-mighty/entered", func(*effects.Context, uuid.UUID, game.TailOutcome) error { return nil }))
