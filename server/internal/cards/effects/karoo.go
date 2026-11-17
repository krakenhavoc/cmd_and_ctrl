package effects

// Karoo — Land (Visions, #5631):
//
//	"This land enters tapped.
//	 When this land enters, sacrifice it unless you return an untapped
//	 Plains you control to its owner's hand.
//	 {T}: Add {C}{W}."
//
// karooReturningUntapped (karoo_return_untapped_land.go), with the rest
// of its cycle.
//
// No simplification.
func init() {
	Register(karooReturningUntapped("d4e875d9-2245-470d-aa2f-1dfe66ce2d15", "Karoo", "Plains", "W", karooReturn))
}

var karooReturn = newReturnLandsOrSacrifice(
	"Karoo — return an untapped Plains you control to its owner's hand (or sacrifice Karoo)", 1,
	"choose-cards/karoo-return-an-untapped-plains", "vis/karoo-after-return",
	untappedLandOfType("Plains"))
