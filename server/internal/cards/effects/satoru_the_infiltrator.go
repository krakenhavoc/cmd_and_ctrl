package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Satoru, the Infiltrator — Legendary Creature — Human Ninja Rogue
// {U}{B}, 2/3 (EDHREC rank 3232):
//
//	"Menace
//	 Whenever Satoru and/or one or more other nontoken creatures you
//	 control enter, if none of them were cast or no mana was spent
//	 to cast them, draw a card."
//
// The cheat-into-play commander. Menace rides PrintedKeywords. The
// trigger fires on a nontoken creature — Satoru itself included —
// entering under the controller's control from anywhere but the
// stack: a reanimation, a flicker, a "put onto the battlefield"
// (b30NontokenCreatureYouControlEnteredUncast). "One or more" is
// one draw per batch (OncePerBatch), so a mass
// reanimation draws once, as printed. A cast Satoru draws nothing,
// as printed.
//
// One declared simplification, weaker than printed: the "or no mana
// was spent to cast them" half is not read. The engine records a
// mana spend only under a strict-mode cast — a permissive-mode cast
// paid on paper and a genuinely free cast (cascade, "without paying
// its mana cost") leave the same log — so counting "no spend
// recorded" as "no mana was spent" would draw off ordinary casts and
// ship the card stronger (#259). Only creatures that were not cast
// at all draw the card.
func init() {
	Register(Spec{
		OracleID:        "7555c429-5f2d-4171-b6b0-8e3c8da7f314",
		Name:            "Satoru, the Infiltrator",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"Only creatures that weren't cast at all (reanimated, blinked, put onto the battlefield) draw the card — a creature you cast without spending any mana doesn't count."},
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b30NontokenCreatureYouControlEnteredUncast(ev, source, g)
			}, "Satoru, the Infiltrator — draw a card", Do(DrawCards{N: 1}))),
		},
	})
}
