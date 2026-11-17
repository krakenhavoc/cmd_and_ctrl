package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dimir Machinations — Sorcery {2}{B} (Ravnica: City of Guilds, #5600):
//
//	"Look at the top three cards of target player's library. Exile any
//	 number of those cards, then put the rest back in any order.
//	 Transmute {1}{B}{B} ({1}{B}{B}, Discard this card: Search your
//	 library for a card with the same mana value as this card, reveal it,
//	 put it into your hand, then shuffle. Transmute only as a sorcery.)"
//
// Portent's look (only the caster learns the cards, CR 701.20) with a
// choose-cards pick of any number of them, none included. The picked
// cards are exiled face up, together, and the rest go back on top in the
// order the caster chooses once the exile has settled (CR 608.2c). A
// target player gone by resolution makes it do nothing (CR 608.2b).
// Transmute is the keyword's constructor (transmute.go, CR 702.53a),
// finding a mana value 3 card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cab67380-4e1a-4210-a98b-d2a21022b4b7",
		Name:         "Dimir Machinations",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			var victim uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetPlayer {
					victim = t.ID
				}
			}
			if victim == uuid.Nil {
				return nil
			}
			looked := ctx.Game.LookAtTopOfPlayersLibraryForEffect(ctx.Controller(), victim, 3)
			if len(looked) == 0 {
				return nil
			}
			ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:    ctx.Controller(),
				FromPlayer: victim,
				Source:     ctx.Source(),
				Question:   "Dimir Machinations — exile any number of those cards; the rest go back on top in any order",
				Cards:      looked,
				Min:        0,
				Max:        len(looked),
				LiveZone:   game.ZoneLibrary,
				ThenKey:    dimirMachinationsExile,
				Carry:      looked,
				Asked:      game.NewAskedItem(item),
			})
			return nil
		},
		Activated: []ActivatedAbility{Transmute("{1}{B}{B}")},
	})
}

// dimirMachinationsExile exiles the cards picked; the carry is the cards
// looked at, top-first, and the ones not picked ride the exile's tail. An
// on-disk key: never renamed, never reused.
var dimirMachinationsExile = askedPick("choose-cards/dimir-machinations-exile", func(ctx *Context, r game.CardsChosen) error {
	picked := make(map[uuid.UUID]bool, len(r.Picked))
	for _, id := range r.Picked {
		picked[id] = true
	}
	var rest []uuid.UUID
	for _, id := range r.Carry {
		if !picked[id] {
			rest = append(rest, id)
		}
	}
	return ctx.Game.ExileCardsKeyedForEffect(r.Picked, withCarry(ctx.ItemTail(dimirMachinationsPutBack), rest...))
})

// dimirMachinationsPutBack is "then put the rest back in any order": the
// cards not exiled, carried top-first, still in the library. An on-disk
// key: never renamed, never reused.
var dimirMachinationsPutBack = CardTail("then/dimir-machinations-put-back", func(ctx *Context, o game.TailOutcome) error {
	return PutInLibraryInAnyOrder{
		Cards:     o.Tail.Carry,
		From:      game.ZoneLibrary,
		Placement: game.LibraryPlaceTop,
		Label:     "Dimir Machinations — put the rest back in any order",
	}.Apply(ctx)
})
