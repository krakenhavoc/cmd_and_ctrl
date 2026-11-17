package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "be3c138b-045b-4b08-a12f-a5679d972051",
		Name:         "Ominous Parcel",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{2}, {T}, Sacrifice this artifact: Search your library for a basic land card, reveal it, put it into your hand, then shuffle.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Tutors: 1},
				Cost:    Plus(ManaCost("{2}"), TapCost(), SacrificeThis()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: IsBasicLand,
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Ominous Parcel — a basic land card, to your hand",
					}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "{5}, {T}, Sacrifice this artifact: It deals 4 damage to target creature.",
				Purpose: ForTargets(DamageToTarget(0, 4)),
				Cost:    Plus(ManaCost("{5}"), TapCost(), SacrificeThis()),
				Targets: TargetCreature("target creature"),
				Effect:  DealDamageToTheTarget(4),
			},
		},
	})
}
