package effects

// Urza's Saga creatures with keywords, batch 727 (#5624): the keys a card
// needs registered as an on-disk identity.

// S82OldB727SacrificeAnEnchantmentOrThisAnswer is Endless Wurm's tax answered,
// keyed: "sacrifice an enchantment", or sacrifice the creature. Index 0 asks
// the controller for the enchantment; any other index sacrifices the Wurm. A
// controller with no enchantment is not asked and sacrifices it
// (SacrificeThisUnlessYouSacrificeA).
var S82OldB727SacrificeAnEnchantmentOrThisAnswer = sacrificeOneOrThisAnswer(
	"option-pick/sacrifice-this-unless-you-sacrifice-an-enchantment", "an enchantment", Enchantment())
