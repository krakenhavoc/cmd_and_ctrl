package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const rootManipulationGrant = "root-manipulation/attack-gain-life"

// Root Manipulation — Sorcery {3}{B}{G} (Secrets of Strixhaven, #4456):
//
//	"Until end of turn, creatures you control get +2/+2 and gain menace and
//	 "Whenever this creature attacks, you gain 1 life.""
//
// One sentence, one effect: the creatures you control as the spell resolves
// are locked in (CR 611.2c) and each gets the pump, menace and the quoted
// trigger together (GrantAbilitiesFor, Stir the Pride's shape). The trigger
// is each creature's own ability, so "you" is that creature's controller.
// All of it ends at cleanup (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d870948e-721a-46fd-86b6-ecd42e3aa9bd",
		Name:         "Root Manipulation",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: rootManipulationGrant,
			Triggered: []game.TriggeredAbility{
				WheneverThisAttacks("Root Manipulation — you gain 1 life", Do(GainLife{Amount: 1})),
			},
			Text: "Whenever this creature attacks, you gain 1 life.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return GrantAbilitiesFor{
				Match: And(Creature(), YouControl()),
				Keys:  []string{rootManipulationGrant},
				Also:  []game.Mod{game.ModifyPTMod(2, 2), game.AddKeywordsMod("menace")},
				Label: "Root Manipulation — +2/+2, menace and \"whenever this creature attacks, you gain 1 life\"",
			}.Apply(ctx)
		},
	})
}
