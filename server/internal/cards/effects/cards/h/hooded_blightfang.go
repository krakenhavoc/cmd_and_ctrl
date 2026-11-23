package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hooded Blightfang — Creature — Snake {2}{B}, 1/4 (Foundations Jumpstart,
// BNS batch 32):
//
//	"Deathtouch
//	 Whenever a creature you control with deathtouch attacks, each opponent
//	 loses 1 life and you gain 1 life.
//	 Whenever a creature you control with deathtouch deals damage to a
//	 planeswalker, destroy that planeswalker."
//
// Two triggers, each judged on the creature as the event happened: the
// attacker's deathtouch is read through the layered view as it is declared
// (the Blightfang itself counts), and a damage source that has since left
// is judged by its last known card. The attack drain is Acolyte of
// Aclazotz's (life loss then life gain) and fires once per deathtouch attacker. The second
// trigger names no target: it destroys the planeswalker the damage event
// names (any planeswalker, yours included), and nothing happens if it has
// already left (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "342eb666-838c-43d0-bf52-2f704f2a80e3",
		Name:            "Hooded Blightfang",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, hoodedBlightfangDeathtouchAttacks,
				"Hooded Blightfang — each opponent loses 1 life and you gain 1 life",
				effects.Do(effects.EachOpponentLosesLife{N: 1}, effects.GainLife{Amount: 1})),
			effects.On(game.EventDealDamage, hoodedBlightfangHitAPlaneswalker,
				"Hooded Blightfang — destroy that planeswalker", hoodedBlightfangDestroy),
		},
	})
}

// hoodedBlightfangYourDeathtouchCreature reports whether `id` is a creature
// `source`'s controller controls that has deathtouch.
func hoodedBlightfangYourDeathtouchCreature(source *game.Card, id game.Card, ok bool) bool {
	return ok && id.IsCreature() && id.Controller == source.Controller && game.HasKeyword(&id, "deathtouch")
}

func hoodedBlightfangDeathtouchAttacks(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !effects.AttackDeclaredByYou(ev, source.Controller) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return hoodedBlightfangYourDeathtouchCreature(source, c, ok)
}

func hoodedBlightfangHitAPlaneswalker(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || ev.Amount <= 0 {
		return false
	}
	victim, ok := g.LookupCardForEffect(ev.Target)
	if !ok || !victim.IsPlaneswalker() || !g.Battlefield.Contains(ev.Target) {
		return false
	}
	dealer, ok := g.LookupCardForEffect(ev.Source)
	return hoodedBlightfangYourDeathtouchCreature(source, dealer, ok)
}

func hoodedBlightfangDestroy(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	id := item.Trigger.Event.Target
	if !effects.OnBattlefield(g, id) {
		return nil
	}
	return effects.DestroyTarget{Target: id}.Apply(effects.NewContext(g, item))
}
