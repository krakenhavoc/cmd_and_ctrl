package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gate to the Afterlife — Artifact {3} (Doctor Who Commander):
//
//	"Whenever a nontoken creature you control dies, you gain 1 life. Then you
//	 may draw a card. If you do, discard a card.
//	 {2}, {T}, Sacrifice this artifact: Search your graveyard, hand, and/or
//	 library for a card named God-Pharaoh's Gift and put it onto the
//	 battlefield. If you search your library this way, shuffle. Activate only
//	 if there are six or more creature cards in your graveyard."
//
// The trigger looks back in time (CR 603.10a): the creature was last
// controlled by you as it died, and a token does not count. It gains 1 life
// and then asks the loot as a linked "you may draw a card. If you do, discard
// a card" (Rook Turret's shape): declining the draw declines the discard. The
// ability is Say Its Name's three-zone search (SearchLibrary AlsoGraveyard /
// AlsoHand, #5408), the library shuffled if and only if it was searched, with
// the creature-card count of CR 602.5b's "activate only if" as its
// activation condition (GraveyardAtLeast). God-Pharaoh's Gift is a stated
// name, so the search may find nothing (CR 701.23b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5a682ec5-78e6-4238-b5c2-e94b7a8f8066",
		Name:         "Gate to the Afterlife",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, cmdr2S30NontokenCreatureYouControlDied,
				"Gate to the Afterlife — you gain 1 life, then you may draw a card; if you do, discard a card",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (GainLife{Amount: 1}).Apply(ctx); err != nil {
						return err
					}
					return MayChoice{
						Question: "Gate to the Afterlife — draw a card, then discard a card?",
						Draws:    1,
						YesLabel: "Draw",
						NoLabel:  "Decline",
						Then:     celestusAnswer,
					}.Apply(ctx)
				}),
		},
		Activated: []ActivatedAbility{{
			Label:     "{2}, {T}, Sacrifice this artifact: Search your graveyard, hand, and/or library for a card named God-Pharaoh's Gift and put it onto the battlefield. If you search your library this way, shuffle. Activate only if there are six or more creature cards in your graveyard.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      Plus(ManaCost("{2}"), TapCost(), SacrificeThis()),
			Condition: GraveyardAtLeast(6, func(c game.Card) bool { return c.IsCreature() }),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:        item.Controller,
					Predicate:     func(c game.Card) bool { return c.Name == "God-Pharaoh's Gift" },
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					Shuffle:       true,
					AlsoGraveyard: true,
					AlsoHand:      true,
					Reason:        "Gate to the Afterlife — a card named God-Pharaoh's Gift from your graveyard, hand or library",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
