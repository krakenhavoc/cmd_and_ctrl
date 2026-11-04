package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rousing Refrain — Sorcery {3}{R}{R} (Secrets of Strixhaven Commander,
// #4834):
//
//	"Add {R} for each card in target opponent's hand. Until end of turn,
//	 you don't lose this mana as steps and phases end. Exile Rousing
//	 Refrain with three time counters on it.
//	 Suspend 3—{1}{R} (Rather than cast this card from your hand, you may
//	 pay {1}{R} and exile it with three time counters on it. At the
//	 beginning of your upkeep, remove a time counter. When the last is
//	 removed, you may cast it without paying its mana cost.)"
//
// Suspend is the printed special action (Suspend(3, "{1}{R}"), CR 702.62a),
// which also puts the two exile-zone triggers on the card.
//
// Cast normally, the spell counts the cards in the targeted opponent's hand
// as it resolves and adds that much {R}, marked to last until end of turn
// (KeepManaUntilEndOfTurn, #2166: the step-boundary sweep leaves it
// until cleanup). Then the spell exiles ITSELF and puts three time counters
// on it: a card in exile with time counters is suspended (CR 702.62b), so
// the printed suspend triggers count it down and let its owner cast it
// again for free. The exile is a keyed tail on the exile
// (ExileTarget.ThenTail), so the counters land on the card that really
// reached exile; a spell with no legal target does not resolve, goes to the
// graveyard and is never exiled, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "515c522f-51d5-4f6d-b9b1-ea3c7f6a378e",
		Name:           "Rousing Refrain",
		Completeness:   CompletenessFull,
		SpecialActions: []game.SpecialAction{Suspend(3, "{1}{R}")},
		Targets:        TargetPlayer("target opponent", Opponent()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (AddMana{
				Player: item.Controller,
				Amount: ManaN(b14HandSize(ctx.Game, TargetedPlayer(ctx)), "R"),
				Riders: []game.ManaSpendRider{KeepManaUntilEndOfTurn()},
			}).Apply(ctx); err != nil {
				return err
			}
			return ExileTarget{Target: item.SourceCardID, ThenTail: ctx.ItemTail(rousingRefrainSuspended)}.Apply(ctx)
		},
	})
}

// rousingRefrainSuspended is "with three time counters on it": the card
// that reached exile is suspended. An on-disk key.
var rousingRefrainSuspended = CardTail("rousing-refrain/three-time-counters", func(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if err := ctx.Game.AddCounterByForEffect(ctx.Controller(), id, game.CounterTime, 3); err != nil {
			return err
		}
	}
	return nil
})
