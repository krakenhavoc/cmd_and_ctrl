package effects

// midB03TakeRevealRestGraveyard is "you may reveal a … card from among them
// and put it into your hand. Put the rest into your graveyard" (Tapping at
// the Window). The existing takeRestGraveyard does not reveal the pick.
// An on-disk key, never renamed or reused.
var midB03TakeRevealRestGraveyard = RegisterTake("choose-cards/s82-mid-b03-take-reveal-rest-graveyard",
	TakeDef{Reveal: true, Then: TakeRestIntoGraveyard})
