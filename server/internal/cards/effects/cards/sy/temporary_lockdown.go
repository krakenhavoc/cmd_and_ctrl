package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Temporary Lockdown — Enchantment {1}{W}{W} (Dominaria United):
//
//	"When this enchantment enters, exile each nonland permanent with
//	 mana value 2 or less until this enchantment leaves the battlefield."
//
// The set is fixed as the trigger resolves (CR 608.2c): every nonland
// permanent of any controller, tokens included, with mana value 2 or
// less. Each is exiled through ExileUntil{ThisLeaves} (Abdel Adrian,
// Gorion's Ward's chain), so all of them come back, as new objects under
// their owners' control, the moment this leaves (CR 610.3), and a
// Lockdown that left before the trigger resolved exiles nothing
// (CR 610.3b). A token that is exiled ceases to exist (CR 111.7) and does
// not return.
//
// No simplification.
func init() {
	temporaryLockdownExiledOne = effects.CardTail("then/temporary-lockdown-exiled-one", func(ctx *effects.Context, o game.TailOutcome) error {
		return temporaryLockdownExileNext(ctx, o.Tail.Carry)
	})
	effects.Register(effects.Spec{
		OracleID:     "b21c99b3-b975-4428-9c9a-326ded2fa8dc",
		Name:         "Temporary Lockdown",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepExile, Partial: true}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Temporary Lockdown — exile each nonland permanent with mana value 2 or less until this leaves the battlefield",
				temporaryLockdownEnters),
		},
	})
}

func temporaryLockdownEnters(g *game.Game, item *game.StackItem) error {
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if !c.IsLand() && c.ManaValue() <= 2 {
			ids = append(ids, c.InstanceID)
		}
	}
	return temporaryLockdownExileNext(effects.NewContext(g, item), ids)
}

// temporaryLockdownExileNext exiles the permanents one by one through each
// exile's continuation. The ones still to exile ride the tail's Carry, so a
// prompt an exile pauses on is a restore point.
func temporaryLockdownExileNext(ctx *effects.Context, ids []uuid.UUID) error {
	for len(ids) > 0 {
		id := ids[0]
		ids = ids[1:]
		if !effects.OnBattlefield(ctx.Game, id) {
			continue
		}
		t := ctx.ItemTail(temporaryLockdownExiledOne)
		t.Carry = append([]uuid.UUID(nil), ids...)
		return effects.ExileUntil{
			Target:     id,
			ThisLeaves: true,
			Label:      "Temporary Lockdown — returns when Temporary Lockdown leaves the battlefield",
			ThenTail:   t,
		}.Apply(ctx)
	}
	return nil
}

// temporaryLockdownExiledOne is one exile settled, then the next. An
// on-disk key, registered in init (it refers to temporaryLockdownExileNext,
// which refers to it).
var temporaryLockdownExiledOne game.TailRef
