package effects

// Quark's Bar — Land (Star Trek, #4973):
//
//	"This land enters tapped unless you control two or more other lands.
//	 When this land enters untapped, choose one —
//	 • This land deals 1 damage to each opponent.
//	 • Surveil 1.
//	 {T}: Add {R}."
//
// The red member of Star Trek's surveil-land cycle (trekB29SurveilLand).
// The damage comes from the land, by last known information if it has
// left by then (CR 608.2h).
//
// No simplification.
func init() {
	trekB29SurveilLand("800d0f33-31c9-43da-a649-8860646eca1f", "Quark's Bar", "R",
		ModeSteps("This land deals 1 damage to each opponent.", nil, DamageEachOpponent{N: 1}))
}
