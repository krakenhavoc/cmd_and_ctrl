package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Warlock Class — Enchantment — Class {B} (Adventures in the Forgotten
// Realms):
//
//	"(Gain the next level as a sorcery to add its ability.)
//	 At the beginning of your end step, if a creature died this turn, each
//	 opponent loses 1 life.
//	 {1}{B}: Level 2
//	 When this Class becomes level 2, look at the top three cards of your
//	 library. Put one of them into your hand and the rest into your
//	 graveyard.
//	 {6}{B}: Level 3
//	 At the beginning of your end step, each opponent loses life equal to
//	 the life they lost this turn. (Damage causes loss of life.)"
//
// One line per level (ADR 0071):
//
//   - Level 1 is an end-step trigger with an intervening "if" (CR 603.4)
//     read off the game-wide turn tally: any player's creature, a token
//     included (CondAnyCreatureDied), checked as the step begins and again
//     as it resolves.
//   - Level 2 is BecomesLevel's trigger with Codecracker Hound's look at
//     the top of the library and Wail of the Forgotten's rest to the
//     graveyard: exactly one card goes to the hand, the rest are put (not
//     discarded) into the graveyard.
//   - Level 3 is Archfiend of Despair's end-step drain, gated at level 3.
//     Each opponent's loss is the life they lost this turn, read
//     for every opponent before any is applied, so it includes the level 1
//     trigger's loss that resolved earlier in the same end step.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "766c2c67-d1ee-44ff-b9bb-355361646ef1",
		Name:         "Warlock Class",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			LevelUp(2, ManaCost("{1}{B}")),
			LevelUp(3, ManaCost("{6}{B}")),
		},
		Triggered: []game.TriggeredAbility{
			AtYourEndStepIfAnyCreatureDied("Warlock Class — each opponent loses 1 life",
				Do(EachOpponentLosesLife{N: 1})),
			BecomesLevel(2, "Warlock Class — level 2: look at the top three cards of your library, put one into your hand and the rest into your graveyard",
				warlockClassLook),
			AtLevel(3, AtYourEndStep("Warlock Class — each opponent loses life equal to the life they lost this turn",
				eachOpponentLosesTheLifeTheyLostThisTurn)),
		},
	})
}

func warlockClassLook(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := ctx.Controller()
	return TakeFromLibraryToHand{
		Player: player,
		Cards:  g.LookAtTopOfLibraryForEffect(player, 3),
		Max:    1,
		Label:  "Warlock Class — put one of them into your hand and the rest into your graveyard",
		Key:    takeRestGraveyard,
	}.Apply(ctx)
}
