package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prowler, Misguided Mentor — Legendary Creature — Human Rogue Villain
// {2}{G}, 3/3 (Marvel's Spider-Man, #4460):
//
//	"Prowler can't be blocked by creatures with power 2 or less.
//	 Whenever Prowler deals combat damage to a player, put a +1/+1
//	 counter on another target creature you control."
//
// The block restriction is Legolas Greenleaf's, read live when blockers
// are declared (CR 509.1b). The trigger's target is another creature, so a
// self-pick is refused; it is chosen as the trigger goes on the stack and
// re-checked as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8a56c2a7-7131-4495-aa27-52f48e5f1fb6",
		Name:         "Prowler, Misguided Mentor",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnSelf(), PowerLE(2), "creatures with power 2 or less"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisDealsCombatDamageToAPlayer(
				"Prowler, Misguided Mentor — put a +1/+1 counter on another target creature you control",
				PlusOneCounterOnTheTarget),
				Another(TargetCreature("another target creature you control", YouControl()))),
		},
	})
}
