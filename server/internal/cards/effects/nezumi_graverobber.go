package effects

import (
	"github.com/google/uuid"

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
const nezumiGraverobberOracleID = "1373fa8d-0a39-4451-a482-ce2071438dc8"

func init() {
	Register(Spec{
		OracleID:     nezumiGraverobberOracleID,
		Name:         "Nezumi Graverobber",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}: Exile target card from an opponent's graveyard. If no cards are in that graveyard, flip this creature.",
			Cost:    ManaCost("{1}{B}"),
			Targets: TargetCardInGraveyard("target card from an opponent's graveyard", Not(YouOwn())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				card, ok := g.LookupCardForEffect(id)
				if !ok {
					return nil
				}
				graveyardOwner := card.Owner
				return ExileTarget{
					Target:   id,
					ThenTail: withCarry(ctx.ItemTail(nezumiGraverobberExiled), graveyardOwner),
				}.Apply(ctx)
			},
		}},
	})
	Register(Spec{
		OracleID:     nezumiGraverobberOracleID + "#1",
		Name:         "Nighteyes the Desecrator",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{4}{B}: Put target creature card from a graveyard onto the battlefield under your control.",
			Cost:    ManaCost("{4}{B}"),
			Targets: TargetCardInGraveyard("target creature card from a graveyard", Creature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if id, ok := b16FirstLegalTargetCard(ctx); ok {
					return ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}

// nezumiGraverobberExiled is the clause after the exile (#4253 slice 9).
var nezumiGraverobberExiled = CardTail("nezumi-graverobber/after-exile", func(ctx *Context, o game.TailOutcome) error {
	item, graveyardOwner := ctx.Item, tailCarry0(o)

	return flipThisWhen(ctx.Game, item, func(g *game.Game, _ uuid.UUID) bool {
		p := g.PlayerByIDForEffect(graveyardOwner)
		return p != nil && p.Graveyard != nil && p.Graveyard.Size() == 0
	})
})
