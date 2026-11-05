package effects

// Warship Flight Crew — Creature — Klingon Pilot {1}{R}, 2/2 (Star Trek
// Commander, #4974):
//
//	(no rules text)
//
// A vanilla body is still registered so it resolves, enters and attacks
// without the manual sandbox path; its cost, types and P/T ride the card
// data from Scryfall.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d55977f8-1655-4d97-adbf-ef6cb979ea55",
		Name:         "Warship Flight Crew",
		Completeness: CompletenessFull,
	})
}
