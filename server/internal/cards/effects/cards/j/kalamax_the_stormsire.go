package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kalamax, the Stormsire — Legendary Creature — Elemental Dinosaur
// {1}{G}{U}{R}, 4/4 (Special Guests, S82 BNS):
//
//	"Whenever you cast your first instant spell each turn, if Kalamax is
//	 tapped, copy that spell. You may choose new targets for the copy.
//	 Whenever you copy an instant spell, put a +1/+1 counter on Kalamax."
//
// The first ability is Double Vision's copy with two conditions. "First
// instant spell each turn" is the cast tally's Instants count, bumped before
// the cast event (so 1 means this spell is the first). "If Kalamax is
// tapped" is an intervening "if" (CR 603.4): checked as the spell is cast and
// again as the trigger resolves, so untapping in response removes the copy
// and a Kalamax that has left the battlefield is not tapped. The copy is made
// under your control from last-known information (CR 608.2h) with the
// new-targets choice (CR 707.10c).
//
// The second ability watches EventSpellCopied: any copy of an instant spell
// that you control, Kalamax's own included, whatever made it. A copy is not
// cast (CR 707.10), so it never triggers the first ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b8501a37-23e4-4873-80cb-d1c0f6c95155",
		Name:         "Kalamax, the Stormsire",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{Name: "Kalamax is tapped", Test: func(ctx *effects.Context) bool {
				return kalamaxIsTapped(ctx.Game, ctx.Item.SourceCardID)
			}}, effects.On(game.EventCast,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if ev.Kind != game.EventCast || ev.Actor != source.Controller || !source.Tapped {
						return false
					}
					spell, ok := g.LookupCardForEffect(ev.CardID)
					return ok && spell.IsInstant() && g.CastTallyFor(ev.Actor).Instants == 1
				},
				"Kalamax, the Stormsire — copy that spell; you may choose new targets for the copy",
				effects.CopyTheTriggeringSpellWithNewTargets)),
			effects.On(game.EventSpellCopied,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if ev.Kind != game.EventSpellCopied || ev.Actor != source.Controller {
						return false
					}
					spell, ok := g.LookupCardForEffect(ev.CardID)
					return ok && spell.IsInstant()
				},
				"Kalamax, the Stormsire — put a +1/+1 counter on Kalamax",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}

// kalamaxIsTapped is "Kalamax is tapped" as the trigger resolves: on the
// battlefield and tapped.
func kalamaxIsTapped(g *game.Game, id uuid.UUID) bool {
	z := g.FindCardZoneForEffect(id)
	if z == nil || z.Kind != game.ZoneBattlefield {
		return false
	}
	c, ok := g.LookupCardForEffect(id)
	return ok && c.Tapped
}
