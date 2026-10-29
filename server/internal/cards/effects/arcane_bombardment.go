package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arcane Bombardment — Enchantment {4}{R}{R}:
//
//	"Whenever you cast your first instant or sorcery spell each turn,
//	 exile an instant or sorcery card at random from your graveyard.
//	 Then copy each card exiled with this enchantment. You may cast any
//	 number of the copies without paying their mana costs."
//
// Built on CopyAndCast (#3258, ADR 0066's 2026-10-13 card-copy
// amendment), and the first card to read it as a standing set rather
// than a one-shot: the trigger exiles ONE more card each turn and then
// copies EVERY card it has ever exiled that is still in exile (CR 607.2a),
// so the pile grows by a card a turn and the copies are made again from
// the same originals each time. The originals are never cast, only
// copied; the copies are not cards and cease to exist as they resolve or
// when the caster declines them. They are cast, or not, while the trigger
// is still resolving (CR 608.2g, #2947).
//
//   - "First instant or sorcery spell each turn" is the per-turn tally's
//     InstantSorcery count (the cast path bumps it before EventCast), so
//     a creature spell cast between two instants does not count and the
//     second instant does not trigger. Copies cast off a previous trigger
//     count toward it like any spell, which only matters for a later
//     turn's first spell.
//   - The exile is at random from the controller's graveyard through the
//     keyed random stream (randomDraw), and it is not optional: with no
//     instant or sorcery card left the exile does nothing and the copy
//     still happens from what is already exiled. "Exiled with this
//     enchantment" is read back off the event log by the trigger's label
//     (b27ExiledWith), so a card exiled by another effect is not in it,
//     and an Arcane Bombardment that leaves and returns starts again from
//     none (CR 400.7).
//   - A card that a replacement sent somewhere else, or that left exile
//     before the copy, is not in the set (the log record drops it).
//
// The free casts are grants taken with ordinary cast_spell actions after
// the trigger resolves, in whatever order the caster likes; see the
// CopyAndCast file for the window.
func init() {
	Register(Spec{
		OracleID:     "1e36a42f-58e2-49bc-8c14-7f75e1d8ef98",
		Name:         "Arcane Bombardment",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast,
				AllOf(YouCast(Or(Instant(), Sorcery())), youCastYourFirstInstantOrSorceryEachTurn),
				arcaneBombardmentLabel,
				arcaneBombardmentResolve),
		},
	})
}

// arcaneBombardmentLabel is the trigger's stack label: the key its
// exiled-with record is read by.
const arcaneBombardmentLabel = "Arcane Bombardment — exile an instant or sorcery card at random, then copy the exiled cards"

// youCastYourFirstInstantOrSorceryEachTurn is the "first ... each turn"
// half of the condition; YouCast supplies the "instant or sorcery spell"
// half. The tally is bumped before EventCast, so one means this spell is
// the first.
func youCastYourFirstInstantOrSorceryEachTurn(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.Kind == game.EventCast && ev.Actor == source.Controller &&
		g.CastTallyFor(ev.Actor).InstantSorcery == 1
}

// arcaneBombardmentResolve exiles one card at random, then copies
// everything exiled with the enchantment.
func arcaneBombardmentResolve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	pick := g.ChooseAtRandomForEffect(randomDraw(ctx), graveyardIDs(g, ctx.Controller(), func(c game.Card) bool {
		return c.IsInstant() || c.IsSorcery()
	}), 1)
	if len(pick) == 0 {
		return arcaneBombardmentCopy(ctx)
	}
	return g.ExileCardKeyedForEffect(pick[0], ctx.ItemTail(arcaneBombardmentExiled))
}

// arcaneBombardmentCopy is "copy each card exiled with this enchantment.
// You may cast any number of the copies without paying their mana costs."
func arcaneBombardmentCopy(ctx *Context) error {
	var set []uuid.UUID
	set = append(set, b27ExiledWith(ctx.Game, ctx.Source(), arcaneBombardmentLabel)...)
	return CopyAndCast{Cards: set, Free: true}.Apply(ctx)
}

// arcaneBombardmentExiled is "then copy the exiled cards" after the exile (#4253 slice 9).
var arcaneBombardmentExiled = CardTail("arcane-bombardment/copy-exiled", func(ctx *Context, o game.TailOutcome) error {
	g, item := ctx.Game, ctx.Item

	return arcaneBombardmentCopy(NewContext(g, item))
})
