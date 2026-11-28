package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b790_helpers.go — Visions, batch 790 (S82 older sets, #5631).

// Old790CreatureEnteredOfTypes is the "Whenever a <type> or <type> enters"
// condition (any controller): the entering permanent has one of the creature
// types, each read as a word of the source's text (CR 612.2).
func Old790CreatureEnteredOfTypes(types ...string) When {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if ev.Kind != game.EventETB || ev.CardID == uuid.Nil {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		if !ok {
			return false
		}
		for _, t := range types {
			if c.HasSubtype(source.TextSubtype(t)) {
				return true
			}
		}
		return false
	}
}

// Old790DestroyTheEnteredCreature is "destroy it. It can't be regenerated"
// for the permanent whose entry triggered the ability, if it is still on the
// battlefield as the trigger resolves.
func Old790DestroyTheEnteredCreature(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || item.Trigger.Event.CardID == uuid.Nil {
		return nil
	}
	id := item.Trigger.Event.CardID
	if !OnBattlefield(g, id) {
		return nil
	}
	return DestroyTarget{Target: id, CantBeRegenerated: true}.Apply(NewContext(g, item))
}

// Old790DestroyAllOfTypes is "destroy all <type> and <type>. They can't be
// regenerated", the types read as words of the resolving ability's text.
func Old790DestroyAllOfTypes(types ...string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		words := ctx.TextSubtypes(types...)
		return DestroyAllMatching{
			Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
				for _, w := range words {
					if c.HasSubtype(w) {
						return true
					}
				}
				return false
			},
			CantBeRegenerated: true,
		}.Apply(ctx)
	}
}

// Old790DeathWatch is "its controller loses life equal to its power and you
// gain life equal to its toughness" for the enchanted creature that died,
// both read as it last existed on the battlefield (CR 608.2h). The controller
// is the one it had as it died.
func Old790DeathWatch(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	info, ok := g.LastKnownPermanentForEffect(item.Trigger.Event.CardID)
	if !ok {
		return nil
	}
	ctx := NewContext(g, item)
	if info.Power > 0 {
		if err := (GainLife{Player: info.Controller, Amount: -info.Power}).Apply(ctx); err != nil {
			return err
		}
	}
	if info.Toughness > 0 {
		return GainLife{Player: item.Controller, Amount: info.Toughness}.Apply(ctx)
	}
	return nil
}

// Old790AllNonWallCreaturesYouControlAttacked is Mob Mentality's condition:
// the attack the source's controller declared has every non-Wall creature
// they control attacking, and at least one. It is paired with OncePerBatch.
func Old790AllNonWallCreaturesYouControlAttacked(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !AttackDeclaredByYou(ev, source.Controller) {
		return false
	}
	wall := source.TextSubtype("Wall")
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != source.Controller || !c.IsCreature() || c.HasSubtype(wall) {
			continue
		}
		if c.AttackingTarget == uuid.Nil {
			return false
		}
		n++
	}
	return n > 0
}

// Old790EnchantedGetsPowerPerAttacker is "enchanted creature gets +X/+0 until
// end of turn, where X is the number of attacking creatures", counted as the
// trigger resolves.
func Old790EnchantedGetsPowerPerAttacker(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		host := EnchantedCreatureOf(ctx)
		if host == uuid.Nil {
			return nil
		}
		return BoostUntilEOT{Target: host, Power: len(AttackingCreatures(g)), Label: label}.Apply(ctx)
	}
}

// Old790YouDrawACard is Betrayal's body: you draw a card.
func Old790YouDrawACard(g *game.Game, item *game.StackItem) error {
	return DrawCards{N: 1}.Apply(NewContext(g, item))
}

// Old790KnightOfTheMistsDecline is the keyed decline of Knight of the Mists's
// "you may pay {U}": unpaid, the targeted Knight is destroyed and can't be
// regenerated. An on-disk key: never renamed, never reused.
var Old790KnightOfTheMistsDecline = PayUnlessElse("s82-old-b790/knight-of-the-mists-destroy-target-knight",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		if len(r.Carry) == 0 || !OnBattlefield(ctx.Game, r.Carry[0]) {
			return nil
		}
		return DestroyTarget{Target: r.Carry[0], CantBeRegenerated: true}.Apply(ctx)
	})

// Old790KnightOfTheMistsETB asks the Knight's controller to pay {U}; if they
// don't, the target Knight is destroyed. A target that is no longer legal by
// resolution does nothing, and nobody is asked (CR 608.2b).
func Old790KnightOfTheMistsETB(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id := FirstLegalCardTarget(ctx)
	if id == uuid.Nil {
		return nil
	}
	return PayUnless{
		Chooser:  ctx.Controller(),
		Cost:     "{U}",
		Question: "Knight of the Mists — pay {U}? If you don't, destroy target Knight and it can't be regenerated.",
		Else:     Old790KnightOfTheMistsDecline,
		Carry:    []uuid.UUID{id},
	}.Apply(ctx)
}
