package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Acquisitions Expert — Creature — Human Rogue {1}{B}, 1/2 (Zendikar Rising,
// #5508):
//
//	"When this creature enters, target opponent reveals a number of cards
//	 from their hand equal to the number of creatures in your party. You
//	 choose one of those cards. That player discards that card. (Your party
//	 consists of up to one each of Cleric, Rogue, Warrior, and Wizard.)"
//
// The party (CR 700.8) is counted as the trigger resolves, the Expert itself
// being the Rogue if nothing else is. The opponent picks which cards to
// reveal (a hand no larger than the party is revealed whole), only those are
// revealed, and the Expert's controller picks one of them to be discarded:
// the shared RevealFromHandPick flow Klaw, Sonic Subjugator uses.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "07e5f59e-3a98-4e01-a368-0be60e0463ba",
		Name:         "Acquisitions Expert",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Acquisitions Expert — target opponent reveals a number of cards from their hand equal to the creatures in your party; you choose one and they discard it", acquisitionsExpertEnters),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}

func acquisitionsExpertEnters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	return acquisitionsExpertReveal.Ask(ctx, effects.TargetedPlayer(ctx), effects.FinB9PartySize(g, item.Controller))
}

// acquisitionsExpertReveal is the reveal-N-and-pick flow. Both keys are
// on-disk identities: never renamed, never reused.
var acquisitionsExpertReveal = effects.NewRevealFromHandPick("Acquisitions Expert", "discards",
	"choose-cards/acquisitions-expert-reveal", "choose-cards/acquisitions-expert-discard",
	func(ctx *effects.Context, owner, card uuid.UUID) error {
		return ctx.Game.DiscardCardsKeyedForEffect(owner, []uuid.UUID{card}, ctx.Source(), game.KeyedTail{})
	})
