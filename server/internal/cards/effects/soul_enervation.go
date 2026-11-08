package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soul Enervation — Enchantment {3}{B} (Murders at Karlov Manor, #4469):
//
//	"Flash
//	 When this enchantment enters, target creature gets -4/-4 until end
//	 of turn.
//	 Whenever one or more creature cards leave your graveyard, each
//	 opponent loses 1 life and you gain 1 life."
//
// The entry trigger shrinks the still-legal target (CR 608.2b). The
// second trigger is Chalk Outline's batch trigger: a mass return drains
// once, not once per card, and the creature cards may leave in any
// direction. The drain is the footprinted pair EachOpponentLosesLife
// and GainLife.
//
// No simplification.
func init() {
	const shrink = "Soul Enervation — target creature gets -4/-4 until end of turn"
	Register(Spec{
		OracleID:        "5b4508a3-1d21-46ec-8a50-cd3171a7348a",
		Name:            "Soul Enervation",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters(shrink, func(g *game.Game, item *game.StackItem) error {
				return vsPumpTheTarget(NewContext(g, item), -4, -4, shrink)
			}), TargetCreature("target creature")), ForTargets(HarmUntilEOT(0, -4, -4))),
			OncePerBatch(OnAny(leftGraveyardEvents,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return creatureCardLeftYourGraveyard(ev, source, g)
				},
				"Soul Enervation — each opponent loses 1 life and you gain 1 life",
				Do(EachOpponentLosesLife{N: 1}, GainLife{Amount: 1}))),
		},
	})
}
