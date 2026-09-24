package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sting, the Glinting Dagger — Legendary Artifact — Equipment {2}:
//
//	"Equipped creature gets +1/+1 and has haste.
//	 At the beginning of each combat, untap equipped creature.
//	 Equipped creature has first strike as long as it's blocking or
//	 blocked by a Goblin or Orc.
//	 Equip {2}"
//
// The untap trigger is AtBeginningOfEachCombat (any player's turn,
// unlike AtBeginningOfYourCombat) — the effect reads the current host
// off the Dagger itself via attachedHostFor rather than off anything
// captured at Build, since which creature is equipped can change
// between combats.
//
// SIMPLIFICATION, DECLARED: the conditional first strike is not
// implemented. "As long as it's blocking or blocked by a Goblin or
// Orc" is a condition on live block-declaration state
// (Card.BlockingTarget, set at CR 509), and the layer engine has no
// invalidation hook for that state changing — StaticAbility carries
// DependsOnLifeTotal, DependsOnHandSize, DependsOnAttackingStatus and
// DependsOnSpellsCast (layer_listener.go), and none of those fire when
// a block is declared or locked in. A static built on
// Card.BlockingTarget without that hook would read whatever the layer
// cache last computed — right only when some UNRELATED event happens
// to force a recompute in the same window — which is a coin flip, not
// an implementation. Left out until a DependsOnBlockStatus-shaped hook
// exists.
func init() {
	Register(Spec{
		OracleID:     "973c49a1-8425-40b9-8bdc-2c2434222314",
		Name:         "Sting, the Glinting Dagger",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The equipped creature doesn't gain first strike from blocking or being blocked by a Goblin or Orc.",
		},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached("haste"),
		},
		Triggered: []game.TriggeredAbility{
			AtBeginningOfEachCombat("Sting, the Glinting Dagger — untap equipped creature", stingUntapEquippedCreature),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}

func stingUntapEquippedCreature(g *game.Game, item *game.StackItem) error {
	host := attachedHostFor(g, item.SourceCardID)
	if host == nil {
		return nil
	}
	return UntapTarget{Target: host.InstanceID}.Apply(NewContext(g, item))
}
