package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "8a56c2a7-7131-4495-aa27-52f48e5f1fb6",
		Name:         "Prowler, Misguided Mentor",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.PowerLE(2), "creatures with power 2 or less"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisDealsCombatDamageToAPlayer(
				"Prowler, Misguided Mentor — put a +1/+1 counter on another target creature you control",
				effects.PlusOneCounterOnTheTarget),
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
		},
	})
}
