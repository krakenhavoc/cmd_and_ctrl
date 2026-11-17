package effects

// Everglades — Land (Visions, #5631):
//
//	"This land enters tapped.
//	 When this land enters, sacrifice it unless you return an untapped
//	 Swamp you control to its owner's hand.
//	 {T}: Add {C}{B}."
//
// karooReturningUntapped (karoo_return_untapped_land.go), with the rest
// of its cycle.
//
// No simplification.
func init() {
	Register(karooReturningUntapped("ef3b8b0c-cea7-4bae-934c-9c65fd64245d", "Everglades", "Swamp", "B", evergladesReturn))
}

var evergladesReturn = newReturnLandsOrSacrifice(
	"Everglades — return an untapped Swamp you control to its owner's hand (or sacrifice Everglades)", 1,
	"choose-cards/everglades-return-an-untapped-swamp", "vis/everglades-after-return",
	untappedLandOfType("Swamp"))
