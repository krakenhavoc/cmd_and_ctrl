package game

// ManaRestrictYourCommander permits the token only to cast a commander
// the payer owns (CR 106.6, CR 903.3d): Queen's Diamond's "Spend this
// mana only to cast your commander". "Casting a commander" is a spell
// that is a commander (CR 903.3d), so the tag refuses an activation,
// every other spell, another player's commander, and a payment that does
// not say what it casts: weaker than printed, never stronger.
const ManaRestrictYourCommander = "cast:your-commander"
