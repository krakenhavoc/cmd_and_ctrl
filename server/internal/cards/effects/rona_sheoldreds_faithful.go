package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rona, Sheoldred's Faithful — Legendary Creature — Human Wizard
// {1}{U}{B}{B}, 3/4 (Dominaria United, #5487):
//
//	"Whenever you cast an instant or sorcery spell, each opponent loses 1
//	 life.
//	 You may cast this card from your graveyard by discarding two cards in
//	 addition to paying its other costs."
//
// The trigger is a cast trigger over instants and sorceries draining each
// opponent for 1. The graveyard cast is Retrace's model (CR 702.81):
// the printed mana cost plus a discard of two cards as an additional cost,
// carried on a graveyard-bound offer (CastFromGraveyardByDiscarding). A
// graveyard with Rona in it and fewer than two other cards in hand cannot
// pay it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "52e92296-7af9-4e68-8045-40b6d6b73910",
		Name:             "Rona, Sheoldred's Faithful",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{CastFromGraveyardByDiscarding("{1}{U}{B}{B}", 2)},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Or(Instant(), Sorcery()),
				"Rona, Sheoldred's Faithful — each opponent loses 1 life",
				Do(EachOpponentLosesLife{N: 1})),
		},
	})
}
