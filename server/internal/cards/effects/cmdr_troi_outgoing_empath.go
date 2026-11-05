package effects

// Cmdr. Troi, Outgoing Empath — Legendary Creature — Betazoid Human Advisor
// {2}{U}{R}, 3/3 (Star Trek Commander, #4974):
//
//	"Vigilance
//	 {T}: Target opponent whose turn it is may cast a spell from their hand
//	 without paying its mana cost. If they do, you copy that spell. You may
//	 choose new targets for the copy. If they don't, you draw a card."
//
// Vigilance is printed and enforced. The ability can target only an
// opponent who is the active player, so it is used on an opponent's turn;
// an opponent who stops being the active player before it resolves is no
// longer a legal target and the ability does nothing (CR 608.2b).
//
// The cast is made during the ability's resolution (CR 608.2g): the
// opponent is offered one card from their hand with a spell to cast for
// {0} (X is 0, CR 107.3b), and may decline. Timing restrictions are waived
// by the instruction (CR 608.2g), so a sorcery or creature may be cast.
// Once they have answered, a spell they cast is copied by you, the copy
// controlled by you with new targets offered (CR 707.10c); a permanent
// spell's copy becomes a token as it resolves (CR 111.13). If they cast
// nothing, including when they have no spell to cast, you draw a card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9a52cb6b-b0d8-4649-b65a-5f9a9d2a7eb2",
		Name:            trekB39TroiName,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target opponent whose turn it is may cast a spell from their hand without paying its mana cost. If they do, you copy that spell. You may choose new targets for the copy. If they don't, you draw a card.",
			Cost:    TapCost(),
			Targets: TargetPlayer("target opponent whose turn it is", trekB39OpponentWhoseTurnItIs()),
			Effect:  trekB39TroiOffer,
		}},
	})
}
