package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Surge to Victory — Sorcery {4}{R}{R} (Secrets of Strixhaven Commander,
// #4834):
//
//	"Exile target instant or sorcery card from your graveyard. Creatures you
//	 control get +X/+0 until end of turn, where X is that card's mana value.
//	 Whenever a creature you control deals combat damage to a player this
//	 turn, copy the exiled card. You may cast the copy without paying its
//	 mana cost."
//
// One target, exiled through the CR 614 window: only a card that really
// reached exile counts, and the rest of the spell is the keyed continuation
// that is told so (a target that left the graveyard fizzles the whole spell,
// CR 608.2b). X is the exiled card's mana value, with an {X} in its cost
// counted as 0 off the stack (CR 202.3e). The pump is the creatures its
// controller has as the spell resolves (CR 611.2c), so a creature that enters
// later this turn gets no +X/+0 but still triggers the copy.
//
// "Whenever ... this turn" is a CR 603.7 repeating delayed trigger fired once
// per creature of yours that deals combat damage to a player, simultaneous
// damage included. Each firing copies the exiled card (CR 707.12) as long as
// it is still in exile, and offers a free cast of the copy while the trigger
// resolves (CopyAndCast: a spell cast without paying its mana cost takes X as
// 0, CR 107.3b). A copy not cast ceases to exist; the original stays exiled
// for the next trigger. Casting the copy triggers "whenever you cast" and
// magecraft as any cast does; copying it does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8ea414c1-88cf-46d4-9a54-6dee9847c537",
		Name:         "Surge to Victory",
		Completeness: CompletenessFull,
		Targets: TargetCardInGraveyard("target instant or sorcery card from your graveyard",
			YouOwn(), Or(Instant(), Sorcery())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return ctx.Game.ExileCardKeyedForEffect(id, ctx.ItemTail(surgeToVictoryExiled))
		},
	})
}

// surgeToVictoryExiled is everything after the exile: the +X/+0 and the
// repeating delayed trigger, both only for a card that reached exile. An
// on-disk key: never renamed, never reused.
var surgeToVictoryExiled = CardTail("surge-to-victory/exiled", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	exiled, ok := ctx.Game.LookupCardForEffect(o.IDs[0])
	if !ok {
		return nil
	}
	if err := (BoostUntilEOT{
		Match: And(Creature(), YouControl()),
		Power: exiled.ManaValue(),
		Label: "Surge to Victory — +X/+0, where X is the exiled card's mana value",
	}).Apply(ctx); err != nil {
		return err
	}
	return DelayedOnEvent{
		Label:     "Surge to Victory — copy the exiled card; you may cast the copy without paying its mana cost",
		On:        []game.EventKind{game.EventDealDamage},
		Condition: socB14YourCreatureDealtCombatDamageToAPlayerCondition,
		Cards:     []uuid.UUID{o.IDs[0]},
		Body:      socB14CopyTheExiledCardBody,
		Repeats:   true,
	}.Apply(ctx)
})
