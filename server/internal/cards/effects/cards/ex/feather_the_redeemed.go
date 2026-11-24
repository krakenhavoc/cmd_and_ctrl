package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Feather, the Redeemed — Legendary Creature — Angel {R}{W}{W}, 3/4 (Secret
// Lair Drop, #5462):
//
//	"Flying
//	 Whenever you cast an instant or sorcery spell that targets a creature
//	 you control, exile that card instead of putting it into your graveyard
//	 as it resolves. If you do, return it to your hand at the beginning of
//	 the next end step."
//
// The trigger marks the spell on the stack (Game.MarkSpellExilesAndReturnsForEffect,
// Lilah's mark with a return), and the mark is the spell's own resolution exit
// (resolution_exits.go): a countered or fizzled spell goes to the graveyard
// as usual, and a spell that resolves is exiled, and a delayed trigger returns
// that exiled object to its owner's hand at the beginning of the next end step
// (CR 603.7, 400.7). "Targets a creature you control" is read as the spell is
// cast: any one of its targets, a creature on the battlefield its caster controls.
// A copy is not a card, so it is never marked.
//
// No simplification.
func init() {
	const label = "Feather, the Redeemed — exile that card instead of putting it into your graveyard as it resolves; return it to your hand at the beginning of the next end step"
	mark := effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery()), label,
		func(g *game.Game, item *game.StackItem) error {
			ev := effects.NewContext(g, item).Trigger().Event
			if ev.CardID == uuid.Nil {
				return nil
			}
			g.MarkSpellExilesAndReturnsForEffect(ev.CardID)
			return nil
		})
	inner := mark.AppliesTo
	mark.AppliesTo = func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
		return inner(ev, source, ch, g) && featherTargetsYourCreature(ev, source, g)
	}
	effects.Register(effects.Spec{
		OracleID:        "aa219936-661b-4ccb-8741-78b70cff2b1a",
		Name:            "Feather, the Redeemed",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{mark},
	})
}

// featherTargetsYourCreature reports whether the spell on the stack targets a
// creature on the battlefield its caster controls.
func featherTargetsYourCreature(ev game.Event, source *game.Card, g *game.Game) bool {
	item := g.StackItemForEffect(ev.CardID)
	if item == nil {
		return false
	}
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard {
			continue
		}
		if c, ok := g.LookupCardForEffect(t.ID); ok && c.IsCreature() && effects.OnBattlefield(g, t.ID) && c.Controller == source.Controller {
			return true
		}
	}
	return false
}
