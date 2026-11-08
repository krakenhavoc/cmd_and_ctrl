package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Surveillance Monitor — Creature — Vedalken Detective {3}{U}, 3/3 (Murders
// at Karlov Manor, #5223):
//
//	"When this creature enters, you may collect evidence 4. (Exile cards
//	 with total mana value 4 or greater from your graveyard.)
//	 Whenever you collect evidence, create a 1/1 colorless Thopter artifact
//	 creature token with flying."
//
// The enters trigger is the effect form of collect evidence 4
// (RegisterCollectEvidence) with nothing after it; a graveyard holding less
// than 4 in all asks nothing (CR 701.59b). "Whenever you collect evidence"
// is WheneverYouCollectEvidence, off game.EventCollectEvidence: it sees
// every collection its controller makes, the one its own enters trigger
// makes included, since the Monitor is on the battlefield by then.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9be5a350-c38e-425f-9d6a-8eb111132a89",
		Name:         "Surveillance Monitor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Surveillance Monitor — you may collect evidence 4",
				func(g *game.Game, item *game.StackItem) error {
					return surveillanceMonitorEvidence.Ask(NewContext(g, item), "Surveillance Monitor — collect evidence 4 (exile cards with total mana value 4 or more from your graveyard)? Choose none to decline.")
				}),
			WheneverYouCollectEvidence("Surveillance Monitor — create a 1/1 Thopter with flying",
				Do(CreateToken{Template: TokenCard("1/1 colorless Thopter artifact with flying"), N: 1})),
		},
	})
}

// surveillanceMonitorEvidence is "you may collect evidence 4." with nothing
// after it. An on-disk key: never renamed, never reused.
var surveillanceMonitorEvidence = RegisterCollectEvidence("choose-cards/surveillance-monitor-collect-evidence", 4, nil)
