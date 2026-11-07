package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bumi, Unleashed — Legendary Creature — Human Noble Ally {3}{R}{G}, 5/4
// (Avatar: The Last Airbender, #4459):
//
//	"Trample
//	 When Bumi enters, earthbend 4.
//	 Whenever Bumi deals combat damage to a player, untap all lands you
//	 control. After this phase, there is an additional combat phase.
//	 Only land creatures can attack during that combat phase."
//
// The enters trigger is the keyword's "target land you control"
// (Badgermole Cub's shape). The combat-damage trigger untaps every land
// its controller controls (Sword of Feast and Famine's body; an earthbent
// land creature is a land, so it untaps too), adds one combat phase
// after this one (CR 500.8), and binds the restriction to that combat
// alone (#5277, ADR 0188 Decision 13: AddPhases.DuringThatCombat). The
// restriction is registered as that combat begins and ends with it, so
// it touches neither the combat Bumi connected in nor any other added
// combat, even a newer one that runs first (the 2025-10-02 ruling). It
// is "can't attack" on every creature that isn't a land, read live and
// judged as attackers are declared (CR 508.1c, 611.2c), so a land that
// becomes a creature during that combat may attack and a creature that
// enters then may not. Two such triggers in one combat give two combats,
// each with its own restriction.
//
// No simplification.
func init() {
	const label = "Bumi, Unleashed — untap all lands you control; an additional combat phase where only land creatures can attack"
	extra := ExtraCombatAfterThisPhase()
	extra.DuringThatCombat = OnlyLandCreaturesCanAttack("Bumi, Unleashed — only land creatures can attack during this combat")
	Register(Spec{
		OracleID:        "1fef13d8-5da2-46a6-91ed-7a31bd46e538",
		Name:            "Bumi, Unleashed",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			avB14WhenEntersEarthbend("Bumi, Unleashed — earthbend 4", 4),
			WheneverThisDealsCombatDamageToAPlayer(label, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := untapAllLandsControlledBy(g, item.Controller, ctx); err != nil {
					return err
				}
				return extra.Apply(ctx)
			}),
		},
	})
}
