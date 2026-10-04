package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Anzrag, the Quake-Mole — Legendary Creature — Mole God {2}{R}{G},
// 8/4:
//
//	"Whenever Anzrag becomes blocked, untap each creature you control.
//	 After this phase, there is an additional combat phase.
//	 {3}{R}{R}{G}{G}: Anzrag must be blocked each combat this turn if
//	 able."
//
// "Becomes blocked" is EventBecomesBlocked, emitted once per blocked
// attacker when the block declaration is complete (CR 509.1h), so
// several blockers still trigger it once (the 2024-02-02 ruling). The
// body is Aurelia's: untap all creatures you control, then an
// additional combat phase after this one (CR 500.8). It can trigger
// again in that combat and add another (the ruling).
//
// The activated ability is a must-be-blocked requirement on Anzrag for
// the rest of the turn, a data record pinned to this object (CR
// 400.7) that lasts until the cleanup step, so every declare blockers
// step that turn reads it. Only one blocker is required; a defender with no
// creature able to block, or one that would have to pay a cost to
// block, isn't forced to (CR 509.1c, the rulings).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4adcd967-9ff7-4940-b8b9-0c4215bbcb75",
		Name:         "Anzrag, the Quake-Mole",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBecomesBlocked, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			}, "Anzrag, the Quake-Mole — untap each creature you control, additional combat",
				untapAllYouControlThenExtraCombat),
		},
		Activated: []ActivatedAbility{{
			Label: "{3}{R}{R}{G}{G}: Anzrag must be blocked each combat this turn if able.",
			Cost:  ManaCost("{3}{R}{R}{G}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				return BlockRequirementUntilEOT{
					Target: item.SourceCardID,
					Kind:   game.BlockRequirementMustBeBlocked,
					Label:  "Anzrag, the Quake-Mole — must be blocked each combat this turn",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
