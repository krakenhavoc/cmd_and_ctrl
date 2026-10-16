package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// captainAmericaShielded is the condition both receivers of Captain
// America's last ability share: he has a shield counter on him now.
func captainAmericaShielded(source game.Card) bool {
	return source.Counters[game.CounterShield] > 0
}

// Captain America, Super-Soldier — Legendary Creature — Human Soldier
// Hero {1}{W}{W}, 3/2:
//
//	"First strike
//	 Captain America enters with a shield counter on him. (If he would
//	 be dealt damage or destroyed, remove a shield counter from him
//	 instead.)
//	 As long as Captain America has a shield counter on him, you and
//	 other Heroes you control have hexproof."
//
// The shield counter is CR 122.1c's, which the engine owns (#2000): the
// entry is a self-replacement that puts it on as he enters. The last
// ability has two receivers, each spelled in its own place, and both
// read the counter as it is now:
//
//   - other Heroes you control (CR 702.11b) are a layer-6 grant whose
//     condition is the counter, re-answered when a counter moves;
//   - you (CR 702.11c) are Spec.PlayerKeywordsIf (#2013, ADR 0072's
//     2026-10-16 amendment), asked on every targeting query.
//
// The first damage or destroy effect that reaches him spends the counter
// and switches both off.
//
// No simplification.
func init() {
	heroes := TribeFilter{Tribes: []string{"Hero"}, Others: true, YoursOnly: true}
	heroGrant := KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return captainAmericaShielded(*source) && heroes.Matches(target, g, source)
	}, KeywordHexproof)
	heroGrant.Label = "As long as Captain America has a shield counter on him, other Heroes you control have hexproof."

	Register(Spec{
		OracleID:        "ac1b6ffc-12a9-4d9c-a9fc-68a839a40738",
		Name:            "Captain America, Super-Soldier",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterShield, 1, "Captain America enters with a shield counter on him."),
		},
		Static: []game.StaticAbility{heroGrant},
		PlayerKeywordsIf: []game.PlayerKeywordIf{
			YouHaveWhile(KeywordHexproof,
				"As long as Captain America has a shield counter on him, you have hexproof.",
				func(_ *game.Game, source game.Card) bool { return captainAmericaShielded(source) }),
		},
	})
}
