package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Agatha's Soul Cauldron — Legendary Artifact {2}:
//
//	"You may spend mana as though it were mana of any color to activate
//	 abilities of creatures you control.
//	 Creatures you control with +1/+1 counters on them have all
//	 activated abilities of all creature cards exiled with Agatha's
//	 Soul Cauldron.
//	 {T}: Exile target card from a graveyard. When a creature card is
//	 exiled this way, put a +1/+1 counter on target creature you
//	 control."
//
// The {T} exile is linked (ExileTarget.Linked, ADR 0155 §8). Once the card
// has landed in exile and is a creature card, a CR 603.12 reflexive
// trigger targets a creature you control for the counter; a noncreature
// card, or one a replacement sent elsewhere, makes none.
//
// The borrow lends the creature cards linked to this Cauldron object to
// each creature you control with a +1/+1 counter, read after layers 1-5
// (GrantActivatedAbilitiesOf over ExiledWithLenders): a creature that
// loses its last counter loses the rows the same pass (the counter bump).
//
// The spend rider covers every ability of a creature you control, its
// own, granted or lent (ADR 0155 Decision 7): an activation whose object
// is a creature controlled by the payer. Another player's creature with
// an ability any player may activate is not covered.
//
// No simplification.
func init() {
	const clause = "Creatures you control with +1/+1 counters on them have all activated abilities of all creature cards exiled with Agatha's Soul Cauldron."
	borrow := GrantActivatedAbilitiesOf(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return source != nil && target.IsCreature() && target.Controller == source.Controller &&
			target.Counters[game.CounterPlusOne] > 0
	}, game.AbilityBorrow{
		Label:   clause,
		Lenders: ExiledWithLenders(ByController(Creature())),
	})
	borrow.DependsOnExile = true
	Register(Spec{
		OracleID:     "c259e16f-2a44-4552-8678-815f757a02e8",
		Name:         "Agatha's Soul Cauldron",
		Completeness: CompletenessFull,
		AnyColorSpend: []game.AnyColorSpendStatic{{
			Label: "You may spend mana as though it were mana of any color to activate abilities of creatures you control.",
			Whose: game.AnyColorSpendYou,
			Covers: func(ctx game.ManaSpendContext) bool {
				return ctx.Purpose == game.SpendPurposeActivate && slices.Contains(ctx.Types, "Creature") &&
					ctx.Payer != uuid.Nil && ctx.Controller == ctx.Payer
			},
		}},
		Static: []game.StaticAbility{borrow},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Exile target card from a graveyard. When a creature card is exiled this way, put a +1/+1 counter on target creature you control.",
			Cost:    TapCost(),
			Targets: TargetCardInGraveyard("target card from a graveyard"),
			Effect:  agathasSoulCauldronExile,
		}},
	})
}

// agathasSoulCauldronExile is the {T} ability: the linked exile, then the
// reflexive counter trigger when what landed is a creature card.
func agathasSoulCauldronExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		id := t.ID
		return ExileTarget{Target: id, Linked: true, Then: func(ctx *Context, exiled bool) error {
			if !exiled {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok || !c.IsCreature() {
				return nil
			}
			return WhenYouDo("Agatha's Soul Cauldron — put a +1/+1 counter on target creature you control",
				agathasSoulCauldronCounterBody).Apply(ctx)
		}}.Apply(ctx)
	}
	return nil
}
