package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hollow Warrior — Artifact Creature — Golem Warrior {4}, 4/4 (Prophecy, S82 OLD
// batch 676, #5617):
//
//	"This creature can't attack or block unless you tap an untapped creature you
//	 control not declared as an attacking or blocking creature this combat. (This
//	 cost is paid as attackers or blockers are declared.)"
//
// The #6190 seam, in both halves: a cost paid in permanents as attackers are
// declared (CR 508.1h) and again as blockers are declared (CR 509.1d), the
// creature named by the player and tapped. It is tapped as a cost, not with the
// {T} symbol, so a creature that arrived this turn may pay (CR 302.6,
// 602.5a). The creature must be untapped and not declared as an attacker or a
// blocker this combat (NotInCombatThisCombat, which reads the declaration
// records, so a creature that was declared and has since left combat can't
// pay). The Warrior can't pay with itself (ExcludeSource), being declared
// itself. Declining the cost
// is not attacking or blocking at all.
//
// No simplification.
func init() {
	const label = "an untapped creature you control not declared as an attacking or blocking creature this combat"
	preds := []effects.CardPredicate{effects.NotInCombatThisCombat(), effects.Creature(), effects.Untapped()}
	// The Warrior is itself being declared, so it can't pay its own cost.
	attack := effects.CantAttackUnlessYouTap(label, preds...)
	attack.NonMana.Tap.ExcludeSource = true
	block := effects.CantBlockUnlessYouTap(label, preds...)
	block.NonMana.Tap.ExcludeSource = true
	effects.Register(effects.Spec{
		OracleID:     "18ca77d9-4740-4be4-913a-e21c9ed9e43f",
		Name:         "Hollow Warrior",
		Completeness: effects.CompletenessFull,
		AttackTaxes:  []game.AttackTax{attack},
		BlockTaxes:   []game.BlockTax{block},
	})
}
