package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Athreos, God of Passage — Legendary Enchantment Creature — God {1}{W}{B}, 5/4
// (Secret Lair Drop):
//
//	"Indestructible
//	 As long as your devotion to white and black is less than seven, Athreos
//	 isn't a creature.
//	 Whenever another creature you own dies, return it to your hand unless
//	 target opponent pays 3 life."
//
// The devotion clause is the Theros Gods' (GodUnlessDevotionToColors, Mogis).
// The trigger is a death the controller OWNS (a stolen creature that dies is
// returned to its owner's hand, and the trigger is the owner's Athreos'; a
// creature of yours that an opponent controls dies and still counts), another
// than Athreos, read from the creature as it last existed (CR 603.10a). The
// opponent is chosen as the ability goes on the stack. As it resolves that
// opponent either pays 3 life (CR 119.4) or the card returns to its owner's hand,
// if it is still the card that died, in the graveyard (CR 400.7). A token that
// died is gone, so there is nothing to return and nothing to ask.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "af206aed-1f58-4f5a-ba42-0e0d1f282940",
		Name:            "Athreos, God of Passage",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Static:          []game.StaticAbility{effects.GodUnlessDevotionToColors(7, "W", "B")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				return ok && dead.InstanceID != source.InstanceID && dead.Owner == source.Controller && !dead.IsToken()
			}, "Athreos, God of Passage — return it to your hand unless target opponent pays 3 life",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					victim, ok := effects.FirstLegalPlayerTarget(ctx)
					if !ok {
						return nil
					}
					return athreosAsk.Ask(ctx, victim, item.Trigger.Event.CardID)
				}),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}

var athreosAsk = effects.NewOldB015PayThreeLifeUnless("option-pick/athreos-god-of-passage-pay-or-return",
	"Athreos, God of Passage — let the creature card return to its owner's hand, or pay 3 life?", "Let it return to hand",
	func(ctx *effects.Context, card uuid.UUID) error {
		if !effects.DiedCardStillInGraveyard(ctx, card) {
			return nil
		}
		return effects.ReturnFromGraveyard{Target: card, Dest: game.ZoneHand}.Apply(ctx)
	})
