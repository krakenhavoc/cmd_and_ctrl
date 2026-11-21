package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ilharg, the Raze-Boar — Legendary Creature — Boar God {3}{R}{R}, 6/6
// (Ravnica Remastered, #5948):
//
//	"Trample
//	 Whenever Ilharg attacks, you may put a creature card from your hand onto
//	 the battlefield tapped and attacking. Return that creature to your hand at
//	 the beginning of the next end step.
//	 When Ilharg dies or is put into exile from the battlefield, you may put it
//	 into its owner's library third from the top."
//
// The put is Kinscaer Sentry's (hand_attacker_put.go): the creature enters
// tapped and attacking a player, planeswalker or battle its controller chooses
// (CR 508.4), never declared as an attacker, so it triggers nothing that watches
// attacks. The return is Zara's delayed trigger: at the next end step, whoever's
// turn it is, the creature goes back to its owner's hand if it is still the
// object that entered (CR 400.7). "Your hand" is the hand of the card's owner,
// the player it came from. The last ability is God-Eternal Oketra's: a leaves-the-
// battlefield look-back for itself (CR 603.10a) that puts the card third from
// the top of its owner's library if it is still in the graveyard or in exile.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8d571129-9030-47e0-9624-a49fb63e5a1b",
		Name:            "Ilharg, the Raze-Boar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Ilharg, the Raze-Boar — you may put a creature card from your hand onto the battlefield tapped and attacking; return it to your hand at the beginning of the next end step",
				gapB74IlhargAttack),
			effects.Optional(effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.B22SelfDiedOrWasExiledFromBattlefield(ev, source)
			}, "Ilharg, the Raze-Boar — put it into its owner's library third from the top", effects.TuckSelfThirdFromTop),
				"Ilharg, the Raze-Boar — put it into its owner's library third from the top?"),
		},
	})
}

// gapB74IlhargPutter is the put: keys "choose-cards/ilharg-the-raze-boar-put"
// and "option-pick/ilharg-the-raze-boar-attacks".
var gapB74IlhargPutter = effects.NewHandAttackerPutter("ilharg-the-raze-boar", "Ilharg, the Raze-Boar", gapB74IlhargReturnTail)

// gapB74IlhargReturnTail is "return that creature to your hand at the beginning
// of the next end step", scheduled once the creature entered. An on-disk key.
var gapB74IlhargReturnTail = effects.EnteredTail("ilharg-the-raze-boar/return-at-end-step", func(ctx *effects.Context, entered uuid.UUID, _ game.TailOutcome) error {
	if entered == uuid.Nil {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(entered)
	if !ok {
		return nil
	}
	return effects.ScheduleDelayedTrigger{
		Label:  "Ilharg, the Raze-Boar — return that creature to your hand",
		Cards:  []uuid.UUID{entered},
		Body:   effects.ZaraReturnToHandBody,
		Params: game.EffectParams{Object: game.ObjectRef{ID: entered, Epoch: c.ObjectEpoch}},
	}.Apply(ctx)
})

func gapB74IlhargAttack(g *game.Game, item *game.StackItem) error {
	gapB74IlhargPutter.Offer(g, item, effects.HandCardsMatching(g, item.Controller, effects.Creature()),
		"Ilharg, the Raze-Boar — you may put a creature card from your hand onto the battlefield tapped and attacking")
	return nil
}
