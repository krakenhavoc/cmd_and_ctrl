package effects

// take_defs.go — the takes the catalog shares (#4253). A TakeFromLibraryToHand
// names its rest-of-the-card as a registered key (RegisterTake), so the
// choose_cards prompt of a look holds no closure and a table waiting on it
// is a restore point. Cards whose consequence is the same share a key: what
// differs between them (the prompt's header, the count, the filter, "may")
// is the take's own data, not the registration's.
//
// Each key is on disk (ledgered "choose <key>") and is never renamed or
// reused.

var (
	// takeRestBottomRandom is "…and put the rest on the bottom of your
	// library in a random order" (Fomori Vault, Memory Deluge, The Lady of
	// Otaria, Marina Vendrell, Yuna's Whistle, Artificer Class, Prophetic
	// Titan).
	takeRestBottomRandom = RegisterTake("choose-cards/take-rest-bottom-random",
		TakeDef{Then: TakeRestOnBottomInRandomOrder})

	// takeRestBottomAny is "…and the rest on the bottom of your library in
	// any order" (Dig Through Time, Stock Up, the look-and-take helpers).
	takeRestBottomAny = RegisterTake("choose-cards/take-rest-bottom-any-order",
		TakeDef{Then: TakeRestOnBottomInAnyOrder})

	// takeRevealRestBottomRandom is "you may reveal a … card from among
	// them and put it into your hand. Put the rest on the bottom of your
	// library in a random order" (Monumental Henge, Horn of the Mark).
	takeRevealRestBottomRandom = RegisterTake("choose-cards/take-reveal-rest-bottom-random",
		TakeDef{Reveal: true, Then: TakeRestOnBottomInRandomOrder})

	// takeRevealRestBottomAny is the same with "in any order" (Duskwatch
	// Recruiter, Growing Rites of Itlimoc, Gavony Dawnguard, Oath of
	// Nissa).
	takeRevealRestBottomAny = RegisterTake("choose-cards/take-reveal-rest-bottom-any-order",
		TakeDef{Reveal: true, Then: TakeRestOnBottomInAnyOrder})

	// takeRestGraveyard is "…and put the rest into your graveyard" (Wail
	// of the Forgotten, Tamiyo, Collector of Tales).
	takeRestGraveyard = RegisterTake("choose-cards/take-rest-graveyard",
		TakeDef{Then: TakeRestIntoGraveyard})

	// sorinMirthlessTake is Sorin the Mirthless's +1: revealed, taken, and
	// the life loss equal to its mana value.
	sorinMirthlessTake = RegisterTake("choose-cards/sorin-the-mirthless-take",
		TakeDef{Reveal: true, Then: sorinMirthlessPayForTheCard})

	// expressiveIterationTake is Expressive Iteration's look: one into
	// your hand, then the bottom-and-exile of the rest.
	expressiveIterationTake = RegisterTake("choose-cards/expressive-iteration-take",
		TakeDef{Then: expressiveIterationBottomAndExile})
)
