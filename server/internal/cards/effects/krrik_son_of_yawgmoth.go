package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// K'rrik, Son of Yawgmoth — Legendary Creature — Phyrexian Horror
// Minion {4}{B/P}{B/P}{B/P}, 2/2:
//
//	"({B/P} can be paid with either {B} or 2 life.)
//	 Lifelink
//	 For each {B} in a cost, you may pay 2 life rather than pay that
//	 mana.
//	 Whenever you cast a black spell, put a +1/+1 counter on K'rrik."
//
// # The third ability (ADR 0131, #2531)
//
// "For each {B} in a cost, you may pay 2 life rather than pay that
// mana" is Spec.LifeForMana. It changes how its controller PAYS, not the
// cost (the 2019-08-23 ruling): at every mana payment the engine marks
// each {B} — and the {B} half of a hybrid symbol like {B/G} or {2/B} —
// as payable with life, exactly like a printed {B/P}, and never generic
// mana or {C}. The player claims it at announcement through
// `phyrexian_life` (CR 601.2b), auto-tap never pays it, and the price
// shown, the mana value and K'rrik's own cast trigger are unchanged.
// Every mana payment its controller makes is covered: casts (alternative
// and additional costs included), activated abilities, attack taxes, a
// mana ability's own mana component (a filter land's "{B}, {T}: Add
// ..."; ManaAbilityParams.PhyrexianLife) and the payments made while a
// spell or ability resolves (ward {B}, Rhystic Study's tax;
// ResolvePayUnlessWithLife), the last two added by PR 2.
//
// K'rrik's OWN {B/P}{B/P}{B/P} was always payable with life, so a
// six-mana commander can be cast for {4} and 6 life on turn four.
//
// The cast trigger fires once per black spell, checked on the spell's
// COLOUR rather than on a black pip in its cost — a colour-indicator
// or effect-coloured spell counts, and a colourless spell with {B} in
// its cost does not.
func init() {
	Register(Spec{
		OracleID:        "cbe3a4e7-5dbe-4f58-8ee6-a1762b65acfd",
		Name:            "K'rrik, Son of Yawgmoth",
		Completeness:    CompletenessFull,
		LifeForMana:     YouMayPayLifeForMana("B"),
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(OfColor("B"), "K'rrik, Son of Yawgmoth — put a +1/+1 counter on it", putCounterOnSelf),
		},
	})
}
