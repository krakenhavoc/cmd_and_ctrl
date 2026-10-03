package effects

// Quicksilver Wall — Creature — Wall {2}{U}, 1/6:
//
//	"Defender (This creature can't attack.)
//	 {4}: Return this creature to its owner's hand. Any player may
//	 activate this ability."
//
// Defender is a printed keyword (CR 702.3). The activation is an
// any-player row (CR 602.2, 602.1b): whoever activates it pays the {4}
// out of their own pool (CR 602.1a), and the Wall goes to its OWNER's
// hand, whoever controls it or activated the ability. It is not a cost:
// the Wall stays on the battlefield until the ability resolves and can
// be answered. A Wall that has already left the battlefield, or left
// and came back as a new object (CR 400.7), is not moved.
//
// No Purpose: the bot never pays to bounce a permanent it does not
// control (ADR 0106 owner decision 2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "82372776-bedd-4893-92ad-f1272f9c3250",
		Name:            "Quicksilver Wall",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []ActivatedAbility{{
			Label:     "{4}: Return this creature to its owner's hand. Any player may activate this ability.",
			Cost:      ManaCost("{4}"),
			AnyPlayer: true,
			Effect:    returnThisPermanentToOwnersHand,
		}},
	})
}
