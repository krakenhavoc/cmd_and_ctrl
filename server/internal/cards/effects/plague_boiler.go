package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Plague Boiler — Artifact {3}:
//
//	"At the beginning of your upkeep, put a plague counter on this
//	 artifact.
//	 {1}{B}{G}: Put a plague counter on this artifact or remove a plague
//	 counter from it.
//	 When this artifact has three or more plague counters on it,
//	 sacrifice it. If you do, destroy all nonland permanents."
//
// ADR 0107 §1 (#1858).
//
//   - "Put … or remove" is a choice made as the ability resolves (CR
//     608.2d), not a mode: the activator is asked then, and "remove" is
//     offered only while there is a counter to remove.
//   - The wipe is a CR 603.8 state trigger. It triggers the moment the
//     third counter lands, and a fourth while it waits does nothing more;
//     removing one in response does not stop it (the condition is not an
//     intervening "if"). The destruction happens only if the Boiler was
//     really sacrificed, and every nonland permanent goes at once.
//
// No simplification.
func init() {
	const plague = "plague"
	Register(Spec{
		OracleID:     "fef502af-6e79-4c55-a86a-b45adb3fc64a",
		Name:         "Plague Boiler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Plague Boiler — put a plague counter", putACounterOnThis(plague)),
			TriggerWithPurpose(WhenThisHasAtLeast(plague, 3, "Plague Boiler — sacrifice it and destroy all nonland permanents",
				SacrificeThisThen(func(ctx *Context) error {
					return DestroyAllMatching{Match: Nonland()}.Apply(ctx)
				})), game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy}}),
		},
		Activated: []ActivatedAbility{{
			Label: "{1}{B}{G}: Put a plague counter on this artifact or remove a plague counter from it.",
			Cost:  ManaCost("{1}{B}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				c, ok := g.LookupCardForEffect(item.SourceCardID)
				if !ok || !onBattlefield(g, item.SourceCardID) {
					return nil
				}
				options := []game.ChoiceOption{{Label: "Put a plague counter on Plague Boiler"}}
				if c.Counters[plague] > 0 {
					options = append(options, game.ChoiceOption{Label: "Remove a plague counter from Plague Boiler"})
				}
				return PickOption{
					Question: "Plague Boiler — put a plague counter on it or remove one?",
					Options:  options,
					Then: func(ctx *Context, index int) error {
						switch index {
						case 0:
							return putACounterOnThis(plague)(ctx.Game, ctx.Item)
						case 1:
							return removeACounterFromThis(plague)(ctx.Game, ctx.Item)
						}
						return nil
					},
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
