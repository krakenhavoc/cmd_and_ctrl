package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lord Skitter, Sewer King — Legendary Creature — Rat Noble {2}{B}, 3/3
// (Wilds of Eldraine, #4471):
//
//	"Whenever another Rat you control enters, exile up to one target card
//	 from an opponent's graveyard.
//	 At the beginning of combat on your turn, create a 1/1 black Rat
//	 creature token with "This token can't block.""
//
// The enters trigger watches any other permanent with the Rat subtype
// entering under your control (a token Rat counts, a changeling counts);
// "Rat" is a word of the card's text. Its "up to one" is Gollum, the
// Abandoned's zero-to-one graveyard clause, a card whose owner is not
// you. The combat trigger makes Twisted Sewer-Witch's Rat, which is
// itself another Rat entering, so on your turn it also lets you exile a
// card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f5bed085-bb7a-4250-9599-b2a8a95e6b6f",
		Name:         "Lord Skitter, Sewer King",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev, source, g, source.TextSubtype("Rat"))
			},
			Targets: effects.TargetCardInGraveyard("up to one target card from an opponent's graveyard", effects.Not(effects.YouOwn())).WithCount(0, 1),
			Key:     "Lord Skitter, Sewer King — exile up to one target card from an opponent's graveyard",
			Effect:  effects.ExileFirstTarget,
		},
			effects.AtBeginningOfYourCombat("Lord Skitter, Sewer King — create a 1/1 black Rat creature token with \"This token can't block.\"",
				effects.Do(effects.CreateToken{Template: effects.CantBlockRatToken(), N: 1})),
		},
	})
}
