package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shivan Meteor — Sorcery {3}{R}{R}:
//
//	"Shivan Meteor deals 13 damage to target creature.
//	 Suspend 2—{1}{R}{R} (Rather than cast this card from your hand, you may
//	 pay {1}{R}{R} and exile it with two time counters on it. At the
//	 beginning of your upkeep, remove a time counter. When the last is
//	 removed, you may cast it without paying its mana cost.)"
//
// Rift Bolt's shape with a creature clause and a bigger number: a damage
// body and the one-line `Suspend(2, "{1}{R}{R}")` declaration. Everything
// about the keyword is the engine's (suspend.go, CR 702.62): the special
// action, which a sorcery may take only at sorcery speed (CR 702.62c), the
// two time counters, the upkeep countdown, the free cast when the last
// counter comes off and the haste a suspended creature would get. The
// target is chosen when the spell is cast from exile on that upkeep, not
// when it was suspended, so a creature that arrives meanwhile is a legal
// pick.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6cf5edd6-1ca2-46a5-83b4-7ddff590ef96",
		Name:         "Shivan Meteor",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(DamageToTarget(0, 13)),
		SpecialActions: []game.SpecialAction{
			Suspend(2, "{1}{R}{R}"),
		},
		OnResolve: damageToFirstTarget(13),
	})
}
