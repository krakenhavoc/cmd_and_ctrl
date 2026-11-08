package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Imodane's Recruiter // Train Troops — Creature — Human Knight {2}{R},
// 2/2 // Sorcery — Adventure {4}{W} (Wilds of Eldraine, #4471):
//
//	Imodane's Recruiter — "When this creature enters, creatures you
//	 control get +1/+0 and gain haste until end of turn."
//	Train Troops — "Create two 2/2 white Knight creature tokens with
//	 vigilance. (Then exile this card. You may cast the creature later
//	 from exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID
// and "#1"). The entry trigger is Scurry of Gremlins' team pump: the set
// of creatures is fixed as it resolves (CR 611.2c), the Recruiter
// included, so a creature that enters afterwards gets neither. Train
// Troops makes the token table's vigilant Knight twice. The Adventure
// lifecycle is the engine's (CR 715.3d).
//
// No simplification.
const imodanesRecruiterOracleID = "50bee80e-89df-48c2-b608-ba66ff3e794a"

func init() {
	Register(Spec{
		OracleID:     imodanesRecruiterOracleID,
		Name:         "Imodane's Recruiter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Imodane's Recruiter — creatures you control get +1/+0 and gain haste until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					yours := And(Creature(), YouControl())
					if err := (BoostUntilEOT{Match: yours, Power: 1, Label: "Imodane's Recruiter — +1/+0"}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{Match: yours, Keywords: []string{"haste"}, Label: "Imodane's Recruiter — haste"}.Apply(ctx)
				}),
		},
	})
	Register(Spec{
		OracleID:     imodanesRecruiterOracleID + "#1",
		Name:         "Train Troops",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Template: TokenCard("2/2 white Knight with vigilance"), N: 2}.Apply(ctx)
		},
	})
}
