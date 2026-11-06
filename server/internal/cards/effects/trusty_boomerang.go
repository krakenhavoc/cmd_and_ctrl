package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Trusty Boomerang — Artifact — Equipment {1} (Avatar: The Last
// Airbender, #4459):
//
//	"Equipped creature has "{1}, {T}: Tap target creature. Return Trusty
//	 Boomerang to its owner's hand."
//	 Equip {1}"
//
// The quoted ability is an ADR 0093 grant (Thornbite Staff's shape): it is
// the EQUIPPED creature's own activated ability, so the creature taps to
// pay for it, CR 302.6 keeps a creature that came under its controller's
// control this turn from using it, and moving the Boomerang moves it.
// "Trusty Boomerang" in the granted text is the Equipment that granted it,
// which the activation names (StackItem.GrantedBy, stamped as the ability
// is activated): the target is tapped, then that Boomerang returns to its
// owner's hand if it is still on the battlefield. If the target has become
// illegal the ability does nothing at all (CR 608.2b), so the Boomerang
// stays.
//
// No simplification.
const trustyBoomerangGrant = "trusty-boomerang/tap-and-return"

func init() {
	Register(Spec{
		OracleID:     "3c032e74-69c6-48ca-95d4-51aa3d894c14",
		Name:         "Trusty Boomerang",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: trustyBoomerangGrant,
			Activated: []ActivatedAbility{{
				Label:   "{1}, {T}: Tap target creature. Return Trusty Boomerang to its owner's hand.",
				Cost:    Plus(ManaCost("{1}"), TapCost()),
				Targets: TargetCreature("target creature"),
				Effect:  trustyBoomerangThrow,
			}},
			Text: "{1}, {T}: Tap target creature. Return Trusty Boomerang to its owner's hand.",
		}},
		Static:    []game.StaticAbility{GrantAbilitiesToAttached(trustyBoomerangGrant)},
		Activated: []ActivatedAbility{EquipAbility("{1}")},
	})
}

// trustyBoomerangThrow taps the target, then returns the Boomerang that
// granted the ability.
func trustyBoomerangThrow(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target := FirstLegalBattlefieldTarget(ctx)
	if target == uuid.Nil {
		return nil
	}
	if err := (TapTarget{Target: target}).Apply(ctx); err != nil {
		return err
	}
	boomerang := ctx.Grantor()
	if boomerang == uuid.Nil || !onBattlefield(g, boomerang) {
		return nil
	}
	return BounceToHand{Target: boomerang}.Apply(ctx)
}
