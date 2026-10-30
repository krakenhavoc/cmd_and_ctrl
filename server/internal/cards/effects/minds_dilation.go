package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mind's Dilation — {5}{U}{U} Enchantment (Marvel Super Heroes, #4455):
//
//	"Whenever an opponent casts their first spell each turn, that player
//	 exiles the top card of their library. If it's a nonland card, you may
//	 cast it without paying its mana cost."
//
// The trigger is The Lord of Pain's first-spell test (the per-turn cast
// tally already counts the spell that fired it), narrowed to opponents. The
// exile is the opponent's own library, a one-card mill to exile whose tail is
// Amped Raptor's: it grants the free cast to the card if it is a nonland card
// still in exile, and offers it during the resolution (CR 608.2g), so a
// declined card stays in exile and a land is never offered. The cast is by
// Mind's Dilation's controller, not the opponent who owns the card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d4b5e09b-9ca2-4b51-a4f8-080741810df8",
		Name:         "Mind's Dilation",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, mindsDilationFirstSpell,
				"Mind's Dilation — that player exiles the top card of their library; you may cast it without paying its mana cost",
				mindsDilationExile),
		},
	})
}

// mindsDilationFirstSpell is "an opponent casts their first spell each turn".
func mindsDilationFirstSpell(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.Actor != uuid.Nil && ev.Actor != source.Controller && g.CastTallyFor(ev.Actor).Total == 1
}

func mindsDilationExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	caster := ctx.Trigger().Event.Actor
	if caster == uuid.Nil {
		return nil
	}
	tail := ctx.ItemTail(exileUntilNonlandGranted)
	tail.Permission = &game.CastPermission{
		Cost:        "{0}",
		Timing:      game.TimingFlash,
		CastOnly:    true,
		Duration:    g.UntilEndOfTurnDuration(),
		LapseOnPass: game.LapseStaysInExile,
		Source:      ctx.Source(),
		SourceName:  "Mind's Dilation",
		Label:       "Mind's Dilation — cast it without paying its mana cost",
	}
	return g.MillToZoneKeyedForEffect(caster, 1, game.ZoneExile, game.MillUntil{}, tail)
}
