package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b245_helpers.go — shared pieces of Amonkhet batch 245 (#5535).

// CantAttackUnlessYouHaveCardsInGraveyard is "This creature can't attack
// unless there are N or more cards in your graveyard" (River Serpent): the
// attack half of the sentence, a fact about the attacking creature's
// controller read live at declaration (CR 508.1c).
func CantAttackUnlessYouHaveCardsInGraveyard(n int) game.StaticAbility {
	s := selfAttackTargetRestriction(game.AttackTargetRestriction{ControllerGraveyardAtLeast: n})
	s.Label = "Can't attack unless there are " + NumberWord(n) + " or more cards in your graveyard."
	return s
}
