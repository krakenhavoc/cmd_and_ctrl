package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dust Animus — Creature — Spirit {1}{W}, 2/3 (Outlaws of Thunder
// Junction, #4468):
//
//	"Flying
//	 If you control five or more untapped lands, this creature enters with
//	 two +1/+1 counters and a lifelink counter on it.
//	 Plot {1}{W}"
//
// Two entry replacements (CR 614.1c, Blacksnag Buzzard's
// b19EntersWithCountersCounted), each reading the same condition as the
// creature enters: the untapped lands its controller controls then. Lands
// tapped to cast it are tapped by the time it enters, so with exactly
// five lands, casting it for {1}{W} leaves three untapped and no
// counters; a free plotted cast taps nothing. The lifelink counter gives
// it lifelink (CR 122.1b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "494f8c29-7626-43ed-b0b8-fb9afd4cd5d3",
		Name:            "Dust Animus",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SpecialActions:  []game.SpecialAction{Plot("{1}{W}")},
		Replacements: []game.ReplacementEffect{
			b19EntersWithCountersCounted(game.CounterPlusOne, func(g *game.Game, src *game.Card) int {
				if dustAnimusFiveUntappedLands(g, src) {
					return 2
				}
				return 0
			}, "Dust Animus enters with two +1/+1 counters if you control five or more untapped lands"),
			b19EntersWithCountersCounted(game.CounterLifelink, func(g *game.Game, src *game.Card) int {
				if dustAnimusFiveUntappedLands(g, src) {
					return 1
				}
				return 0
			}, "Dust Animus enters with a lifelink counter if you control five or more untapped lands"),
		},
	})
}

// dustAnimusFiveUntappedLands is "if you control five or more untapped
// lands", read as the creature enters. It walks the live slice because
// it runs inside an entry replacement.
func dustAnimusFiveUntappedLands(g *game.Game, src *game.Card) bool {
	if g.Battlefield == nil || src == nil {
		return false
	}
	n := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == src.Controller && c.IsLand() && !c.Tapped {
			n++
		}
	}
	return n >= 5
}
