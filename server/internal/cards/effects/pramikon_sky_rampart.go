package effects

// Pramikon, Sky Rampart — Legendary Creature — Wall {U}{R}{W}, 1/5:
//
//	"Flying, defender
//	 As Pramikon enters, choose left or right.
//	 Each player may attack only the nearest opponent in the chosen
//	 direction and planeswalkers controlled by that opponent."
//
// Flying and defender are printed keywords. Pramikon's controller
// chooses left or right as it enters (CR 614.12), and while it is on the
// battlefield every player, its controller included, may attack only
// their nearest opponent that way and planeswalkers that opponent
// controls (ADR 0206). Left is the next player in turn order; a seat
// whose player has left the game is skipped. Battles can't be attacked,
// since they are not listed. A creature put onto the battlefield
// attacking is not affected (CR 508.4c). Two Pramikons (or Pramikon and
// Teyo, Geometric Tactician) pointing different ways leave a player in a
// game of three or more nobody to attack, which is the ruling; a copy
// that entered without choosing restricts nothing, which is the other.
// One that loses its abilities restricts nothing either.
//
// No simplification. Guffster's Party deck request (#4184, #4197).
func init() {
	Register(Spec{
		OracleID:     "377d0fa7-1b3d-4b97-b066-dceeff3734d0",
		Name:         "Pramikon, Sky Rampart",
		Completeness: CompletenessFull,
		AsEnters:     ChooseLeftOrRightAsEnters("Pramikon, Sky Rampart"),

		AttackOnlyNearestOpponentInChosenDirection: true,
	})
}
