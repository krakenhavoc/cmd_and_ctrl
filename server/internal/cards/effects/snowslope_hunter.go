package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Snowslope Hunter — Creature — Goblin Ranger {2}{R}, 2/3 (The Hobbit):
//
//	"Sacrifice another creature or artifact: Exile the top card of your
//	 library. You may play it until the end of your next turn. Activate
//	 only during your turn and only once each turn."
//
// The cost is Bartolomé del Presidio's sacrifice of "another creature or
// artifact" (Another is object identity, CR 109.1, so Snowslope Hunter
// cannot feed itself). The exile is Prosper's Mystic Arcanum's
// (b20ExileTopUntilEndOfNextTurn): "play", so a land is not stranded, and
// the duration is ADR 0063's "until the end of your next turn". The two
// timing limits are Vivi Ornitier's AllConditions pair, counted per ability
// announcement (CR 602.1b).
//
// No simplification.
func init() {
	const label = "Sacrifice another creature or artifact: Exile the top card of your library. You may play it until the end of your next turn. Activate only during your turn and only once each turn."
	Register(Spec{
		OracleID:     "fdf7f144-56e4-4f88-b81a-b85473922355",
		Name:         "Snowslope Hunter",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerSacOutlet},
			Condition: AllConditions(DuringYourTurn(), OncePerTurnActivation(label)),
			Cost: game.AbilityCost{
				SacrificeOther: Another(sacrificeSpec("another creature or artifact",
					Or(Creature(), Artifact()))),
			},
			Effect: b20ExileTopUntilEndOfNextTurn,
		}},
	})
}
