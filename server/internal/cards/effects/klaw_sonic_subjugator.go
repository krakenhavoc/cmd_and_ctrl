package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Klaw, Sonic Subjugator — Legendary Creature — Human Rogue Villain {2}{B},
// 2/2:
//
//	"Sonic Attack — When Klaw enters, target player reveals a number of
//	 cards from their hand equal to one plus the number of creature cards
//	 in your graveyard. You choose one of them. That player discards that
//	 card."
//
// The number is counted as the trigger resolves (the creature cards in its
// controller's graveyard then; Klaw is on the battlefield, not among them).
// The TARGET player chooses which cards to reveal (a choose_cards prompt
// over their own hand; with the number at least their hand size they reveal
// it all and nobody is asked), and only those are revealed, to every player
// (CR 701.20a). Klaw's controller then picks one of the revealed cards and
// its owner discards it (CR 701.9a): through the one discard path, so
// madness and discard triggers see it. An empty hand reveals nothing and
// asks nothing. RevealFromHandPick is the shared flow with Taster of Wares;
// both prompts are keyed and plain data (#4588).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1c233e08-d5a3-48d7-9074-20393896a480",
		Name:         "Klaw, Sonic Subjugator",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Sonic Attack — target player reveals one plus the number of creature cards in your graveyard from their hand; you choose one and they discard it", klawEnters),
				TargetPlayer("target player")),
		},
	})
}

func klawEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	n := 1 + b11CreatureCardsInGraveyard(g, item.Controller)
	return klawReveal.Ask(ctx, TargetedPlayer(ctx), n)
}

// klawReveal is the shared reveal-N-and-pick flow (RevealFromHandPick): the
// target player's cards are revealed, Klaw's controller picks one and its
// owner discards it. Both keys are on-disk identities: never renamed, never
// reused.
var klawReveal = NewRevealFromHandPick("Klaw, Sonic Subjugator", "discards",
	"choose-cards/klaw-sonic-subjugator-reveal", "choose-cards/klaw-sonic-subjugator-discard",
	func(ctx *Context, owner, card uuid.UUID) error {
		return ctx.Game.DiscardCardsKeyedForEffect(owner, []uuid.UUID{card}, ctx.Source(), game.KeyedTail{})
	})
