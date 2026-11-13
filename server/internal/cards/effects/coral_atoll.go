package effects

// Coral Atoll — Land (Visions, #5631):
//
//	"This land enters tapped.
//	 When this land enters, sacrifice it unless you return an untapped
//	 Island you control to its owner's hand.
//	 {T}: Add {C}{U}."
//
// karooReturningUntapped (karoo_return_untapped_land.go), with the rest
// of its cycle.
//
// No simplification.
func init() {
	Register(karooReturningUntapped("3f347ebf-e0d2-4ae0-ad84-df7a460404e0", "Coral Atoll", "Island", "U", coralAtollReturn))
}

var coralAtollReturn = newReturnLandsOrSacrifice(
	"Coral Atoll — return an untapped Island you control to its owner's hand (or sacrifice Coral Atoll)", 1,
	"choose-cards/coral-atoll-return-an-untapped-island", "vis/coral-atoll-after-return",
	untappedLandOfType("Island"))
