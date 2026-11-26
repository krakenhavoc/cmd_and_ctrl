package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cyclops Gladiator — Creature — Cyclops Warrior {1}{R}{R}{R}, 4/4 (Magic
// 2011, #5579):
//
//	"Whenever this creature attacks, you may have it deal damage equal to its
//	 power to target creature defending player controls. If you do, that
//	 creature deals damage equal to its power to this creature."
//
// An attack trigger whose target is read off the creature it attacks
// (TargetCreatureDefendingPlayerControls, CR 508.1 / 506.3d), with a "you may"
// asked as it resolves (CR 603.5). On a yes the Gladiator deals damage equal
// to its power, then the target, if the Gladiator is still on the
// battlefield to be hurt, deals damage equal to its own power to the
// Gladiator. Both amounts are read before either hit lands. A target that
// became illegal leaves the whole ability without effect (CR 608.2b).
//
// No simplification.
func init() {
	attack := effects.Optional(
		effects.WheneverThisAttacks("Cyclops Gladiator — you may have it deal damage equal to its power to target creature defending player controls",
			S82OldB475GladiatorDuel),
		"Cyclops Gladiator — deal damage equal to its power to the target creature?")
	attack.TargetsFrom = effects.TargetCreatureDefendingPlayerControls
	effects.Register(effects.Spec{
		OracleID:     "ca5711d1-8cc9-4540-a567-778b56d51f50",
		Name:         "Cyclops Gladiator",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{attack},
	})
}

// S82OldB475GladiatorDuel is the "yes" body: the source deals damage equal
// to its power to the first target still legal, then that creature deals
// damage equal to its power to the source while the source is still on the
// battlefield.
func S82OldB475GladiatorDuel(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	target, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	mine := effects.SourcePowerNow(ctx)
	theirs := effects.B31CurrentPowerOf(g, target)
	if err := (effects.DealDamage{Source: item.SourceCardID, Target: target, Amount: mine}).Apply(ctx); err != nil {
		return err
	}
	if !effects.OnBattlefield(g, item.SourceCardID) {
		return nil
	}
	return effects.DealDamage{Source: target, Target: item.SourceCardID, Amount: theirs}.Apply(ctx)
}
