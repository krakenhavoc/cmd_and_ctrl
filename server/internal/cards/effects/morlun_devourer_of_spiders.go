package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Morlun, Devourer of Spiders — Legendary Creature — Vampire Villain
// {X}{B}{B}, 2/1 (Marvel's Spider-Man, #4460):
//
//	"Lifelink
//	 Morlun enters with X +1/+1 counters on him.
//	 When Morlun enters, he deals X damage to target opponent."
//
// X is the value announced for the spell (CR 107.3m). The counters are the
// cast-time entry counters (Feral Hydra's XCounters), so a counter doubler
// applies. The trigger reads X off the permanent as it is built and carries
// it on the item (Ajani's Anguish), so it is exact even if Morlun has left
// by the time it resolves; a copy or a cheat onto the battlefield has X = 0
// and deals no damage. The damage is Morlun's, so lifelink gains the life.
//
// No simplification.
func init() {
	const label = "Morlun, Devourer of Spiders — he deals X damage to target opponent"
	Register(Spec{
		OracleID:                   "273ad15c-366a-4691-978b-d6028b57005a",
		Name:                       "Morlun, Devourer of Spiders",
		Completeness:               CompletenessFull,
		PrintedKeywords:            []string{"lifelink"},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters(game.CounterPlusOne)},
		XMatters:                   true,
		Triggered:                  []game.TriggeredAbility{etbDealsCastXDamage(label, TargetPlayer("target opponent", Opponent()))},
	})
}
