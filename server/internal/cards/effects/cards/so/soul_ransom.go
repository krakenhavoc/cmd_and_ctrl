package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soul Ransom — Enchantment — Aura {2}{U}{B} (Gatecrash, S82 OLD b407,
// #5565):
//
//	"Enchant creature
//	 You control enchanted creature.
//	 Discard two cards: This Aura's controller sacrifices it, then draws two
//	 cards. Only your opponents may activate this ability."
//
// Control Magic's control change and Detention Vortex's opponents-only row
// (OpponentsOnly: only an opponent of the Aura's controller activates it, and
// pays the discard). The Aura's controller is read before the sacrifice
// (CR 608.2h), sacrifices it, and then draws two cards; a creature's controller
// who is not the Aura's controller never draws.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f4ae18b5-2cf1-4720-8237-40a0c7b040ae",
		Name:         "Soul Ransom",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static:       []game.StaticAbility{effects.ControlAttachedBySource()},
		Activated: []effects.ActivatedAbility{{
			Label:         "Discard two cards: This Aura's controller sacrifices it, then draws two cards. Only your opponents may activate this ability.",
			Purpose:       game.Purpose{Answers: game.AnswerValue},
			Cost:          effects.DiscardN(2, "two cards"),
			OpponentsOnly: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				aura, ok := ctx.SourcePermanent()
				if !ok {
					return nil
				}
				owner := aura.Controller
				if err := (effects.SacrificePermanent{Target: ctx.Source()}).Apply(ctx); err != nil {
					return err
				}
				return effects.DrawCards{Player: owner, N: 2}.Apply(ctx)
			},
		}},
	})
}
