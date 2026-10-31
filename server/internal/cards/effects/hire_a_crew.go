package effects

// Hire a Crew — {2}{R} Instant (Marvel Super Heroes, #4455):
//
//	"Create a 2/1 black Villain creature token with menace, then creatures
//	 you control get +1/+0 until end of turn. (A creature with menace can't
//	 be blocked except by two or more creatures.)"
//
// The pump is a one-shot effect over the creatures you control as the spell
// resolves (CR 611.2c): the new token is already on the battlefield by then
// and is pumped with the rest, and a creature that enters later this turn is
// not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6eb7a870-b626-4326-9e4c-45193f8ce138",
		Name:         "Hire a Crew",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			CreateToken{Template: TokenCard("2/1 black Villain with menace"), N: 1},
			BoostUntilEOT{
				Match: And(Creature(), YouControl()),
				Power: 1,
				Label: "Hire a Crew — creatures you control get +1/+0 until end of turn",
			},
		),
	})
}
