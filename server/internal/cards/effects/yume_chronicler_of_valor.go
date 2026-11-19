package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Yume, Chronicler of Valor — Legendary Creature — Human Druid {1}{W}, 1/2
// (Mystery Booster Commander, #5948):
//
//	"The first spell you cast during each of your turns has assist. When you
//	 cast that spell, if another player spent mana to cast it, you and that
//	 player each draw a card. (As you cast a spell with assist, you may choose
//	 another player. They may pay some or all of the generic mana in the
//	 spell's cost.)"
//
// The first sentence is Spec.FirstSpellHasAssist, a static of the permanent
// read as the spell is announced (game.SpellHasAssistLocked): it is the first
// spell its controller casts on their own turn, counted before it counts itself.
// A spell that already has assist keeps it, and a second spell that turn has
// none unless it prints it. The helper is chosen and pays through the engine's
// ordinary assist payment.
//
// The second sentence is the same first-spell cast, with an intervening "if"
// read as it is cast (CR 603.4): the spell's paid-cost record names a player
// other than the caster who paid mana for it (Paid.AssistedBy with AssistPaid
// above zero; a helper who paid nothing did not spend mana). That helper rides
// the trigger, so the draw still happens if the spell is countered before it
// resolves; the intervening "if" is a fact about how it was cast and cannot
// change. The caster and the helper each draw a card, the caster first.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:            "f6841a6f-7dd3-4976-9891-539189c0e148",
		Name:                "Yume, Chronicler of Valor",
		Completeness:        CompletenessFull,
		FirstSpellHasAssist: true,
		Purpose:             game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return gapB05YumeHelper(ev, source, g) != uuid.Nil
			},
			Key: yumeDrawLabel,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, yumeDrawLabel)
				item.Params.Player = gapB05YumeHelper(ev, source, g)
				return item
			},
			Effect: yumeDraw,
		}},
	})
}

const yumeDrawLabel = "Yume, Chronicler of Valor — you and the player who helped pay each draw a card"

// gapB05YumeHelper is the player other than the caster who spent mana to cast
// the first spell its controller cast this turn, or uuid.Nil: not this
// controller's cast, not their own turn, not their first spell, or no helper
// that paid. The cast tally already counts this spell, so "first" is a total
// of one.
func gapB05YumeHelper(ev game.Event, source *game.Card, g *game.Game) uuid.UUID {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return uuid.Nil
	}
	if g.Turn.ActiveSeat < 0 || g.Turn.ActiveSeat >= len(g.Seats) {
		return uuid.Nil
	}
	if active := g.Seats[g.Turn.ActiveSeat]; active == nil || active.ID != source.Controller {
		return uuid.Nil
	}
	if g.CastTallyFor(source.Controller).Total != 1 {
		return uuid.Nil
	}
	item := g.StackItemForEffect(ev.CardID)
	if item == nil || item.Kind != game.StackItemSpell {
		return uuid.Nil
	}
	if item.Paid.AssistedBy == uuid.Nil || item.Paid.AssistedBy == source.Controller || item.Paid.AssistPaid <= 0 {
		return uuid.Nil
	}
	return item.Paid.AssistedBy
}

func yumeDraw(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
		return err
	}
	if helper := item.Params.Player; helper != uuid.Nil {
		return DrawCards{Player: helper, N: 1}.Apply(ctx)
	}
	return nil
}
