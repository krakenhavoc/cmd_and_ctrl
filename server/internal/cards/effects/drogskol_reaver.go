package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Drogskol Reaver — Creature — Spirit {5}{W}{U}, 3/5 (EDHREC rank
// 4182):
//
//	"Flying
//	 Double strike
//	 Lifelink
//	 Whenever you gain life, draw a card."
//
// Seven mana for a creature that draws two cards every time it
// connects, and more than two in any deck with a second lifelinker or
// a Soul Warden. The three keywords are not decoration — they are the
// engine of the card, and their interaction is why it is worth
// registering rather than approximating: double strike plus lifelink
// means the first-strike damage step gains life, which draws a card
// BEFORE the regular damage step, and then the regular damage gains
// life and draws again.
//
// All of that falls out for free once the keywords are printed and
// the trigger watches lifegain: the engine runs two combat damage
// steps for a double striker, each emits its own life change, and
// each life change is its own trigger. A single AddLife of N is one
// trigger however large N is — the printed ability is per event, not
// per point (Archangel of Thune's reading, and the same helper).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "68175b49-f1e6-4b34-aad9-20d61a43d427",
		Name:            "Drogskol Reaver",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "double strike", "lifelink"},
		Triggered: []game.TriggeredAbility{
			WheneverYouGainLife("Drogskol Reaver — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
