package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Polygraph Orb — Artifact {4}{B} (Murders at Karlov Manor, #5193):
//
//	"When this artifact enters, look at the top four cards of your library.
//	 Put two of them into your hand and the rest into your graveyard. You
//	 lose 2 life.
//	 {2}, {T}, Collect evidence 3: Each opponent loses 3 life unless they
//	 discard a card or sacrifice a creature. (To collect evidence 3, exile
//	 cards with total mana value 3 or greater from your graveyard.)"
//
// The enters trigger is Stargaze's look (TakeFromLibraryToHand with the
// rest put into the graveyard, not milled), and the life is lost after the
// two cards are taken. The activated ability's cost is CollectEvidenceN;
// its effect is Torment of Hailfire's punisher asked once, with "sacrifice
// a creature" in place of a nonland permanent: each opponent, in APNAP
// order, picks losing 3 life, discarding a card or sacrificing a creature,
// and is offered only what they can do.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bb039957-9e88-48cf-b41c-38d42eb87251",
		Name:         "Polygraph Orb",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Polygraph Orb — look at the top four cards of your library, put two into your hand and the rest into your graveyard; you lose 2 life",
				Do(Instruction(func(ctx *Context) error {
					player := ctx.Controller()
					return TakeFromLibraryToHand{
						Player: player,
						Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 4),
						Max:    2,
						Label:  "Polygraph Orb — put two of them into your hand and the rest into your graveyard",
						Key:    takeRestGraveyard,
					}.Apply(ctx)
				}), GainLife{Amount: -2})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}, Collect evidence 3: Each opponent loses 3 life unless they discard a card or sacrifice a creature.",
			Cost:    Plus(ManaCost("{2}"), TapCost(), CollectEvidenceN(3)),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return polygraphOrbPunisher.start(NewContext(g, item), 1)
			},
		}},
	})
}

// polygraphOrbPunisher is the printed punisher: 3 life a refusal, or a
// discard, or a creature sacrificed.
var polygraphOrbPunisher = punisherRepeat{Name: "Polygraph Orb", Life: 3, Creature: true}
