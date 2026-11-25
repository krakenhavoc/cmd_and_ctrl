package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Flummoxed Cyclops — Creature — Cyclops {3}{R}, 4/4 (Theros Beyond Death, #5514):
//
//	"Reach
//	 Whenever two or more creatures your opponents control attack, this creature can't
//	 block this combat."
//
// One trigger per declaration (CR 603.2c: OncePerBatchPerPlayer collapses the
// per-attacker EventAttack) that fires when the attacking player is an opponent and
// controls two or more attacking creatures. It restricts the Cyclops until the end
// of the combat it was made in (CR 511.3, DurationUntilEndOfCombat), not the turn:
// an additional combat is a different combat.
//
// No simplification.
func init() {
	const label = "Flummoxed Cyclops — this creature can't block this combat"
	effects.Register(effects.Spec{
		OracleID:        "789a94be-020e-4f65-8f0d-0a170d4cd2f5",
		Name:            "Flummoxed Cyclops",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatchPerPlayer(effects.On(game.EventAttack, twoOrMoreOpponentCreaturesAttack, label,
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					d, ok := effects.DurationUntilEndOfCombat(ctx)
					if !ok || !effects.OnBattlefield(g, item.SourceCardID) {
						return nil
					}
					return effects.ScopedEffectFor{
						Target:   item.SourceCardID,
						Mods:     []game.Mod{game.AddRestrictionsMod(game.CantBlock)},
						Duration: d,
						Label:    label,
					}.Apply(ctx)
				})),
		},
	})
}

// twoOrMoreOpponentCreaturesAttack: the attacking player is an opponent of the
// source's controller and has two or more attacking creatures.
func twoOrMoreOpponentCreaturesAttack(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventAttack || ev.Actor == uuid.Nil || ev.Actor == source.Controller {
		return false
	}
	return len(effects.B13AttackingCreaturesYouControl(g, ev.Actor)) >= 2
}
