package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chrome Companion — Artifact Creature — Dog {2}, 2/1 (Edge of Eternities,
// #4461):
//
//	"Whenever this creature becomes tapped, you gain 1 life.
//	 {2}, {T}: Put target card from a graveyard on the bottom of its owner's
//	 library."
//
// "Becomes tapped" is Compassionate Healer's ThisBecameTapped (#4534): the
// tap event and the attack that taps it, so a vigilance attack is not one
// and any other tap is, its own {T} cost included (CR 603.2e). The activated
// ability is Sundering Archaic's: any graveyard's card goes to the bottom of
// its owner's library, and a card that left the graveyard in response is
// an illegal target (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1ac6fabb-71db-442f-8c14-ba9e659f906c",
		Name:         "Chrome Companion",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped,
				"Chrome Companion — you gain 1 life", Do(GainLife{Amount: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: Put target card from a graveyard on the bottom of its owner's library.",
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Targets: TargetCardInGraveyard("target card from a graveyard"),
			Effect:  socB37PutTargetGraveyardCardOnBottom,
		}},
	})
}
