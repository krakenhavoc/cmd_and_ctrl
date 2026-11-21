package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spider-Man 2099, Miguel O'Hara — Legendary Creature — Spider Human Hero
// {4}{U}, 4/4 (Marvel's Spider-Man, #4460):
//
//	"When Spider-Man 2099 enters, return up to one target creature to its
//	 owner's hand.
//	 Whenever one or more creatures you control deal combat damage to a
//	 player, draw a card."
//
// The enters trigger takes "up to one", so it may pick nothing and is
// otherwise Bottomless Pool's bounce (any creature, not only an
// opponent's). The damage trigger is one per player the creatures connect
// with in a damage step (CR 603.2c), whatever their number.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c0280d66-699f-4139-828b-a0f79ae2d20d",
		Name:         "Spider-Man 2099, Miguel O'Hara",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Spider-Man 2099 — return up to one target creature to its owner's hand",
				effects.BounceTheTargetToHand),
				effects.TargetCreature("up to one target creature").WithCount(0, 1)),
			effects.WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(nil,
				"Spider-Man 2099 — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
