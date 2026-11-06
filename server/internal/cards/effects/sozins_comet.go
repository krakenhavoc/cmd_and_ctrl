package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const sozinsCometFirebendingGrant = "sozins-comet/firebending-5"

// Sozin's Comet — Sorcery {3}{R}{R} (Avatar: The Last Airbender, #4459):
//
//	"Each creature you control gains firebending 5 until end of turn.
//	 (Whenever it attacks, add {R}{R}{R}{R}{R}. This mana lasts until
//	 end of combat.)
//	 Foretell {2}{R}"
//
// Firebending 5 is Avatar Roku's attack trigger with N = 5, granted as a
// catalog bundle until end of turn (Stir the Pride's shape, ADR 0093 PR
// 4), so it is each creature's own triggered ability and uses the stack
// (the rulings). The set is the creatures you control as the spell
// resolves (CR 611.2c): a creature that enters later does not gain it.
// A creature that already has firebending triggers once for each
// instance (the last ruling). Foretell is the keyword's (CR 702.143a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bc0da02c-2b85-4ddc-9ce0-63934dd14164",
		Name:         "Sozin's Comet",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: sozinsCometFirebendingGrant,
			Triggered: []game.TriggeredAbility{
				WheneverThisAttacks("Sozin's Comet — firebending 5", Do(firebending(5))),
			},
			Text: "Firebending 5 (Whenever this creature attacks, add {R}{R}{R}{R}{R}. This mana lasts until end of combat.)",
		}},
		SpecialActions: []game.SpecialAction{Foretell("{2}{R}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return GrantAbilitiesFor{
				Match: And(Creature(), YouControl()),
				Keys:  []string{sozinsCometFirebendingGrant},
				Label: "Sozin's Comet — firebending 5 until end of turn",
			}.Apply(ctx)
		},
	})
}
