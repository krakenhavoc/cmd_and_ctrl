package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thran Spider — Artifact Creature — Spider {3}, 2/4 (The Brothers' War):
//
//	"Reach
//	 When this creature enters, you and target opponent each create a
//	 tapped Powerstone token.
//	 {7}: Look at the top four cards of your library. You may reveal an
//	 artifact card from among them and put it into your hand. Put the
//	 rest on the bottom of your library in a random order."
//
// The trigger targets an opponent; if that target is illegal on
// resolution the whole ability is removed (CR 608.2b), so the controller
// gets no token either, as printed. The look is Lupinflower Village's
// (takeRevealRestBottomRandom).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "78d710ea-fbf5-4855-b874-b53478bf4a87",
		Name:            "Thran Spider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: Self,
			Targets:   TargetPlayer("target opponent", Opponent()),
			Key:       "Thran Spider — you and target opponent each create a tapped Powerstone token",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				var opp uuid.UUID
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetPlayer {
						opp = t.ID
					}
				}
				if opp == uuid.Nil {
					return nil
				}
				for _, who := range []uuid.UUID{item.Controller, opp} {
					if err := (CreateTokenAdvanced{Controller: who, Spec: Token(PowerstoneToken()).EntersTapped(), N: 1}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{7}: Look at the top four cards of your library. You may reveal an artifact card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{7}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				player := item.Controller
				return TakeFromLibraryToHand{
					Player:   player,
					Cards:    g.LookAtTopOfLibraryForEffect(player, 4),
					Match:    Artifact(),
					Max:      1,
					Optional: true,
					Label:    "Thran Spider — you may reveal an artifact card and put it into your hand",
					Key:      takeRevealRestBottomRandom,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
