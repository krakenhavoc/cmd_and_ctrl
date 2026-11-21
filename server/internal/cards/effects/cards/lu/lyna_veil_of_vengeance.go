package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lyna, Veil of Vengeance — Legendary Creature — Soltari Cleric {W}{U}{B},
// 2/3 (Mystery Booster Commander Edition):
//
//	"Shadow (This creature can block or be blocked by only creatures with
//	 shadow.)
//	 Whenever a creature you control with shadow deals combat damage to a
//	 player, draw a card.
//	 {3}: Put a shadow counter on another target creature. Activate only as
//	 a sorcery."
//
// The trigger is one event per creature (Black Widow's shape), with shadow
// read as the dealer has it (game.HasKeyword), so a shadow counter or a
// granted shadow counts, Lyna herself included. The shadow counter is a
// keyword counter (CR 122.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "82321fa9-7a95-4c9d-988e-2140eaf1a083",
		Name:            "Lyna, Veil of Vengeance",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"shadow"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, shadowCreatureYouControlHitAPlayer, "Lyna, Veil of Vengeance — draw a card",
				effects.Do(effects.DrawCards{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{3}: Put a shadow counter on another target creature. Activate only as a sorcery.",
			Cost:         effects.ManaCost("{3}"),
			SorcerySpeed: true,
			Targets:      effects.Another(effects.TargetCreature("another target creature")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return effects.NothingIfGone(effects.AddCounter{Target: t.ID, Kind: game.CounterShadow, N: 1}.Apply(ctx))
					}
				}
				return nil
			},
		}},
	})
}

// shadowCreatureYouControlHitAPlayer is the condition of "whenever a
// creature you control with shadow deals combat damage to a player".
func shadowCreatureYouControlHitAPlayer(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !effects.CombatDamageToPlayerBy(ev, source.Controller, g) {
		return false
	}
	dealer, ok := g.LookupCardForEffect(ev.Source)
	return ok && game.HasKeyword(&dealer, "shadow")
}
