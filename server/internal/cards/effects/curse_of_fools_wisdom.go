package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Curse of Fool's Wisdom — Enchantment — Aura Curse {4}{B}{B}:
//
//	"Enchant player
//	 Whenever enchanted player draws a card, they lose 2 life and you
//	 gain 2 life.
//	 Madness {3}{B}"
//
// Curse of Opulence's enchant-player shape (Card.AttachedTo names the
// seat) watching EventDrawCard, which fires once per card, so a
// seven-card draw is seven triggers of 2 life each, as printed. The
// drawer is read off the event the trigger fired on, so it is the
// enchanted player who loses the life; the Curse's controller gains
// 2 whether or not the loss went through.
//
// Madness on an Aura is the ordinary madness cast: the discarded Curse
// waits in exile, and casting it for {3}{B} asks for its player target
// as any cast of it does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "93b2a685-8267-4cbd-ac55-db6ff95fe98d",
		Name:         "Curse of Fool's Wisdom",
		Completeness: CompletenessFull,
		Targets:      EnchantPlayer(),
		Madness:      "{3}{B}",
		Triggered: []game.TriggeredAbility{
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source.AttachedTo.Kind == game.TargetPlayer && source.AttachedTo.ID == ev.Actor
			}, "Curse of Fool's Wisdom — they lose 2 life and you gain 2 life",
				func(g *game.Game, item *game.StackItem) error {
					drawer := item.Trigger.Event.Actor
					if g.PlayerByIDForEffect(drawer) != nil {
						if err := g.ChangePlayerLifeForEffect(item.SourceCardID, drawer, -2); err != nil {
							return err
						}
					}
					return GainLife{Player: item.Controller, Amount: 2}.Apply(NewContext(g, item))
				}),
		},
	})
}
