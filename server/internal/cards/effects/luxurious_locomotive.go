package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Luxurious Locomotive — {5} Artifact — Vehicle, 6/5:
//
//	"Whenever this Vehicle attacks, create a Treasure token for each
//	 creature that crewed it this turn.
//	 Crew 1. Activate only once each turn."
//
// The count is the number of DISTINCT creature objects in the Locomotive's
// crew records this turn (ADR 0186 owner decision 4), departed ones
// included: a creature that crewed it and was then sacrificed or bounced
// still crewed it, and one tapped for two crew activations is still one
// creature. The records are read by the Locomotive's object, so a
// Locomotive that left in response still counts. Crew is limited to once
// each turn by the activation tally (CR 602.1b), so a second crew attempt
// is never offered.
//
// No simplification.
func init() {
	const crewLabel = "Crew 1. Activate only once each turn."
	Register(Spec{
		OracleID:     "c68c2a98-9a6d-4199-af3a-78eb92beda48",
		Name:         "Luxurious Locomotive",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     crewLabel,
			Cost:      CrewCost(1),
			Condition: OncePerTurnActivation(crewLabel),
			Purpose:   game.Purpose{Answers: game.AnswerAnimate},
			Effect:    CrewEffect("Luxurious Locomotive"),
		}},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Luxurious Locomotive — create a Treasure for each creature that crewed it", luxuriousLocomotiveTreasures),
		},
	})
}

func luxuriousLocomotiveTreasures(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	seen := map[game.ObjectRef]bool{}
	for _, r := range g.CrewRecordsForEffect(ref, false) {
		seen[r.Creature] = true
	}
	if len(seen) == 0 {
		return nil
	}
	return CreateToken{Template: TreasureToken(), N: len(seen)}.Apply(ctx)
}
