package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Manascape Refractor — Artifact {3}:
//
//	"This artifact enters tapped.
//	 This artifact has all activated abilities of all lands on the
//	 battlefield.
//	 You may spend mana as though it were mana of any color to pay the
//	 activation costs of this artifact's abilities."
//
// A layer-6 borrow over every land on the battlefield, whoever controls
// it (ADR 0155). Mana abilities are activated abilities (CR 605.1a), so
// the lent set is mostly mana rows, including a basic land's intrinsic
// CR 305.6 one. The artifact has no ability of its own, so the rider's
// "this artifact's abilities" is exactly the rows the borrow lent.
//
// No simplification.
func init() {
	const clause = "This artifact has all activated abilities of all lands on the battlefield."
	Register(Spec{
		OracleID:     "ae3517db-bb96-4c29-91dc-5bd68683dced",
		Name:         "Manascape Refractor",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		Static: []game.StaticAbility{HasActivatedAbilitiesOf(game.AbilityBorrow{
			Label:   clause,
			Lenders: BattlefieldLenders(ByController(Land())),
		})},
		AnyColorSpend: BorrowedAbilitiesSpendableAsAnyColor(
			"You may spend mana as though it were mana of any color to pay the activation costs of this artifact's abilities.", clause),
	})
}
