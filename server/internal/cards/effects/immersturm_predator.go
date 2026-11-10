package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Immersturm Predator — Creature — Vampire Dragon {2}{B}{R}, 3/3
// (Foundations, #4465):
//
//	"Flying
//	 Whenever this creature becomes tapped, exile up to one target card
//	 from a graveyard and put a +1/+1 counter on this creature.
//	 Sacrifice another creature: This creature gains indestructible until
//	 end of turn. Tap it."
//
// "Becomes tapped" is Attentive Sunscribe's ThisBecameTapped (the tap
// event, or an attack that taps it). The target is Ambush Wolf's "up to
// one target card from a graveyard": with none chosen the counter still
// goes on, and a chosen card that has left its graveyard by resolution
// stops the whole ability (CR 608.2b). The counter goes on the Predator
// only while it is the same object (CR 400.7).
//
// The activated ability is Unrooted Ancestor's without the {1}: the other
// creature is sacrificed as the cost, then indestructible, then the tap,
// which is itself a "becomes tapped" and triggers the exile.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fcb6c6d4-ea8d-49cf-b6df-377084bdb757",
		Name:            "Immersturm Predator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped,
					"Immersturm Predator — exile up to one target card from a graveyard and put a +1/+1 counter on this creature",
					Do(
						ExileChosenTarget{},
						// The counter waits for a prompt the exile asks (CR 608.2c).
						Instruction(plusOneCounterOnThisIfStillHere),
					)),
				TargetCardInGraveyard("up to one target card from a graveyard").WithCount(0, 1)),
		},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice another creature: This creature gains indestructible until end of turn. Tap it.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive | game.AnswerSacOutlet},
			Cost:    SacrificeAnotherN(1, "another creature", Creature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (GrantKeywordUntilEOT{
					Target:   item.SourceCardID,
					Keywords: []string{"indestructible"},
					Label:    "Immersturm Predator — indestructible until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
				return TapTarget{Target: item.SourceCardID}.Apply(ctx)
			},
		}},
	})
}
