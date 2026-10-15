package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Coin of Fate — Artifact {1}{W} (deck request #3655):
//
//	"When this artifact enters, surveil 1.
//	 {3}{W}, {T}, Exile two creature cards from your graveyard, Sacrifice
//	 this artifact: An opponent chooses one of the exiled cards. You put
//	 that card on the bottom of your library and return the other to the
//	 battlefield tapped. You become the monarch."
//
// The two creature cards are an exile cost picked at announce, read back
// through ctx.Exiled(). The next opponent in turn order chooses; the
// chosen card goes to the bottom of its owner's library, the other
// returns tapped, and the crown is taken whatever happens to the cards.
//
// DECLARED SIMPLIFICATION (which opponent): "an opponent chooses" lets
// the controller name the opponent. The next opponent in turn order is
// always the one asked.
func init() {
	Register(Spec{
		OracleID:     "e43b478b-4cb8-42b1-a5b0-0b1421b84852",
		Name:         "Coin of Fate",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The next opponent in turn order chooses the exiled card, rather than you picking which opponent does."},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Coin of Fate — surveil 1", Do(Surveil{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label: "{3}{W}, {T}, Exile two creature cards from your graveyard, Sacrifice this artifact: An opponent chooses one of the exiled cards. You put that card on the bottom of your library and return the other to the battlefield tapped. You become the monarch.",
			Cost: Plus(ManaCost("{3}{W}"), TapCost(),
				ExileFromGraveyard(2, "two creature cards", MatchCreature), SacrificeThis()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  coinOfFateResolve,
		}},
	})
}

func coinOfFateResolve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	exiled := ctx.Exiled()
	monarch := func(g *game.Game) error { return BecomeTheMonarch{}.Apply(NewContext(g, item)) }
	if len(exiled) == 0 {
		return monarch(g)
	}
	chooser := ctx.Controller()
	if opps := opponentsInTurnOrder(g, ctx.Controller()); len(opps) > 0 {
		chooser = opps[0]
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    chooser,
		FromPlayer: ctx.Controller(),
		Source:     ctx.Source(),
		Question:   "Coin of Fate — choose one of the exiled cards to go on the bottom of their library; the other returns to the battlefield tapped",
		Cards:      exiled,
		Min:        1,
		Max:        1,
		Zone:       game.ZoneExile,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			ctx := NewContext(g, item)
			if len(picked) == 0 {
				return monarch(g)
			}
			return PutIntoLibrary{Card: picked[0], ToBottom: true, Then: func(g *game.Game, _ bool) error {
				ctx := NewContext(g, item)
				for _, id := range exiled {
					if id == picked[0] {
						continue
					}
					if err := (ReturnFromExile{Target: id, Tapped: true}).Apply(ctx); err != nil {
						return err
					}
				}
				return monarch(g)
			}}.Apply(ctx)
		},
	})
	return nil
}
