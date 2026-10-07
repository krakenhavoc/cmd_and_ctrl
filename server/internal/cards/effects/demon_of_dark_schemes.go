package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Demon of Dark Schemes — Creature — Demon {3}{B}{B}{B}, 5/5:
//
//	"Flying
//	 When this creature enters, all other creatures get -2/-2 until end
//	 of turn.
//	 Whenever another creature dies, you get {E} (an energy counter).
//	 {2}{B}, Pay {E}{E}{E}{E}: Put target creature card from a graveyard
//	 onto the battlefield tapped under your control."
//
// The -2/-2 locks its set as the trigger resolves (CR 611.2c) and spares
// the Demon itself, read as the object the trigger came from. "Another
// creature" is anyone's. ADR 0129 §2 (#1995): "Pay {E}{E}{E}{E}" is the
// energy cost component; the creature card comes from any graveyard and
// enters tapped under the activator's control.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "2b2f2abd-6c0f-49a5-a44b-f366614c944d",
		Name:            "Demon of Dark Schemes",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose: game.Purpose{Sweep: game.Sweep{
			Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 2}},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Demon of Dark Schemes — all other creatures get -2/-2 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return BoostUntilEOT{
						Match:     And(Creature(), OtherThan(ctx.Source())),
						Power:     -2,
						Toughness: -2,
						Label:     "Demon of Dark Schemes — -2/-2",
					}.Apply(ctx)
				}),
			TriggerWithPurpose(On(game.EventLTB, AnotherCreatureDied,
				"Demon of Dark Schemes — you get {E}", Do(GetEnergy{N: 1})), game.Purpose{Energy: 1}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}, Pay {E}{E}{E}{E}: Put target creature card from a graveyard onto the battlefield tapped under your control.",
			Cost:    Plus(ManaCost("{2}{B}"), PayEnergy(4)),
			Targets: targetCreatureInAnyGraveyard(),
			Effect:  returnFirstGraveyardTargetTapped(true),
		}},
	})
}
