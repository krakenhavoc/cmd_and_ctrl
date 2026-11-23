package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Roiling Canopy — Land:
//
//	"This land enters tapped.
//	 Whenever a Forest you control enters, if you control at least five
//	 other Forests, target creature you control gets +3/+3 until end of
//	 turn.
//	 {T}: Add {G}."
//
// A targeted landfall-style trigger with an intervening "if" (CR 603.4):
// the five-other-Forests count is checked as the Forest enters, so the
// trigger is not even put on the stack otherwise. "Other" is other than
// the Forest that entered; the Canopy has no land type of its own, so
// it never counts. Forests are counted by effective land type, so a
// land that has become a Forest counts.
//
// The second check, as the ability resolves, reads the Forest that
// entered off the trigger's event and counts the Forests OTHER than it,
// so a Forest that left in response changes nothing while five others
// stay.
func init() {
	pump := effects.On(game.EventETB, effects.RfMiscAForestEnteredWithFiveOthers,
		"Roiling Canopy — target creature you control gets +3/+3 until end of turn",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			if item.Trigger == nil || effects.RfMiscAForestsYouControlOtherThan(g, item.Controller, item.Trigger.Event.CardID, ctx.TextSubtype("Forest")) < 5 {
				return nil // CR 603.4 re-check; see the comment above
			}
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return effects.BoostUntilEOT{
					Target: t.ID, Power: 3, Toughness: 3,
					Label: "Roiling Canopy — +3/+3",
				}.Apply(ctx)
			}
			return nil
		})
	pump.Targets = effects.TargetCreature("target creature you control", effects.YouControl())
	effects.Register(effects.Spec{
		OracleID:     "c6e87760-4cd0-4281-82d4-7a377ea960ad",
		Name:         "Roiling Canopy",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Triggered: []game.TriggeredAbility{pump},
	})
}
