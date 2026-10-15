package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kayla's Music Box — Legendary Artifact {2} (#3787):
//
//	"{W}, {T}: Look at the top card of your library, then exile it face
//	 down. (You may look at it any time.)
//	 {T}: Until end of turn, you may play cards you own exiled with
//	 Kayla's Music Box."
//
// Urianger Augurelt's two abilities without the discount and without the
// "you may": the exile is not optional (ExileTopFaceDownLinked with
// Required), and the permission is narrowed to cards its holder OWNS
// (PlayCardsExiledWith.OwnedByPlayer). That clause is not redundant:
// every card the Box exiled came off its controller's library, but a
// Box that changes hands is a new controller who owns none of them, and
// a card a thief could play off the old owner's exile is stronger than
// printed.
//
// "You may look at it any time" is the face-down kind the exile lands
// in (FaceDownHidden: the controller of the permanent that exiled it
// may look, nobody else may — CR 406.3), the same one hideaway uses.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8110f15d-82db-4299-a959-09c22475daec",
		Name:         "Kayla's Music Box",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{W}, {T}: Look at the top card of your library, then exile it face down.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{W}"), TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return ExileTopFaceDownLinked{
						Source:   HiddenRefOfActivation(item),
						Name:     "Kayla's Music Box",
						Required: true,
					}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "{T}: Until end of turn, you may play cards you own exiled with Kayla's Music Box.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    TapCost(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return PlayCardsExiledWith{
						Source:        HiddenRefOfActivation(item),
						OwnedByPlayer: true,
						Label:         "Kayla's Music Box — play cards you own exiled with it",
					}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
