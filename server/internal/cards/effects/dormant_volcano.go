package effects

// Dormant Volcano — Land (Visions, #5631):
//
//	"This land enters tapped.
//	 When this land enters, sacrifice it unless you return an untapped
//	 Mountain you control to its owner's hand.
//	 {T}: Add {C}{R}."
//
// karooReturningUntapped (karoo_return_untapped_land.go), with the rest
// of its cycle.
//
// No simplification.
func init() {
	Register(karooReturningUntapped("38ba1956-5505-4a7a-b6af-e75715b1401f", "Dormant Volcano", "Mountain", "R", dormantVolcanoReturn))
}

var dormantVolcanoReturn = newReturnLandsOrSacrifice(
	"Dormant Volcano — return an untapped Mountain you control to its owner's hand (or sacrifice Dormant Volcano)", 1,
	"choose-cards/dormant-volcano-return-an-untapped-mountain", "vis/dormant-volcano-after-return",
	untappedLandOfType("Mountain"))
