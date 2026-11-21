package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Archmage's Newt — Creature — Salamander Mount {1}{U}:
//
//	"Whenever this creature deals combat damage to a player, target
//	 instant or sorcery card in your graveyard gains flashback until end
//	 of turn. The flashback cost is equal to its mana cost. That card
//	 gains flashback {0} until end of turn instead if this creature is
//	 saddled.
//	 Saddle 3"
//
// Whether the Newt is saddled is read when the trigger is built, off
// the live permanent, and carried on the item: a Newt removed in
// response is judged as it last existed (CR 608.2h). Saddle is sorcery
// speed, so the designation cannot change while the trigger waits.
//
// No simplification.
func init() {
	const label = "Archmage's Newt — target instant or sorcery card in your graveyard gains flashback"
	effects.Register(effects.Spec{
		OracleID:     "75ffebc4-8db9-4de6-a330-e3f41cdccecc",
		Name:         "Archmage's Newt",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Saddle(3)},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			Key:     label,
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Source == source.InstanceID && effects.CombatDamageToPlayerBy(ev, source.Controller, g)
			},
			Targets: effects.TargetCardInGraveyard("target instant or sorcery card in your graveyard",
				effects.YouOwn(), effects.Or(effects.Instant(), effects.Sorcery())),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, label)
				if source.Saddled {
					item.Params.Amount = 1
				}
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				ts := ctx.LegalTargets()
				if len(ts) == 0 {
					return nil
				}
				grant := effects.GrantFlashbackToCard{Target: ts[0].ID, Label: "Flashback — its mana cost (Archmage's Newt)"}
				if item.Params.Amount == 1 {
					grant.Cost = "{0}"
					grant.Label = "Flashback {0} (Archmage's Newt, saddled)"
				}
				return grant.Apply(ctx)
			},
		}},
	})
}
