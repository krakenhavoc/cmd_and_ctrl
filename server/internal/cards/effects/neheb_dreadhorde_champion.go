package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Neheb, Dreadhorde Champion — Legendary Creature — Zombie Minotaur Warrior
// {2}{R}{R}, 5/4:
//
//	"Trample
//	 Whenever Neheb deals combat damage to a player or planeswalker, you may
//	 discard any number of cards. If you do, draw that many cards and add
//	 that much {R}. Until end of turn, you don't lose this mana as steps and
//	 phases end."
//
// The discard is the player's choice of any number, up to the whole hand
// (none is allowed), and the draw and the mana are the run's continuation
// told how many cards really left the hand (nehebDiscardDrawMana): that many
// cards drawn, then that much {R} added carrying the keep mark
// (KeepManaUntilEndOfTurn), which the step-boundary sweep reads. With an empty
// hand there is nothing to discard and nothing happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ed8e9a2e-5fb8-4f39-b954-dbcbc96d909f",
		Name:            "Neheb, Dreadhorde Champion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, ThisDealtCombatDamageToAPlayerOrPlaneswalker,
				"Neheb, Dreadhorde Champion — you may discard any number of cards; if you do, draw that many and add that much {R}",
				func(g *game.Game, item *game.StackItem) error {
					hand := len(allHandCardIDs(g, item.Controller))
					if hand == 0 {
						return nil
					}
					return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
						Player:   item.Controller,
						Source:   item.SourceCardID,
						N:        hand,
						UpTo:     true,
						Question: "Neheb, Dreadhorde Champion — discard any number of cards, then draw that many and add that much {R}",
					}, nehebDiscardDrawMana.then, nil)
				}),
		},
	})
}
