package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// SnowManaOfASpellsColorWasSpent is "{S} of any of that spell's colors was
// spent to cast it" (Boreal Outrider), read off the EventCast: some mana
// of the spell's payment came from a snow source and was one of the spell's
// colors (CR 107.4h). The record was made as the mana was spent, so a snow
// land that has left the battlefield still counts.
func SnowManaOfASpellsColorWasSpent(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	spell, ok := g.LookupCardForEffect(ev.CardID)
	if !ok {
		return false
	}
	colors := spell.EffectiveColors()
	for _, t := range g.StackItemPaidForEffect(ev.CardID).Spent().Tokens() {
		if t.SourceKinds.HasAny(game.ManaSourceSnow) && slices.Contains(colors, t.Color) {
			return true
		}
	}
	return false
}

// ThisWasTheNthSpellYouCast is "if it was the Nth spell you cast this turn"
// for a permanent's own entry trigger (Codespell Cleric): the ordinal its
// spell had when it was cast (CR 601.2i, CastProvenance.CastOrdinal), cast
// by the player who controls it now. A permanent that was not cast, or was
// put onto the battlefield some other way, was not a spell you cast.
func ThisWasTheNthSpellYouCast(n int) When {
	return func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		if source == nil || source.Provenance.CastOrdinal != n {
			return false
		}
		cast, known := source.CastByItsController()
		return known && cast
	}
}

// ResplendentMarshalCandidates are the creature cards in `player`'s
// graveyard other than `self`: "another creature card from your graveyard".
func ResplendentMarshalCandidates(ctx *Context, player, self uuid.UUID) []uuid.UUID {
	return GraveyardCardIDs(ctx, player, func(c game.Card) bool {
		return c.InstanceID != self && c.IsCreature()
	})
}

// sourceGainsUntilEndOfTurn is "it gains <keyword> until end of turn" as a
// step of a Do, for the source of the ability (Infernal Pet, Bloodsky
// Berserker).
func SourceGainsUntilEndOfTurn(keyword, label string) func(ctx *Context) error {
	gain := ThisGainsKeywordUntilEndOfTurn(keyword, label)
	return func(ctx *Context) error { return gain(ctx.Game, ctx.Item) }
}
