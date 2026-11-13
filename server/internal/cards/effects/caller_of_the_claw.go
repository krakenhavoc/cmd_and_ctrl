package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Caller of the Claw — Creature — Elf {2}{G}, 2/2 (Legions, #5609):
//
//	"Flash
//	 When this creature enters, create a 2/2 green Bear creature token
//	 for each nontoken creature put into your graveyard from the
//	 battlefield this turn."
//
// Flash (CR 702.8a) is what lets it follow a block or a wrath. The count
// is Rise of the Dread Marn's, read over the turn's event log as the
// trigger resolves and limited to cards you own, since a card is put
// into its owner's graveyard (CR 400.3): a creature of yours an opponent
// controlled counts, one of theirs you controlled does not, and a card
// that has since left your graveyard still counts. Tokens never do.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ec79255f-4b4d-4a0a-89d8-c59ba640275b",
		Name:            "Caller of the Claw",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Caller of the Claw — create a 2/2 green Bear for each nontoken creature put into your graveyard from the battlefield this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					n := nontokenCreaturesDiedThisTurnOwnedBy(g, ctx.Controller())
					return CreateToken{Template: TokenCard("2/2 green Bear"), N: n}.Apply(ctx)
				}),
		},
	})
}
