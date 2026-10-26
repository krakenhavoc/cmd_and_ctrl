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
// through ctx.Exiled(). "An opponent chooses": the controller names the
// opponent (a choose-a-player prompt, skipped with a single opponent),
// and that opponent chooses; the chosen card goes to the bottom of its
// owner's library, the other returns tapped, and the crown is taken
// whatever happens to the cards.
func init() {
	Register(Spec{
		OracleID:     "e43b478b-4cb8-42b1-a5b0-0b1421b84852",
		Name:         "Coin of Fate",
		Completeness: CompletenessFull,
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
	if len(exiled) == 0 {
		return coinOfFateMonarch(g, item.Controller, item.SourceCardID)
	}
	// "An opponent chooses": the controller names which opponent. With
	// only one opponent left there is nothing to name, so no prompt.
	if opps := ctx.Opponents(); len(opps) == 1 {
		return coinOfFateAsk(g, item.Controller, item.SourceCardID, exiled, opps[0])
	}
	// Keyed (#4253): the exiled cards ride the prompt as its Carry.
	return ChoosePlayer{
		Among:    Opponents,
		Question: "Coin of Fate — choose an opponent to choose one of the exiled cards",
		ThenKey:  coinOfFateOpponentChosen,
		Carry:    exiled,
	}.Apply(ctx)
}

// coinOfFateOpponentChosen runs once the controller has named the
// opponent. Nobody to name (no opponent could be asked about) leaves the
// controller to choose, the only seat left to ask. An on-disk key: never
// renamed, never reused.
var coinOfFateOpponentChosen = OptionPickThen("option-pick/coin-of-fate-opponent", func(ctx *Context, r game.OptionPicked) error {
	chooser := ctx.Controller()
	if r.Option != nil && r.Option.Player != uuid.Nil {
		chooser = r.Option.Player
	}
	return coinOfFateAsk(ctx.Game, ctx.Controller(), r.Source, r.Carry, chooser)
})

func coinOfFateMonarch(g *game.Game, controller, source uuid.UUID) error {
	return BecomeTheMonarch{}.Apply(keyedContext(g, controller, source))
}

// coinOfFateAsk puts the "choose one of the exiled cards" prompt to
// `chooser`. The controller is the prompt's FromPlayer and the exiled
// cards its Carry, so the prompt holds no closure (#4253).
func coinOfFateAsk(g *game.Game, controller, source uuid.UUID, exiled []uuid.UUID, chooser uuid.UUID) error {
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    chooser,
		FromPlayer: controller,
		Source:     source,
		Question:   "Coin of Fate — choose one of the exiled cards to go on the bottom of their library; the other returns to the battlefield tapped",
		Cards:      exiled,
		Min:        1,
		Max:        1,
		ThenKey:    coinOfFatePicked,
		LiveZone:   game.ZoneExile,
		Carry:      exiled,
	})
	return nil
}

// coinOfFatePicked is the keyed answer to "choose one of the exiled cards":
// the chosen one goes to the bottom of its owner's library, the other
// returns tapped, and the crown is taken whatever happens to the cards. An
// on-disk key: never renamed, never reused.
var coinOfFatePicked = game.RegisterChooseCardsThen("choose-cards/coin-of-fate-picked", func(g *game.Game, r game.CardsChosen) error {
	controller, source, exiled := r.From, r.Source, r.Carry
	if len(r.Picked) == 0 {
		return coinOfFateMonarch(g, controller, source)
	}
	picked := r.Picked[0]
	return PutIntoLibrary{Card: picked, ToBottom: true, Then: func(g *game.Game, _ bool) error {
		ctx := keyedContext(g, controller, source)
		for _, id := range exiled {
			if id == picked {
				continue
			}
			if err := (ReturnFromExile{Target: id, Tapped: true}).Apply(ctx); err != nil {
				return err
			}
		}
		return coinOfFateMonarch(g, controller, source)
	}}.Apply(keyedContext(g, controller, source))
})
