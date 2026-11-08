package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Akawalli, the Seething Tower — Legendary Creature — Fungus {1}{B}{G},
// 3/3 (The Lost Caverns of Ixalan, #4470):
//
//	"Descend 4 — As long as there are four or more permanent cards in
//	 your graveyard, Akawalli gets +2/+2 and has trample.
//	 Descend 8 — As long as there are eight or more permanent cards in
//	 your graveyard, Akawalli gets an additional +2/+2 and can't be
//	 blocked by more than one creature."
//
// Echo of Dusk's two statics on one condition, twice. Descend 4 is a
// layer 7c +2/+2 (CR 613.4c) and a layer 6 trample grant (CR 613.1f);
// Descend 8 is a second layer 7c +2/+2, so with eight or more permanent
// cards (CR 110.4) Akawalli is a 7/7 with trample. Its "can't be blocked
// by more than one creature" is Outland Colossus's MaxBlockers rule
// (CR 509.1b) on Akawalli alone, with the scope also asking for eight
// permanent cards, read live as blockers are declared.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0679dcef-e1c3-4739-a769-18a4ffdcfde7",
		Name:         "Akawalli, the Seething Tower",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			SelfPumpWhileDescend(4, 2, 2),
			SelfKeywordWhileDescend(4, "trample"),
			SelfPumpWhileDescend(8, 2, 2),
		},
		BlockRules: []game.BlockRule{MaxBlockers(akawalliDescendEight, 1)},
	})
}

// akawalliDescendEight is "Akawalli, while there are eight or more
// permanent cards in your graveyard", the scope of its blocker limit.
func akawalliDescendEight(g *game.Game, c, source *game.Card) bool {
	return OnSelf()(g, c, source) && permanentCardsInGraveyard(g, source.Controller) >= 8
}
