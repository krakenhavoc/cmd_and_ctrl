package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rory Williams — Legendary Creature — Human Soldier {W}{U}, 3/3 (Doctor Who,
// #5948):
//
//	"Partner with Amy Pond
//	 First strike, lifelink
//	 The Last Centurion — When you cast this spell from anywhere other than
//	 exile, exile it with three time counters on it. It gains suspend. Then
//	 investigate."
//
// Partner with is PartnerWith's trigger (and the deck-construction rule);
// first strike and lifelink are printed keywords. The cast trigger fires only
// for a cast from a zone other than exile (read off the cast event's origin
// zone, so a suspend cast from exile does not loop). It exiles the spell
// from the stack without countering it, with three time counters, and the
// card gains suspend (the exile tail, as Sinister Concierge), and then
// investigates whether or not the spell was still there to exile.
//
// No simplification.
func init() {
	t := effects.WhenYouCastThisSpell("Rory Williams — exile it with three time counters on it; it gains suspend; then investigate",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			if err := g.ExileSpellKeyedForEffect(item.SourceCardID, game.KeyedTail{
				Ref: game.GainSuspendExileTail, Player: ctx.Controller(), N: 3,
			}); err != nil {
				return err
			}
			return effects.Investigate{N: 1}.Apply(ctx)
		})
	self := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
		return ev.OldZone != game.ZoneExile && self(ev, source, ch, g)
	}
	effects.Register(effects.Spec{
		OracleID:        "16a44586-ea1d-45f1-b9de-da1c11f729df",
		Name:            "Rory Williams",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", "lifelink"},
		Purpose:         game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.PartnerWith("Rory Williams", "Amy Pond"),
			t,
		},
	})
}
