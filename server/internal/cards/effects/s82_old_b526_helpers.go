package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b526_helpers.go — Eventide, S82 older sets batch 526 (#5590): the
// Mimics' cast trigger and the "another <colour> spell this turn" test.

// S82OldB526Mimic is the Mimic cycle's one triggered ability: "Whenever you
// cast a spell that's both <colour A> and <colour B>, this creature has base
// power and toughness P/T until end of turn" plus either "and can't be
// blocked this turn" (blockable false) or "and gains <keyword> until end of
// turn". The spell's colours are read as it was cast (CR 601.2i); the Mimic
// is the object that was on the battlefield as the trigger resolves, so one
// that left does nothing (CR 400.7).
func S82OldB526Mimic(name, colorA, colorB string, power, toughness int, unblockable bool, keyword string) game.TriggeredAbility {
	both := And(OfColor(colorA), OfColor(colorB))
	label := name + " — base power and toughness " + fmt.Sprintf("%d/%d", power, toughness) + " until end of turn"
	return WheneverYouCast(both, label, func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		if err := UntilEndOfTurn(ctx, item.SourceCardID, nil, name+" — base power and toughness until end of turn",
			game.SetBasePTMods(power, toughness)...); err != nil {
			return err
		}
		if unblockable {
			return RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked, Label: name + " — can't be blocked this turn"}.Apply(ctx)
		}
		return GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{keyword}, Label: name + " — gains " + keyword + " until end of turn"}.Apply(ctx)
	})
}

// S82OldB526CastAnotherOfColor is "you've cast another <colour> spell this
// turn" for the player and object named: the spell or permanent itself is
// not "another" (CR 700.2), so a Dream Thief cast as a blue spell does not
// count itself, and one put onto the battlefield without being cast counts
// every blue spell cast. Dream Thief and Soul Reap.
func S82OldB526CastAnotherOfColor(ctx *Context, player, self uuid.UUID, color string) bool {
	return ctx.Game.CastAnotherSpellOfColorThisTurnForEffect(player, self, color)
}
