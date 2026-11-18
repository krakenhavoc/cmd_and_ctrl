package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wizard's Spellbook — Artifact — Book {5}{U}{U} (Adventures in the
// Forgotten Realms, S82):
//
//	"{T}: Exile target instant or sorcery card from a graveyard. Roll a d20.
//	 Activate only as a sorcery.
//	 1—9 | Copy that card. You may cast the copy.
//	 10—19 | Copy that card. You may cast the copy by paying {1} rather
//	 than paying its mana cost.
//	 20 | Copy each card exiled with this artifact. You may cast any number
//	 of the copies without paying their mana costs."
//
// The exile is linked to this Spellbook object (ExileTarget.Linked, CR
// 607.2a) and is the first sentence; the d20 follows it once the exile has
// settled (the exile's tail), whether or not the card landed in exile. Rows
// 1—9 and 10—19 copy "that card" (CR 707.12: a copy of the card, cast while
// the ability resolves, CR 608.2g; CopyAndCast) and offer a cast paying its
// mana cost or {1} instead. Row 20 copies every card exiled with this
// Spellbook object, the one just exiled included, and offers any number of
// free casts. A copy that is not cast ceases to exist.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "740fe41a-16bb-4c97-a03c-f678f15a0a0c",
		Name:         "Wizard's Spellbook",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{T}: Exile target instant or sorcery card from a graveyard. Roll a d20. Activate only as a sorcery.",
			Cost:         TapCost(),
			SorcerySpeed: true,
			Targets:      TargetCardInGraveyard("target instant or sorcery card from a graveyard", Or(Instant(), Sorcery())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return ExileTarget{Target: t.ID, Linked: true, ThenTail: ctx.ItemTail(wizardsSpellbookExiled)}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}

// wizardsSpellbookExiled rolls the d20 once the exile has settled, handing
// the table the card if it reached exile. An on-disk key.
var wizardsSpellbookExiled = CardTail("dice/wizards-spellbook-exiled", func(ctx *Context, o game.TailOutcome) error {
	return RollDice{Sides: 20, N: 1, Then: wizardsSpellbookRolled, TailCarry: append([]uuid.UUID(nil), o.IDs...)}.Apply(ctx)
})

// wizardsSpellbookRolled is the results table, after the d20. An on-disk key.
var wizardsSpellbookRolled = CardTail("dice/wizards-spellbook-rolled", func(ctx *Context, o game.TailOutcome) error {
	if len(o.Rolls) == 0 {
		return nil
	}
	g := ctx.Game
	r := o.Rolls[0]
	if r == 20 {
		ref, ok := ctx.SourceRef()
		if !ok {
			return nil
		}
		var ids []uuid.UUID
		for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
			ids = append(ids, c.InstanceID)
		}
		if len(ids) == 0 {
			return nil
		}
		return CopyAndCast{Cards: ids, Free: true}.Apply(ctx)
	}
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	id := o.Tail.Carry[0]
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
		return nil
	}
	if r >= 10 {
		return CopyAndCast{Cards: []uuid.UUID{id}, Cost: "{1}"}.Apply(ctx)
	}
	return CopyAndCast{Cards: []uuid.UUID{id}}.Apply(ctx)
})
