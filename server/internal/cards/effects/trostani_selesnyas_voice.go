package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Trostani, Selesnya's Voice — Legendary Creature — Dryad {G}{G}{W}{W},
// 2/5:
//
//	"Whenever another creature you control enters, you gain life equal
//	 to that creature's toughness.
//	 {1}{G}{W}, {T}: Populate. (Create a token that's a copy of a
//	 creature token you control.)"
//
// The lifegain is Verdant Sun's Avatar's body without the Avatar's own
// entry ("another"): toughness is read as the trigger resolves, with
// the toughness it had on entering as the fallback if it has left
// (CR 608.2h). The populate is a CR 602 ability with a tap in its
// cost, so Trostani's summoning sickness applies (the engine enforces
// it). It can copy a token the trigger above just gained life for.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e94ef397-f5c5-4b8d-ae27-528352fa1d1e",
		Name:         "Trostani, Selesnya's Voice",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			Key:     "Trostani — gain life equal to its toughness",
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, true)
				return ok && c.IsCreature()
			},
			Build:  b20GainLifeEqualToToughnessBuild("Trostani — gain life equal to its toughness"),
			Effect: b20GainLifeEqualToToughnessEffect,
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}{G}{W}, {T}: Populate.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    Plus(ManaCost("{1}{G}{W}"), TapCost()),
			Effect:  Do(Populate{}),
		}},
	})
}
