package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ominous Parcel — Artifact {1} (Streets of New Capenna, #5491):
//
//	"{2}, {T}, Sacrifice this artifact: Search your library for a basic
//	 land card, reveal it, put it into your hand, then shuffle.
//	 {5}, {T}, Sacrifice this artifact: It deals 4 damage to target
//	 creature."
//
// Expedition Map's search (basic land) and Aeolipile's sacrifice-and-damage:
// the Parcel is sacrificed as a cost and still deals the damage from its
// last known information (CR 113.7a); a target that left is not damaged
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "be3c138b-045b-4b08-a12f-a5679d972051",
		Name:         "Ominous Parcel",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}, {T}, Sacrifice this artifact: Search your library for a basic land card, reveal it, put it into your hand, then shuffle.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Tutors: 1},
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeThis()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: effects.IsBasicLand,
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Ominous Parcel — a basic land card, to your hand",
					}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "{5}, {T}, Sacrifice this artifact: It deals 4 damage to target creature.",
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 4)),
				Cost:    effects.Plus(effects.ManaCost("{5}"), effects.TapCost(), effects.SacrificeThis()),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.DealDamageToTheTarget(4),
			},
		},
	})
}
