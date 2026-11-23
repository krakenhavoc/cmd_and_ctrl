package cardsfl

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
const FleshDuplicateVanishingGrant = "flesh-duplicate/vanishing-3"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "1869eb1e-7786-4fea-8a92-2bd839de47df",
		Name:         "Flesh Duplicate",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:       FleshDuplicateVanishingGrant,
			Text:      "Vanishing 3 (This permanent enters with three time counters on it. At the beginning of your upkeep, remove a time counter from it. When the last is removed, sacrifice it.)",
			Triggered: effects.Cmdr2B17VanishingTriggers(),
		}},
		Replacements: []game.ReplacementEffect{
			effects.EntersAsCopyOf("Flesh Duplicate", effects.AnyCreatureOnBattlefield,
				func(ev *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
					hasGrant := slices.Contains(v.GrantedAbilities, game.GrantKey(FleshDuplicateVanishingGrant))
					if !hasGrant && effects.S82Bns04PrintedVanishing(v) {
						return
					}
					if !hasGrant {
						v.GrantAbility(FleshDuplicateVanishingGrant)
					}
					ev.AddCounterAtETB(game.CounterTime, 3)
				}),
		},
	})
}
