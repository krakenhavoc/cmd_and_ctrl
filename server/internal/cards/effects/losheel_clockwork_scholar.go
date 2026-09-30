package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Losheel, Clockwork Scholar — Legendary Creature — Elephant
// Artificer {2}{W}, 2/4 (slice 296-m):
//
//	"Prevent all combat damage that would be dealt to attacking
//	 artifact creatures you control.
//	 Whenever one or more artifact creatures you control enter, draw a
//	 card. This ability triggers only once each turn."
//
// The prevention is a CR 614/615 replacement over combat damage
// events, narrower than Fog: only combat damage, only an attacking
// creature, only an artifact creature, only this controller's. It is
// a live rule for as long as Losheel is on the battlefield, not a
// one-shot like Fog's turn-scoped effect, so it reads the board on
// every damage event rather than registering a duration.
//
// The draw is Morbid Opportunist's exact shape: OncePerBatch collapses
// a simultaneous entry (an Amass or a token-doubled cast) into the one
// trigger the printed text describes, and b11TriggeredThisTurn gates
// a SECOND entry batch later in the turn out entirely.
//
// No simplification.
func init() {
	losheelDrawLabel := "Losheel, Clockwork Scholar — draw a card"
	Register(Spec{
		OracleID:     "f8c2a972-e38f-47e5-a355-6f30ad09b1ae",
		Name:         "Losheel, Clockwork Scholar",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			{
				Watches: []game.EventKind{game.EventDealDamage},
				AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
					if ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || !ev.IsCombatDamage {
						return false
					}
					target, ok := g.LookupCardForEffect(ev.DamageTarget)
					if !ok {
						return false
					}
					return target.Controller == src.Controller && target.IsArtifact() && target.IsCreature() &&
						target.AttackingTarget != uuid.Nil
				},
				Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
					ev.Cancel()
					return nil
				},
				Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
					return src.Controller
				},
				Label: "Losheel, Clockwork Scholar: prevent combat damage to attacking artifact creatures you control",
			},
		},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				entering, ok := enteredUnderYourControl(ev, source, g, false)
				if !ok || !entering.IsArtifact() || !entering.IsCreature() {
					return false
				}
				return !b11TriggeredThisTurn(g, source.InstanceID, losheelDrawLabel)
			}, losheelDrawLabel, Do(DrawCards{N: 1}))),
		},
	})
}
