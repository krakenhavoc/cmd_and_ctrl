package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Summon: Kujata — Enchantment Creature — Saga Ox {5}{R}, 7/5 (#3656):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Lightning — This creature deals 3 damage to each of up to two
//	 target creatures.
//	 II — Ice — Up to three target creatures can't block this turn.
//	 III — Fire — Discard a card, then draw two cards. When you discard
//	 a card this way, this creature deals damage equal to that card's
//	 mana value to each opponent.
//	 Trample, haste"
//
// A Saga creature (Summon: Ixion's chapter shape). Chapter I deals its
// damage from the Saga creature itself. Chapter II is Ahn-Crop
// Crusher's can't-block restriction on each pick. Chapter III is a
// rummage: the discard is the controller's choice and the draw follows
// it; the reflexive trigger (CR 603.12) is created when a card was
// discarded and carries that card's mana value, read as it left the
// hand, so it fires after the chapter finishes. An empty hand discards
// nothing, still draws two and creates no trigger.
//
// No simplification.
func init() {
	const name = "Summon: Kujata"
	Register(Spec{
		OracleID:        "9fe77f70-85cf-4a18-9451-2fad4c341f07",
		Name:            name,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, SagaChapterLabel(name, 1, "Lightning — 3 damage to each of up to two target creatures"),
				TargetCreature("up to two target creatures").WithCount(0, 2),
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return g.DamageInstanceForEffect(func() error {
						for _, t := range ctx.LegalTargets() {
							if t.Kind != game.TargetCard {
								continue
							}
							if err := (DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 3}).Apply(ctx); err != nil {
								return err
							}
						}
						return nil
					})
				}),
			ChapterTriggerTargeting(2, SagaChapterLabel(name, 2, "Ice — up to three target creatures can't block this turn"),
				TargetCreature("up to three target creatures").WithCount(0, 3),
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (RestrictUntilEOT{Target: t.ID, Restrictions: game.CantBlock, Label: name + " — can't block"}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "Fire — discard a card, then draw two cards; damage to each opponent equal to the discarded card's mana value"),
				summonKujataFire),
		},
	})
}

// summonKujataFire is chapter III: discard, then draw two, then the
// "when you discard a card this way" trigger.
func summonKujataFire(g *game.Game, item *game.StackItem) error {
	controller := item.Controller
	finish := func(g *game.Game, discarded []uuid.UUID) error {
		if err := g.DrawNForEffect(controller, 2); err != nil {
			return err
		}
		if len(discarded) == 0 {
			return nil
		}
		card, ok := g.LookupCardForEffect(discarded[0])
		if !ok {
			return nil
		}
		return ReflexiveTrigger{
			Label:  "Summon: Kujata — damage to each opponent equal to the discarded card's mana value",
			Cards:  discarded,
			Body:   summonKujataFireBody,
			Params: game.EffectParams{Amount: card.ManaValue()},
		}.Apply(NewContext(g, item))
	}
	if p := g.PlayerByIDForEffect(controller); p == nil || p.Hand.Size() == 0 {
		return finish(g, nil)
	}
	g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
		Player: controller,
		Source: item.SourceCardID,
		N:      1,
		Then: func(g *game.Game, _ uuid.UUID, discarded []uuid.UUID) error {
			return finish(g, discarded)
		},
	})
	return nil
}

// summonKujataFireEffect is the reflexive half: Kujata deals damage
// equal to the discarded card's mana value (Params.Amount) to each
// opponent.
func summonKujataFireEffect(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	if p.Amount <= 0 {
		return nil
	}
	ctx := NewContext(g, item)
	return g.DamageInstanceForEffect(func() error {
		for _, opp := range ctx.Opponents() {
			if err := (DealDamage{Source: item.SourceCardID, Target: opp, Amount: p.Amount}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}
