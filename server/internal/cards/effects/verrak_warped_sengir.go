package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Verrak, Warped Sengir — Legendary Creature — Vampire {1}{W}{B}, 2/2:
//
//	"Flying, deathtouch, lifelink
//	 Whenever you activate an ability that isn't a mana ability, if
//	 life was paid to activate it, you may pay that much life again.
//	 If you do, copy that ability. You may choose new targets for the
//	 copy."
//
// The activation event carries the life the announcement paid (#5662,
// game.Event.LifePaid): a printed life cost and the life paid for
// Phyrexian mana symbols (CR 107.4f) in one total, which is both the
// intervening "if" (CR 603.4) and "that much life". A mana ability
// emits a different event and never triggers this (CR 605.3b).
//
// The second payment is made as the trigger resolves and is not an
// activation; the copy is created, not activated (CR 707.10), so
// neither triggers Verrak again. The copy is Verrak's controller's and
// they may choose new targets for it (CR 707.10c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "04853f06-3fd2-445d-9b0d-d68d0f116635",
		Name:            "Verrak, Warped Sengir",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch", "lifelink"},
		Triggered: []game.TriggeredAbility{
			InterveningIf(LifeWasPaidToActivateIt,
				On(game.EventActivateAbility, YouActivatedAnAbilityPayingLife,
					"Verrak, Warped Sengir — pay that much life again to copy that ability",
					Do(MayPayThatLifeAgainToCopyIt))),
		},
	})
}
