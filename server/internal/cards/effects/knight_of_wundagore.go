package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Knight of Wundagore — Creature — Cat Knight Villain {1}{G}, 2/1 (Marvel
// Super Heroes, #4455):
//
//	"Trample
//	 Whenever you put a +1/+1 counter on another creature, put a +1/+1
//	 counter on this creature. This ability triggers only once each turn."
//
// WheneverYouPutCountersOfKindOnACreature (Terrasymbiosis) judges the
// placer off the event and the creature as it is on the battlefield, and
// the match excludes the Knight itself. "Only once each turn" is the
// per-turn trigger record (b11TriggeredThisTurn), so a countered first
// trigger still uses it up.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "39145960-a212-46b4-addb-4b187f6bf6fd",
		Name:            "Knight of Wundagore",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WheneverYouPutCountersOfKindOnACreature(game.CounterPlusOne,
				func(g *game.Game, source *game.Card, c game.Card) bool {
					return c.InstanceID != source.InstanceID && !b11TriggeredThisTurn(g, source.InstanceID, knightOfWundagoreLabel)
				}, knightOfWundagoreLabel, plusOneCountersOnThis(1)),
		},
	})
}

const knightOfWundagoreLabel = "Knight of Wundagore — a +1/+1 counter on it"
