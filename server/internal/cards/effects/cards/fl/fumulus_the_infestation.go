package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fumulus, the Infestation — Legendary Creature — Vampire Insect {3}{B},
// 2/2 (Foundations Jumpstart):
//
//	"Flying, deathtouch
//	 Whenever a player sacrifices a nontoken creature, create a 1/1 black
//	 Insect creature token with flying.
//	 Whenever an Insect, Leech, Slug, or Worm you control attacks,
//	 defending player loses 1 life and you gain 1 life."
//
// The sacrifice trigger is It That Betrays's read of EventSacrifice (emitted
// while the permanent is still on the battlefield, so its card and token-ness
// are known), for ANY player and only for a nontoken creature; the Insect is
// created under Fumulus's controller. The attack trigger fires once per
// attacking creature of those four types (each attacks, each triggers); the
// types are words of the source's text (CR 612.2). "Defending player" is the
// player attacked, or the controller of the planeswalker or battle attacked
// (CR 506.3d), so the drain is judged by b17DefendingPlayer, not by whether a
// player was attacked directly.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c75f8365-71f1-415f-8bbc-9bacb8ddba0b",
		Name:            "Fumulus, the Infestation",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventSacrifice, s82BnsB51ANontokenCreatureWasSacrificed,
				"Fumulus, the Infestation — create a 1/1 black Insect creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 black Insect with flying"), N: 1})),
			effects.On(game.EventAttack, s82BnsB51InsectLeechSlugOrWormYouControlAttacks,
				"Fumulus, the Infestation — defending player loses 1 life and you gain 1 life",
				s82BnsB51FumulusDrain),
		},
	})
}

func s82BnsB51ANontokenCreatureWasSacrificed(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventSacrifice || ev.CardID == uuid.Nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && !c.IsToken()
}

func s82BnsB51InsectLeechSlugOrWormYouControlAttacks(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !effects.AttackDeclaredByYou(ev, source.Controller) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !c.IsCreature() || c.Controller != source.Controller {
		return false
	}
	for _, t := range []string{"Insect", "Leech", "Slug", "Worm"} {
		if c.HasSubtype(source.TextSubtype(t)) {
			return true
		}
	}
	return false
}

// s82BnsB51FumulusDrain is "defending player loses 1 life and you gain 1
// life": the defending player is named by the attack event the trigger came
// from, and the life gain happens whether or not the loss did.
func s82BnsB51FumulusDrain(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if item.Trigger == nil {
		return nil
	}
	if p := effects.B17DefendingPlayer(g, item.Trigger.Event); p != uuid.Nil {
		if err := (effects.GainLife{Player: p, Amount: -1}).Apply(ctx); err != nil {
			return err
		}
	}
	return effects.GainLife{Player: item.Controller, Amount: 1}.Apply(ctx)
}
