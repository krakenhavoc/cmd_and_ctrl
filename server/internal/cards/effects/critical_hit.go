package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Critical Hit — Instant {1}{R} (Adventures in the Forgotten Realms,
// #5681):
//
//	"Target creature gains double strike until end of turn.
//	 When you roll a natural 20, return this card from your graveyard to
//	 your hand. (A natural 20 is a roll that displays 20 on the die.)"
//
// The second ability works only from the graveyard (InGraveyard: "return
// this card from your graveyard" says where it lives), and "you" there is
// the card's owner (CR 108.4). A natural 20 is the number the die shows
// before any modifier (CR 706.2), so the announcement's natural result is
// read and the instruction's modifier is not: The Deck of Many Things
// showing 20 with three cards in hand is a natural 20, and a 17 plus 3 is
// not. An ignored roll never happened (CR 706.6). Two dice that both show
// 20 trigger it twice; the second finds the card gone and does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "992e6845-1c6c-4b74-839c-317823bde15b",
		Name:         "Critical Hit",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Targets: game.ForTargets(BenefitUntilEOT(0, 0, 0, "double strike")), Answers: game.AnswerCombatGrant},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return GrantKeywordUntilEOT{Target: id, Keywords: []string{"double strike"}, Label: "Critical Hit — double strike"}.Apply(ctx)
		},
		Triggered: []game.TriggeredAbility{
			InGraveyard(On(game.EventRollDie, youRolledANatural20,
				"Critical Hit — return it from your graveyard to your hand",
				returnThisCardFromYourGraveyardToYourHand)),
		},
	})
}

// youRolledANatural20 is "when you roll a natural 20": a die of the
// source's controller that displays 20 (the event's Amount, its natural
// result; CR 706.2), whatever the instruction's modifier.
func youRolledANatural20(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventRollDie && ev.PlanarFace == "" && ev.Actor != uuid.Nil && ev.Actor == source.Controller && ev.Amount == 20
}
