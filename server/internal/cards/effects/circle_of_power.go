package effects

// Circle of Power — Sorcery {3}{B}:
//
//	"You draw two cards and you lose 2 life. Create a 0/1 black Wizard
//	 creature token with "Whenever you cast a noncreature spell, this
//	 token deals 1 damage to each opponent."
//	 Wizards you control get +1/+0 and gain lifelink until end of turn."
//
// The Wizard is the shared noncreature-cast token Mysidian Elder and
// Transpose make, so its trigger is the real one. The last sentence is
// a one-shot effect over the Wizards you control AT RESOLUTION (CR
// 611.2c): the new token is already on the battlefield by then and is
// pumped with the rest, and a Wizard that enters later is not. +1/+0
// and lifelink sit in different layers (7c and 6), so they are two
// turn-scoped effects. (#3213, the Kuja deck.)
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2d41fd1f-4a98-45dd-a3c5-f7c5e0152109",
		Name:         "Circle of Power",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			// Each instruction waits for a prompt the one before asks (CR 608.2c, #4155).
			DrawCards{N: 2},
			Instruction(func(ctx *Context) error {
				return ctx.Game.ChangePlayerLifeForEffect(ctx.Item.SourceCardID, ctx.Controller(), -2)
			}),
			CreateToken{Template: NoncreatureCastWizardToken(), N: 1},
			BoostUntilEOT{
				Match: And(OfCreatureType("Wizard"), YouControl()),
				Power: 1,
				Label: "Circle of Power — Wizards get +1/+0 until end of turn",
			},
			GrantKeywordUntilEOT{
				Match:    And(OfCreatureType("Wizard"), YouControl()),
				Keywords: []string{"lifelink"},
				Label:    "Circle of Power — Wizards gain lifelink until end of turn",
			},
		),
	})
}
