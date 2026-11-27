package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Insubordination — Enchantment — Aura {B}{B} (Mercadian Masques, S82 OLD batch 685, #5619):
//
//	"Enchant creature
//	 At the beginning of the end step of enchanted creature's controller, this Aura deals 2 damage to that player unless that creature attacked this turn."
//
// One Thousand Lashes' "of enchanted creature's controller" watch on the end
// step: the player is the controller of the creature the Aura is attached to as
// the end step begins, read off the event. "That creature" is the one the Aura
// enchants as the trigger resolves (or last enchanted, if the Aura has left,
// CR 608.2h); if it attacked this turn nothing happens. The damage is the
// Aura's, dealt to that player.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0d3e843e-eeb4-421e-b55d-ceb21f51a080",
		Name:         "Insubordination",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginEndStep, insubordinationEndStep,
				"Insubordination — 2 damage to that player unless that creature attacked this turn",
				insubordinationDamage),
		},
	})
}

// insubordinationEndStep is the end step of whoever controls the creature this
// Aura is attached to now.
func insubordinationEndStep(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	host := g.AttachedHostOf(source)
	return host != nil && ev.Actor != uuid.Nil && ev.Actor == host.Controller
}

// insubordinationDamage deals 2 damage from the Aura to the end step's player
// unless the enchanted creature attacked this turn.
func insubordinationDamage(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ctx := effects.NewContext(g, item)
	host, ok := effects.EnchantedCreatureAsItResolves(ctx)
	if ok && g.AttackedThisTurn(host) {
		return nil
	}
	dmg := effects.DealDamage{Source: ctx.Source(), Target: item.Trigger.Event.Actor, Amount: 2}
	if ref, ok := ctx.SourceRef(); ok {
		dmg.SourceObject = &ref
	}
	return dmg.Apply(ctx)
}
