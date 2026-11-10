package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Daring Saboteur — Creature — Human Pirate {1}{U}, 2/1:
//
//	"{2}{U}: This creature can't be blocked this turn.
//	 Whenever this creature deals combat damage to a player, you may draw
//	 a card. If you do, discard a card."
//
// The unblockable grant is the until-end-of-turn restriction registry
// (CR 509.1b), pinned to the permanent that activated it (CR 400.7). The
// loot is the shared "you may draw, then discard" prompt.
func init() {
	Register(Spec{
		OracleID:     "d36f64a0-53f7-4c56-9f40-a7927a5c2608",
		Name:         "Daring Saboteur",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{U}: This creature can't be blocked this turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{2}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				return RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked,
					Label: "Daring Saboteur — can't be blocked this turn"}.Apply(NewContext(g, item))
			},
		}},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Daring Saboteur — you may draw a card; if you do, discard a card",
				func(g *game.Game, item *game.StackItem) error {
					return MayChoice{
						Question: "Daring Saboteur — draw a card, then discard a card?",
						Draws:    1,
						YesLabel: "Draw",
						NoLabel:  "Decline",
						Then:     celestusAnswer,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
