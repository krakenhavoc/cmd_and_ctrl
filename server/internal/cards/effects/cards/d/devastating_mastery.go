package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Devastating Mastery — Sorcery {2}{W}{W}{W}{W} (Strixhaven: School of Mages,
// #5502):
//
//	"You may pay {2}{W}{W} rather than pay this spell's mana cost.
//	 If the {2}{W}{W} cost was paid, an opponent chooses up to two nonland
//	 permanents they control and returns them to their owners' hands.
//	 Destroy all nonland permanents."
//
// Baleful Mastery's alternative cost and its "an opponent" (the caster picks
// which, ChoosePlayer; a single opponent is not asked). That opponent is then
// asked, as the spell resolves, for up to two of the nonland permanents they
// control (it does not target) and returns them to their owners' hands; the
// destruction of all nonland permanents follows once the return is done (a
// commander's command-zone question included), so the returned permanents are
// spared. Without the alternative cost the wipe is the whole spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "07bfb7d2-f171-4ce1-876a-f9fb5c7cd11a",
		Name:         "Devastating Mastery",
		Completeness: effects.CompletenessFull,
		Purpose: game.Purpose{
			Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy},
		},
		AlternativeCosts: []game.AlternativeCost{{
			Key:      "devastating-mastery",
			Label:    "Pay {2}{W}{W} rather than pay this spell's mana cost",
			ManaCost: "{2}{W}{W}",
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if !ctx.PaidAltCost("devastating-mastery") {
				return devastatingMasteryWipe(ctx)
			}
			return effects.ChoosePlayer{
				Among:    effects.Opponents,
				Question: "Devastating Mastery — an opponent returns up to two of their nonland permanents",
				ThenKey:  devastatingMasteryOpponent,
			}.Apply(ctx)
		},
	})
}

func devastatingMasteryWipe(ctx *effects.Context) error {
	return effects.DestroyAllMatching{Match: effects.Nonland()}.Apply(ctx)
}

// devastatingMasteryOpponent asks the chosen opponent for up to two nonland
// permanents they control; with none to offer (or no opponent) the wipe runs.
var devastatingMasteryOpponent = effects.ChoosePlayerThen("option-pick/devastating-mastery-opponent", func(ctx *effects.Context, p uuid.UUID, _ []uuid.UUID) error {
	if p == uuid.Nil {
		return devastatingMasteryWipe(ctx)
	}
	mine := effects.NonlandPermanentsControlledBy(ctx.Game, p)
	if len(mine) == 0 {
		return devastatingMasteryWipe(ctx)
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  p,
		Source:   ctx.Source(),
		Question: "Devastating Mastery — choose up to two nonland permanents you control to return to their owners' hands",
		Cards:    mine,
		Min:      0,
		Max:      2,
		LiveZone: game.ZoneBattlefield,
		ThenKey:  devastatingMasteryBounce,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
})

// devastatingMasteryBounce returns what the opponent chose, then destroys all
// nonland permanents once the return has finished.
var devastatingMasteryBounce = effects.AskedPick("choose-cards/devastating-mastery-bounce", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return devastatingMasteryWipe(ctx)
	}
	return effects.BounceAllMatching{Match: effects.B10OneOf(r.Picked), ThenTail: ctx.ItemTail(devastatingMasteryThenWipe)}.Apply(ctx)
})

var devastatingMasteryThenWipe = effects.MassTail("tail/devastating-mastery-wipe", func(ctx *effects.Context, _ []game.Card, _ int) error {
	return devastatingMasteryWipe(ctx)
})
