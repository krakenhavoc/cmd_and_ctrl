package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bioplasm — Creature — Ooze {3}{G}{G}, 4/4 (Guildpact):
//
//	"Whenever this creature attacks, exile the top card of your library. If
//	 it's a creature card, this creature gets +X/+Y until end of turn, where
//	 X is the exiled creature card's power and Y is its toughness."
//
// The top card is exiled through the CR 614 window (Riverwheel Sweep's keyed
// exile), and once it has landed in exile its power and toughness are read
// as that card: a creature card whose power or toughness is a * counts as 0
// in exile (CR 208.2b, 604.3). The exiled card stays in exile whatever it
// is. An empty library exiles nothing. If the Bioplasm has left the
// battlefield by then, it gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f39777be-39af-4a97-a70d-824d713bf145",
		Name:         "Bioplasm",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Bioplasm — exile the top card of your library; if it's a creature card, +X/+Y until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return g.ExileTopKeyedForEffect(item.Controller, item.Controller, 1, ctx.ItemTail(bioplasmExiled))
				}),
		},
	})
}

// bioplasmExiled is the pump, once the card has landed in exile. An on-disk
// key: never renamed, never reused.
var bioplasmExiled = effects.CardTail("bioplasm/exiled-creature-pump", func(ctx *effects.Context, o game.TailOutcome) error {
	var exiled uuid.UUID
	for _, id := range o.IDs {
		if z := ctx.Game.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneExile {
			exiled = id
		}
	}
	if exiled == uuid.Nil {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(exiled)
	if !ok || !c.IsCreature() {
		return nil
	}
	self := ctx.Source()
	if !effects.OnBattlefield(ctx.Game, self) {
		return nil
	}
	return effects.BoostUntilEOT{Target: self, Power: c.Power, Toughness: c.Toughness, Label: "Bioplasm — +X/+Y"}.Apply(ctx)
})
