package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Muxus, Goblin Grandee — Legendary Creature — Goblin Noble {4}{R}{R}, 6/6
// (Jumpstart 2022, #5948):
//
//	"When Muxus enters, reveal the top six cards of your library. Put all
//	 Goblin creature cards with mana value 5 or less from among them onto
//	 the battlefield and the rest on the bottom of your library in a random
//	 order.
//	 Whenever Muxus attacks, it gets +1/+1 until end of turn for each other
//	 Goblin you control."
//
// The put is one simultaneous entry (PutFromLibraryOntoBattlefield with All);
// the rest go to the bottom in a random order. The attack bonus counts the
// Goblins as the trigger resolves (CR 608.2h reads the pump's size then).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "07c82cde-f00d-4454-9df4-9286e344c9bb",
		Name:         "Muxus, Goblin Grandee",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Muxus, Goblin Grandee — reveal the top six cards; put all Goblin creature cards with mana value 5 or less onto the battlefield and the rest on the bottom in a random order",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					player := ctx.Controller()
					return PutFromLibraryOntoBattlefield{
						Player:   player,
						Cards:    g.RevealTopOfLibraryForEffect(player, ctx.Source(), 6, "Muxus, Goblin Grandee — reveal the top six cards"),
						Match:    And(Creature(), HasSubtype("Goblin"), ManaValueLE(5)),
						All:      true,
						Label:    "Muxus, Goblin Grandee — put all Goblin creature cards with mana value 5 or less onto the battlefield",
						ThenTail: PutRestOnBottomInRandomOrderTail,
					}.Apply(ctx)
				}),
			WheneverThisAttacks("Muxus, Goblin Grandee — it gets +1/+1 until end of turn for each other Goblin you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					n := 0
					for _, c := range MatchingBattlefield(ctx, And(YouControl(), HasSubtype("Goblin"))) {
						if c.InstanceID != item.SourceCardID {
							n++
						}
					}
					return BoostUntilEOT{Target: item.SourceCardID, Power: n, Toughness: n, Label: "Muxus, Goblin Grandee — +1/+1 for each other Goblin"}.Apply(ctx)
				}),
		},
	})
}
