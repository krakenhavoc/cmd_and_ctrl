package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Colossal Skyturtle — Enchantment Creature — Turtle {4}{G}{G}{U}, 6/5 (Neon
// Dynasty, #5493):
//
//	"Flying, ward {2}
//	 Channel — {2}{G}, Discard this card: Return target card from your
//	 graveyard to your hand.
//	 Channel — {1}{U}, Discard this card: Return target creature to its
//	 owner's hand."
//
// Two separate channel abilities, each its own activation from the hand. The
// graveyard target is chosen before the discard cost is paid (CR 602.2b), so
// the Skyturtle is never a legal target of its own channel.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "871e4c4e-8105-4cb9-b1cd-53c5f8e64494",
		Name:            "Colossal Skyturtle",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Colossal Skyturtle — ward {2}"),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "Channel — {2}{G}, Discard this card: Return target card from your graveyard to your hand.",
				Purpose: game.Purpose{Answers: game.AnswerGraveyard},
				Cost:    Plus(ManaCost("{2}{G}"), DiscardThis()),
				Zones:   []game.ZoneKind{game.ZoneHand},
				Targets: TargetCardInGraveyard("target card from your graveyard", YouOwn()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return returnTargetFromGraveyardToHand(item, NewContext(g, item))
				},
			},
			{
				Label:   "Channel — {1}{U}, Discard this card: Return target creature to its owner's hand.",
				Purpose: game.Purpose{Answers: game.AnswerRemove},
				Cost:    Plus(ManaCost("{1}{U}"), DiscardThis()),
				Zones:   []game.ZoneKind{game.ZoneHand},
				Targets: TargetCreature("target creature"),
				Effect:  neoB07BounceTheTarget,
			},
		},
	})
}
