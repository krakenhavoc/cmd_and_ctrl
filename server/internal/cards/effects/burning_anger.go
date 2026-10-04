package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const burningAngerGrant = "burning-anger/ping"

// Burning Anger — Enchantment — Aura {4}{R}:
//
//	"Enchant creature
//	 Enchanted creature has "{T}: This creature deals damage equal to
//	 its power to any target.""
//
// Samite Blessing's shape (ADR 0093, ADR 0107 §6): the granted row is
// the enchanted creature's own ability, so it taps that creature, is
// activated by that creature's controller, and "this creature" is the
// enchanted creature itself. The damage is the creature's, as
// printed, read at resolution (a pump that resolved in response
// counts; a creature that has left deals its last-known power,
// CR 608.2h). A summoning-sick creature can't use the {T} ability,
// which is the engine's ordinary rule for a granted {T} cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a2c76ac2-5ab1-450c-b64e-e9a6c4dad1ca",
		Name:         "Burning Anger",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Grants: []AbilityGrant{{
			Key: burningAngerGrant,
			Activated: []ActivatedAbility{{
				Label:   "{T}: This creature deals damage equal to its power to any target.",
				Cost:    TapCost(),
				Targets: TargetAny(),
				Effect:  sourceDealsDamageEqualToItsPowerToTarget,
			}},
			Text: "{T}: This creature deals damage equal to its power to any target.",
		}},
		Static: []game.StaticAbility{GrantAbilitiesToAttached(burningAngerGrant)},
	})
}

// sourceDealsDamageEqualToItsPowerToTarget is "<this> deals damage
// equal to its power to <its target>": the ability's source, with its
// power read now or, if it has left the battlefield, as it last was.
func sourceDealsDamageEqualToItsPowerToTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	power := b43PowerNowOrLastKnown(g, item.SourceCardID)
	if power <= 0 {
		return nil
	}
	targets := ctx.LegalTargets()
	if len(targets) == 0 {
		return nil
	}
	return DealDamage{Source: item.SourceCardID, Target: targets[0].ID, Amount: power}.Apply(ctx)
}
