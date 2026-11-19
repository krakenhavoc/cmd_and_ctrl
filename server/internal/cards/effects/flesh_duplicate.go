package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Flesh Duplicate — Creature — Shapeshifter Rebel {U}{U}, 0/0:
//
//	"You may have this creature enter as a copy of any creature on the
//	 battlefield, except it has vanishing 3 if that creature doesn't have
//	 vanishing. (A permanent with vanishing 3 enters with three time counters
//	 on it. At the beginning of your upkeep, remove a time counter from it.
//	 When the last is removed, sacrifice it.)"
//
// Clone's entry copy with an except clause that grants vanishing 3 (CR 707.9a,
// Phantasmal Image's shape): the two vanishing triggers are a grant bundle the
// copy carries by name, so a copy of the copy has them too, and the entry
// counters ride the entry event as Altered Ego's X counters do, going through
// the counter window. A copied creature that already has vanishing (any card
// whose upkeep trigger is the vanishing row) keeps its own and enters with its
// own counters, so nothing is added. Copying a creature that is itself a Flesh
// Duplicate copy gives vanishing 3 through the grant, and the copy enters with
// three counters again, since the counters are part of vanishing, not of the
// grant bundle. Declining the copy leaves a 0/0 that dies as printed.
//
// No simplification.
const fleshDuplicateVanishingGrant = "flesh-duplicate/vanishing-3"

func init() {
	Register(Spec{
		OracleID:     "1869eb1e-7786-4fea-8a92-2bd839de47df",
		Name:         "Flesh Duplicate",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       fleshDuplicateVanishingGrant,
			Text:      "Vanishing 3 (This permanent enters with three time counters on it. At the beginning of your upkeep, remove a time counter from it. When the last is removed, sacrifice it.)",
			Triggered: cmdr2B17VanishingTriggers(),
		}},
		Replacements: []game.ReplacementEffect{
			EntersAsCopyOf("Flesh Duplicate", anyCreatureOnBattlefield,
				func(ev *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
					hasGrant := slices.Contains(v.GrantedAbilities, game.GrantKey(fleshDuplicateVanishingGrant))
					if !hasGrant && s82Bns04PrintedVanishing(v) {
						return
					}
					if !hasGrant {
						v.GrantAbility(fleshDuplicateVanishingGrant)
					}
					ev.AddCounterAtETB(game.CounterTime, 3)
				}),
		},
	})
}
