package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Charnel Serenade — Sorcery {4}{B}{B} (Murders at Karlov Manor Commander):
//
//	"Surveil 3, then return a creature card from your graveyard to the
//	 battlefield with a finality counter on it. Exile Charnel Serenade with
//	 three time counters on it.
//	 Suspend 3—{2}{B} (...)"
//
// Surveil 3 asks first; its keyed tail (the "then") exiles the spell with
// three time counters and asks which creature card of the graveyard to
// return, so a creature the surveil just binned is a choice. Nothing is
// targeted: the pick is made on resolution from the graveyard as it stands.
// The chosen card enters with a finality counter (CR 122.1) through the
// keyed return, so a question the entry asks is a restore point.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "54c3d8ba-ee7c-4f15-8ee3-e5daf860b428",
		Name:           "Charnel Serenade",
		Completeness:   CompletenessFull,
		SpecialActions: []game.SpecialAction{Suspend(3, "{2}{B}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return Surveil{Player: item.Controller, N: 3, ThenTail: ctx.ItemTail(charnelSerenadeThen)}.Apply(ctx)
		},
	})
}

// charnelSerenadeThen is the rest of the spell after the surveil. An
// on-disk key.
var charnelSerenadeThen = CardTail("charnel-serenade/then", func(ctx *Context, _ game.TailOutcome) error {
	g := ctx.Game
	you := ctx.Controller()
	// The candidates are read now, before the spell leaves the stack.
	candidates := graveyardCardIDs(ctx, you, func(c game.Card) bool { return c.IsCreature() && !c.IsToken() })
	if err := cmdr2S40ExileWithTimeCounters(ctx, 3); err != nil {
		return err
	}
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    you,
		FromPlayer: you,
		Source:     ctx.Source(),
		Question:   "Charnel Serenade — return a creature card from your graveyard to the battlefield with a finality counter on it",
		Cards:      candidates,
		Min:        1,
		Max:        1,
		ThenKey:    charnelSerenadeReturn,
		LiveZone:   game.ZoneGraveyard,
	})
	return nil
})

// charnelSerenadeReturn returns the chosen creature card with a finality
// counter. It runs from the answered prompt and captures nothing. An on-disk
// key.
var charnelSerenadeReturn = game.RegisterChooseCardsThen("choose-cards/charnel-serenade-return",
	func(g *game.Game, r game.CardsChosen) error {
		if !r.Answered || len(r.Picked) == 0 {
			return nil
		}
		return g.ReturnFromGraveyardWithCountersKeyedForEffect(r.Picked[0], r.Chooser, false,
			map[string]int{game.CounterFinality: 1}, game.KeyedTail{})
	})
