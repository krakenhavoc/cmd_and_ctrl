package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ellie, Vengeful Hunter — Legendary Creature — Human Survivor {1}{B}{R}, 3/1
// (Secret Lair Drop):
//
//	"Pay 2 life, Sacrifice another creature: Ellie deals 2 damage to target
//	 player and gains indestructible until end of turn.
//	 Partner—Survivors (You can have two commanders if both have this
//	 ability.)"
//
// Both costs are paid as the ability is activated (CR 602.2b): the life, and a
// creature other than Ellie (object identity, CR 109.1). The damage is to a
// target player, and the indestructible is Yahenni's grant; with its only target
// illegal the whole ability does nothing (CR 608.2b). Partner—Survivors is a
// deck-construction rule read off the oracle text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e619cb66-8c0d-408a-829c-6f02bf25fd0f",
		Name:         "Ellie, Vengeful Hunter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Pay 2 life, Sacrifice another creature: Ellie deals 2 damage to target player and gains indestructible until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive | game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.PayLife(2), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				player, ok := effects.FirstLegalPlayerTarget(ctx)
				if !ok {
					return nil
				}
				if err := (effects.DealDamage{Source: item.SourceCardID, Target: player, Amount: 2}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{
					Target:   item.SourceCardID,
					Keywords: []string{"indestructible"},
					Label:    "Ellie, Vengeful Hunter — indestructible until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
