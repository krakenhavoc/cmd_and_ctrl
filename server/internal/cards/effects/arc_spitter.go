package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// arcSpitterGrant is the granted ability's key.
const arcSpitterGrant = "arc-spitter/ping-a-blocker"

// Arc Spitter — Artifact — Equipment {1} (Streets of New Capenna, #5491):
//
//	"Equipped creature has "{1}: This creature deals 1 damage to target
//	 creature that's blocking it."
//	 Equip {1}"
//
// An ADR 0093 grant (Burning Anger's shape): the quoted ability is the
// equipped creature's own, so "this creature" is the creature that has it
// and the damage is its own, 1 point from it. The target is a creature
// blocking the ability's source (BlockingSource, CR 509.1 / CR 506.4), so
// it is illegal if the equipped creature is not attacking or that blocker
// has left combat, in which case the ability does nothing (CR 608.2b).
// The ability has no {T} in its cost, so it can be activated any number of
// times, and a creature that was just put under its controller's control
// may use it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9376ccf6-1b18-44e9-8f9d-c5ad38927e2f",
		Name:         "Arc Spitter",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: arcSpitterGrant,
			Activated: []ActivatedAbility{{
				Label:   "{1}: This creature deals 1 damage to target creature that's blocking it.",
				Cost:    ManaCost("{1}"),
				Targets: BlockingSource(TargetCreature("target creature that's blocking it")),
				Effect:  ArcSpitterPing,
			}},
			Text: "{1}: This creature deals 1 damage to target creature that's blocking it.",
		}},
		Static:    []game.StaticAbility{GrantAbilitiesToAttached(arcSpitterGrant)},
		Activated: []ActivatedAbility{EquipAbility("{1}")},
	})
}

func ArcSpitterPing(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target := FirstLegalBattlefieldTarget(ctx)
	if target == uuid.Nil {
		return nil
	}
	return DealDamage{Source: item.SourceCardID, Target: target, Amount: 1}.Apply(ctx)
}
