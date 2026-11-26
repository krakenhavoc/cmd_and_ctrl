package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rage Extractor — Artifact {4}{R/P} (New Phyrexia, #5575):
//
//	"({R/P} can be paid with either {R} or 2 life.)
//	 Whenever you cast a spell with {H} in its mana cost, this artifact deals
//	 damage equal to that spell's mana value to any target."
//
// {H} is the Phyrexian mana symbol (CR 107.4f): a spell whose mana cost holds
// a {W/P}-style symbol, however it was paid. The trigger is one per such spell
// the controller casts, aimed as it goes on the stack, and the damage is the
// spell's mana value read while it is still on the stack (CR 202.3e, so {X} is
// its chosen value) when the trigger resolves, or off the card if it has since
// left. A copy of a spell was not cast and does not trigger it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e45f0a12-e1ca-432f-a208-9359dd18513d",
		Name:         "Rage Extractor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventCast, rageExtractorCast,
				"Rage Extractor — it deals damage equal to that spell's mana value to any target",
				effects.SourceDealsComputedDamageToEachLegalTarget(func(ctx *effects.Context) int {
					spell, ok := ctx.Game.LookupCardForEffect(ctx.Trigger().Event.CardID)
					if !ok {
						return 0
					}
					mv, _ := ctx.Game.ManaValueForEffect(spell)
					return mv
				})),
				effects.TargetAny()),
		},
	})
}

// rageExtractorCast is "whenever you cast a spell with {H} in its mana cost":
// a cast by the source's controller of a spell whose cost has a Phyrexian
// mana symbol.
func rageExtractorCast(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	if !ok {
		return false
	}
	cost, err := game.ParseCost(spell.ManaCost)
	return err == nil && cost.HasPhyrexian
}
