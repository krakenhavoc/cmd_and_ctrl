package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Koskun Falls — World Enchantment {2}{B}{B}:
//
//	"At the beginning of your upkeep, sacrifice this enchantment unless
//	 you tap an untapped creature you control.
//	 Creatures can't attack you unless their controller pays {2} for each
//	 creature they control that's attacking you."
//
// #1997, one of the world enchantments ADR 0109 PR 7 left out. The
// upkeep is the ordinary pay-unless prompt (CR 118.12a) with a TAP
// payment (game.PayActionTap): the controller names one untapped
// creature they control and it is tapped, or they decline and Koskun
// Falls is sacrificed. With no untapped creature there is no payment to
// make (CR 118.3), so it is sacrificed. Tapping the creature is not one
// of its own {T} abilities, so a creature that arrived this turn can pay.
// The attack tax is Ghostly Prison's.
//
// A world permanent: the world rule (CR 704.5k, ADR 0109 §8) puts it
// into its owner's graveyard when a newer one enters.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "a9b9e798-debc-4f2c-9770-e16f670082f7",
		Name:         "Koskun Falls",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Koskun Falls — sacrifice it unless you tap an untapped creature you control", koskunFallsUpkeep),
		},
		AttackTaxes: []game.AttackTax{
			AttackTax("{2}", "Creatures can't attack you unless their controller pays {2} for each creature they control that's attacking you."),
		},
	})
}

func koskunFallsUpkeep(g *game.Game, item *game.StackItem) error {
	sourceID := item.SourceCardID
	// A Koskun Falls that left in response (or left and came back, CR
	// 400.7) has nothing to sacrifice and asks nothing.
	if !onBattlefield(g, sourceID) || sourceIsNewObject(g, item) {
		return nil
	}
	pay := TapPayment(1, "creature", "creatures", QueryTypes("creature"))
	words, action := pay.prompt()
	return UpkeepPayUnless{
		Chooser:  item.Controller,
		Cost:     words,
		Action:   action,
		Question: "Koskun Falls — tap an untapped creature you control or sacrifice Koskun Falls?",
		Else:     upkeepSacrificeSource,
	}.Apply(NewContext(g, item))
}
