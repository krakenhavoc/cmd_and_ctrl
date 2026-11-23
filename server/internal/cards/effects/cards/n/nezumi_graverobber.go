package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nezumi Graverobber // Nighteyes the Desecrator — {1}{B} Creature — Rat
// Rogue 2/1 // Legendary Creature — Rat Wizard 4/2, a flip card (#3216,
// ADR 0151):
//
//	Top:     "{1}{B}: Exile target card from an opponent's graveyard.
//	          If no cards are in that graveyard, flip this creature."
//	Flipped: "{4}{B}: Put target creature card from a graveyard onto
//	          the battlefield under your control."
//
// "That graveyard" is the owner of the card that was targeted, read
// before the exile; the emptiness is judged once the exile has settled
// (ExileTarget.ThenTail). If the target is gone as the ability resolves it
// does nothing at all, so it cannot flip. Nighteyes reanimates from any
// graveyard under the activator's control, and is black and {1}{B}
// (CR 710.1c).
//
// No simplification.
const NezumiGraverobberOracleID = "1373fa8d-0a39-4451-a482-ce2071438dc8"

func init() {
	effects.Register(effects.Spec{
		OracleID:     NezumiGraverobberOracleID,
		Name:         "Nezumi Graverobber",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}: Exile target card from an opponent's graveyard. If no cards are in that graveyard, flip this creature.",
			Cost:    effects.ManaCost("{1}{B}"),
			Targets: effects.TargetCardInGraveyard("target card from an opponent's graveyard", effects.Not(effects.YouOwn())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				card, ok := g.LookupCardForEffect(id)
				if !ok {
					return nil
				}
				graveyardOwner := card.Owner
				return effects.ExileTarget{
					Target:   id,
					ThenTail: effects.WithCarry(ctx.ItemTail(nezumiGraverobberExiled), graveyardOwner),
				}.Apply(ctx)
			},
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     NezumiGraverobberOracleID + "#1",
		Name:         "Nighteyes the Desecrator",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{B}: Put target creature card from a graveyard onto the battlefield under your control.",
			Cost:    effects.ManaCost("{4}{B}"),
			Targets: effects.TargetCardInGraveyard("target creature card from a graveyard", effects.Creature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if id, ok := effects.B16FirstLegalTargetCard(ctx); ok {
					return effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}

// nezumiGraverobberExiled is the clause after the exile (#4253 slice 9).
var nezumiGraverobberExiled = effects.CardTail("nezumi-graverobber/after-exile", func(ctx *effects.Context, o game.TailOutcome) error {
	item, graveyardOwner := ctx.Item, effects.TailCarry0(o)

	return effects.FlipThisWhen(ctx.Game, item, func(g *game.Game, _ uuid.UUID) bool {
		p := g.PlayerByIDForEffect(graveyardOwner)
		return p != nil && p.Graveyard != nil && p.Graveyard.Size() == 0
	})
})
