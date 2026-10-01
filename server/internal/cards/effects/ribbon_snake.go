package effects

// Ribbon Snake — Creature — Snake {1}{U}{U}, 2/3:
//
//	"Flying
//	 {2}: This creature loses flying until end of turn. Any player may
//	 activate this ability."
//
// Flying is a printed keyword (CR 702.9). The activation is an
// any-player row (CR 602.2, 602.1b): whoever activates it pays the {2}
// out of their own pool (CR 602.1a), and the Snake loses flying until
// end of turn, so a player about to be attacked can make it blockable
// by their ground creatures. The loss is a layer-6 ability-removing
// effect with its own timestamp (CR 613.1f), so a flying granted LATER
// in the turn puts it back (effects_this_until_end_of_turn.go).
//
// No Purpose: the bot never pays to change a creature it does not
// control (ADR 0106 owner decision 2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "98b53af9-af1b-4d5d-ae5a-bd444c1339d2",
		Name:            "Ribbon Snake",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:     "{2}: This creature loses flying until end of turn. Any player may activate this ability.",
			Cost:      ManaCost("{2}"),
			AnyPlayer: true,
			Effect:    thisLosesKeywordUntilEndOfTurn("flying", "Ribbon Snake — loses flying until end of turn"),
		}},
	})
}
