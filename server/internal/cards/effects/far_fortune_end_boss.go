package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Far Fortune, End Boss — Legendary Creature — Human Mercenary {2}{B}{R},
// 4/5:
//
//	"Start your engines!
//	 Whenever you attack, Far Fortune deals 1 damage to each opponent.
//	 Max speed — If a source you control would deal damage to an opponent
//	 or a permanent an opponent controls, it deals that much damage plus
//	 1 instead."
//
// ADR 0138 (#2122, #2711). "Whenever you attack" is one trigger per attack
// declaration (OncePerBatch, Adeline's shape), and the damage is Far
// Fortune's own, so its own max speed bonus applies to it. The max speed
// replacement is Torbran's shape with a +1: every source you control, to
// an opponent or an opponent's permanent, combat or not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8b9734c1-7185-40f6-8582-9f87109e3a09",
		Name:            "Far Fortune, End Boss",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{StartYourEngines},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			}, "Far Fortune, End Boss — 1 damage to each opponent", Do(DamageEachOpponent{N: 1}))),
		},
		Replacements: []game.ReplacementEffect{
			MaxSpeedReplacement(game.ReplacementEffect{
				Watches: []game.EventKind{game.EventDealDamage},
				AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
					if ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.DamageSource == uuid.Nil {
						return false
					}
					return damageSourceControlledBy(ev, g, src.Controller) && damageHitsAnOpponentOf(ev, g, src.Controller)
				},
				Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
					ev.DamageAmount++
					return nil
				},
				Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
					return src.Controller
				},
				Label: "Far Fortune, End Boss: +1 damage",
			}),
		},
	})
}
