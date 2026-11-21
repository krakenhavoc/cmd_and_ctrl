package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nick Fury, Spymaster — Legendary Creature — Human Spy Hero {4}{W}, 4/4
// (Marvel Super Heroes Commander, #4455):
//
//	"First strike
//	 Whenever a creature you control attacks alone, draw a card. Then you
//	 may put a creature card with mana value 3 or less from your hand onto
//	 the battlefield. It enters tapped and attacking and gains indestructible
//	 until end of turn."
//
// Black Panther, Claws of Bast's attacks-alone trigger. The draw comes first,
// so the card drawn is among the candidates. The put is Ultra Magnus's
// (HandAttackerPutter): the card enters tapped and attacking a defender its
// controller chooses, the lone attacker's own defender first (CR 508.4,
// 506.3c), and was never declared as an attacker. "It ... gains indestructible"
// is the tail on the entered card: nothing entered, nothing gains it. The
// grant is pinned to the permanent that entered, until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b20860da-2e71-414a-a512-1cce7ba64079",
		Name:            "Nick Fury, Spymaster",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.MshB1YouAttackAlone,
				"Nick Fury, Spymaster — draw a card, then you may put a creature card with mana value 3 or less from your hand onto the battlefield tapped and attacking",
				nickFuryAttacksAlone),
		},
	})
}

// nickFuryPutter is the put: keys "choose-cards/nick-fury-spymaster-put" and
// "option-pick/nick-fury-spymaster-attacks".
var nickFuryPutter = effects.NewHandAttackerPutter("nick-fury-spymaster", "Nick Fury, Spymaster", nickFuryIndestructible)

func nickFuryAttacksAlone(g *game.Game, item *game.StackItem) error {
	if err := (effects.DrawCards{N: 1}).Apply(effects.NewContext(g, item)); err != nil {
		return err
	}
	nickFuryPutter.Offer(g, item, effects.HandCardsMatching(g, item.Controller, effects.And(effects.Creature(), effects.ManaValueLE(3))),
		"Nick Fury, Spymaster — you may put a creature card with mana value 3 or less from your hand onto the battlefield tapped and attacking")
	return nil
}

// nickFuryIndestructible is "It ... gains indestructible until end of turn",
// on the card that entered. An on-disk key.
var nickFuryIndestructible = effects.EnteredTail("nick-fury-spymaster/indestructible", func(ctx *effects.Context, entered uuid.UUID, _ game.TailOutcome) error {
	if entered == uuid.Nil {
		return nil // nothing entered
	}
	return effects.GrantKeywordUntilEOT{
		Target:   entered,
		Keywords: []string{"indestructible"},
		Label:    "Nick Fury, Spymaster — indestructible until end of turn",
	}.Apply(ctx)
})
