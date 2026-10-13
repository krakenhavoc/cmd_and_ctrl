package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Drana and Linvala — Legendary Creature — Vampire Angel {1}{W}{W}{B},
// 3/4:
//
//	"Flying, vigilance
//	 Activated abilities of creatures your opponents control can't be
//	 activated.
//	 Drana and Linvala has all activated abilities of all creatures your
//	 opponents control. You may spend mana as though it were mana of any
//	 color to activate those abilities."
//
// Three printed halves, each a shape the engine has:
//
//   - The gate is Linvala, Keeper of Silence's: it looks at the SOURCE of
//     the ability, so the lender is stopped and Drana's own copy is not
//     (the copy's source is Drana, whom you control). That is the
//     printed ruling and the point of the card.
//   - The borrow is a layer-6 lend over the battlefield (ADR 0155).
//   - The rider is an AnyColorSpendStatic that covers only the rows this
//     borrow lent (ManaSpendContext.AbilityOrigin), so Drana's own
//     spells and any other payment stay as printed.
//
// No simplification.
func init() {
	const clause = "Drana and Linvala has all activated abilities of all creatures your opponents control."
	Register(Spec{
		OracleID:        "7abc2a1b-b758-480b-9415-f8f5369ab52f",
		Name:            "Drana and Linvala",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		ActivationRestrictions: []game.ActivationRestriction{
			OpponentsSourcesCantActivate(
				"Drana and Linvala — activated abilities of creatures your opponents control can't be activated.",
				Creature(), false),
		},
		Static: []game.StaticAbility{HasActivatedAbilitiesOf(game.AbilityBorrow{
			Label:   clause,
			Lenders: BattlefieldLenders(ByController(And(Creature(), OpponentControls()))),
		})},
		AnyColorSpend: BorrowedAbilitiesSpendableAsAnyColor(
			"You may spend mana as though it were mana of any color to activate those abilities.", clause),
	})
}
