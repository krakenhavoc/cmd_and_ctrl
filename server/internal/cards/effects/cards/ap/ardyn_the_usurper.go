package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ardyn, the Usurper — Legendary Creature — Elder Human Noble
// {5}{B}{B}{B}, 4/4:
//
//	"Demons you control have menace, lifelink, and haste.
//	 Starscourge — At the beginning of combat on your turn, exile up to
//	 one target creature card from a graveyard. If you exiled a card
//	 this way, create a token that's a copy of that card, except it's
//	 a 5/5 black Demon."
//
// The first ability is three layer-6 tribal grants (TribalKeywordGrant)
// over the Demons Ardyn's controller controls; Ardyn is not a Demon and
// gets none of them.
//
// Starscourge is The Scarab God's reanimation as a trigger: the card is
// exiled from any graveyard, and only if it really was exiled is a token
// copy made, except it's a 5/5 black Demon. That exception sets the
// copy's creature types rather than adding to them (CR 707.9a, as The
// Scarab God's 4/4 black Zombie does), and the copy then has the
// keywords Ardyn grants to Demons. With no creature card in any
// graveyard the trigger has nothing to target and does nothing, which
// is what "up to one" with no choice comes to.
//
// No simplification.
func init() {
	demons := effects.TribeFilter{Tribes: []string{"Demon"}, YoursOnly: true}
	effects.Register(effects.Spec{
		OracleID:     "9ae13026-960a-4d31-b775-d47209a1e313",
		Name:         "Ardyn, the Usurper",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(demons, "menace"),
			effects.TribalKeywordGrant(demons, "lifelink"),
			effects.TribalKeywordGrant(demons, "haste"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.AtBeginningOfYourCombat("Ardyn, the Usurper — Starscourge: exile a creature card and copy it as a 5/5 black Demon",
				ardynStarscourge),
				effects.TargetCardInGraveyard("up to one target creature card from a graveyard", effects.Creature()).WithCount(0, 1)),
		},
	})
}

// ardynStarscourge exiles the chosen card and, if it was exiled, makes
// the Demon copy.
func ardynStarscourge(g *game.Game, item *game.StackItem) error {
	return effects.ExileTargetThenTokenCopy(g, item, ardynDemonCopy)
}

// ardynDemonCopy is the exile's keyed "then" (#4253 slice 10a).
var ardynDemonCopy = effects.ExileThenTokenCopyTail("exile-then-token-copy/ardyn-the-usurper", func(ctx *effects.Context) func(t *game.Card) {
	return ardynDemonExceptionThrough(ctx.Game.TextReaderForEffect(ctx.Item))
})

// ardynDemonExceptionThrough is "except it's a 5/5 black Demon", with
// "black" and "Demon" read through `reader`, the text of the ability
// that makes the copy (CR 612.2, #3984); nil reads the printed words.
func ardynDemonExceptionThrough(reader *game.Card) func(t *game.Card) {
	return func(t *game.Card) {
		t.Power = 5
		t.Toughness = 5
		t.VariableToughness = false
		t.SetCopyExceptionColors(reader.TextColors("B")...)
		t.TypeLine = effects.RetypedTypeLine(t.TypeLine, reader.TextSubtypes("Demon")...)
	}
}
