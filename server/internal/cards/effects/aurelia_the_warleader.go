package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aurelia, the Warleader — Legendary Creature — Angel
// {2}{R}{R}{W}{W}, 3/4:
//
//	"Flying, vigilance, haste
//	 Whenever Aurelia attacks for the first time each turn, untap all
//	 creatures you control. After this phase, there is an additional
//	 combat phase."
//
// "For the first time each turn" reads the per-object attack history
// (ADR 0059 Decision 8): the tally records the attack before the
// trigger harvester sees it, so the first attack reads 1 and the one in
// the added combat reads 2. An Aurelia that left the battlefield and
// came back is a new object and may trigger again (CR 400.7). No main
// phase comes with the combat (the 2024-11-08 ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0f5a3a09-2f07-4774-9e0f-e99d9a444166",
		Name:            "Aurelia, the Warleader",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance", "haste"},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, attacksForTheFirstTimeThisTurn,
				"Aurelia, the Warleader — untap all creatures you control, additional combat",
				untapAllYouControlThenExtraCombat),
		},
	})
}

// attacksForTheFirstTimeThisTurn is "whenever ~ attacks for the first
// time each turn" (Aurelia, Godo, Scourge of the Throne).
func attacksForTheFirstTimeThisTurn(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	return ThisAttacked(ev, source, lki, g) && g.TimesAttackedThisTurn(source.InstanceID) == 1
}
