package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thrull Surgeon — Creature — Thrull {1}{B}, 1/1 (Exodus, S82 OLD b737, #5626):
//
//	"{1}{B}, Sacrifice this creature: Look at target player's hand and choose
//	 a card from it. That player discards that card. Activate only as a
//	 sorcery."
//
// Extortion's look-and-choose with a count of one: only the controller sees
// the hand and picks (CR 701.9a is the target's discard of the chosen card).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ead5a2bf-f48a-43a2-89bb-df3cc587c843",
		Name:         "Thrull Surgeon",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{1}{B}, Sacrifice this creature: Look at target player's hand and choose a card from it. That player discards that card. Activate only as a sorcery.",
			Purpose:      effects.ForTargets(game.TargetPurpose{Slot: 0, Discards: 1}),
			Cost:         effects.Plus(effects.ManaCost("{1}{B}"), effects.SacrificeThis()),
			SorcerySpeed: true,
			Targets:      effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.ChooseFromRevealedHand{Player: effects.TargetedPlayer(ctx), Label: "card", Count: 1, Look: true}.Apply(ctx)
			},
		}},
	})
}
