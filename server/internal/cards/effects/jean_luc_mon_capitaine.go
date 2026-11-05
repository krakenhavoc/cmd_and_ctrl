package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jean-Luc, Mon Capitaine — Legendary Creature — Human Officer
// {2}{U}{R}{W}, 5/3 (Star Trek Commander, #4974):
//
//	"Flying
//	 Federation — Faeries you control enter with an additional +1/+1
//	 counter on them for each creature type among non-Borg creatures you
//	 already control."
//
// Federation is an ability word and has no rules meaning (CR 207.2c).
// The ability is a CR 614.12 entry replacement
// (trekB47FaeriesEnterWithTypeCounters): a Faerie entering under his
// controller — cast, reanimated, flickered or a token — enters with one
// +1/+1 counter for each creature type among the non-Borg creatures that
// player controls as it enters, each type counted once however many
// creatures share it. The entering Faerie is not yet on the battlefield,
// so its own types do not count; Faeries entering together do not count
// each other. Hardened Scales and Doubling Season apply to the counters,
// because they replace the same event.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3855cb66-2e29-4d4b-a752-214843b647ff",
		Name:            "Jean-Luc, Mon Capitaine",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{
			trekB47FaeriesEnterWithTypeCounters("Jean-Luc, Mon Capitaine — Federation: a Faerie enters with an additional +1/+1 counter for each creature type among non-Borg creatures you control"),
		},
	})
}
