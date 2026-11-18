package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Take to the Streets — Sorcery {4}{G} (Streets of New Capenna, #5491):
//
//	"Creatures you control get +2/+2 until end of turn. Citizens you control
//	 get an additional +1/+1 and gain vigilance until end of turn."
//
// Overrun's BoostUntilEOT over the creatures you control as it resolves, then
// the Citizens among them get the additional +1/+1 and vigilance. Both sets
// are fixed at resolution (CR 611.2c), so a creature that arrives later gets
// neither. "Citizen" is a word of the card's text (CR 612.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b3606b80-8d1b-4417-8902-a7ccf6b97cbb",
		Name:         "Take to the Streets",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (BoostUntilEOT{
				Match: And(Creature(), YouControl()), Power: 2, Toughness: 2,
				Label: "Take to the Streets — +2/+2",
			}).Apply(ctx); err != nil {
				return err
			}
			citizens := And(Creature(), YouControl(), HasSubtype(ctx.TextSubtype("Citizen")))
			if err := (BoostUntilEOT{
				Match: citizens, Power: 1, Toughness: 1,
				Label: "Take to the Streets — Citizens get an additional +1/+1",
			}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{
				Match: citizens, Keywords: []string{"vigilance"},
				Label: "Take to the Streets — Citizens gain vigilance",
			}.Apply(ctx)
		},
	})
}
