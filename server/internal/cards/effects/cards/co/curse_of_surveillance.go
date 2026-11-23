package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Curse of Surveillance — Enchantment — Aura Curse {4}{U}:
//
//	"Enchant player
//	 At the beginning of enchanted player's upkeep, any number of target
//	 players other than that player each draw cards equal to the number
//	 of Curses attached to that player."
//
// Court of Cunning's "any number of target players" (the (0, 0) count),
// with the enchanted player barred by a clause built per trigger from the
// event (TargetsFrom): the player whose upkeep it is. The number is read
// as the ability resolves, so every Curse then attached to that player
// counts, this one included, and a Curse that left in response does not.
// Each chosen player draws in turn order (CR 101.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4fd061f2-7269-4eea-bc7c-95a123893efa",
		Name:         "Curse of Surveillance",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPlayer(),
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventBeginUpkeep},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return source.AttachedTo.Kind == game.TargetPlayer && source.AttachedTo.ID == ev.Actor
				},
				Key: "Curse of Surveillance — any number of target players other than that player each draw a card for each Curse attached to that player",
				TargetsFrom: func(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
					enchanted := tc.Event.Actor
					return effects.TargetPlayer("any number of target players other than that player",
						func(_ *game.Game, _ uuid.UUID, p *game.Player) bool { return p.ID != enchanted }).WithCount(0, 0)
				},
				Effect: curseOfSurveillanceDraw,
			},
		},
	})
}

// curseOfSurveillanceDraw has each chosen player draw a card for each Curse
// attached to the enchanted player, counted now.
func curseOfSurveillanceDraw(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	enchanted := item.Trigger.Event.Actor
	n := cursesAttachedToPlayer(g, enchanted)
	if n <= 0 {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer || t.ID == enchanted {
			continue
		}
		if err := (effects.DrawCards{Player: t.ID, N: n}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// cursesAttachedToPlayer counts the Curses on the battlefield attached to a
// player. Caller holds g.mu.
func cursesAttachedToPlayer(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.AttachedTo.Kind == game.TargetPlayer && c.AttachedTo.ID == player && c.HasSubtype("Curse") {
			n++
		}
	}
	return n
}
