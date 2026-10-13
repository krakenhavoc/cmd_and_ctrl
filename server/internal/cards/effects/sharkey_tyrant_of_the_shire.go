package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sharkey, Tyrant of the Shire — Legendary Creature — Avatar Rogue
// {2}{U}{B}, 2/4:
//
//	"Activated abilities of lands your opponents control can't be
//	 activated unless they're mana abilities.
//	 Sharkey has all activated abilities of lands your opponents control
//	 except mana abilities.
//	 Mana of any type can be spent to activate Sharkey's abilities."
//
// The gate is Pithing Needle's mana exemption on Linvala's "your
// opponents" shape: it looks at the SOURCE land, so the land is stopped
// and Sharkey's copy is not. The borrow leaves out mana rows (ADR 0155
// Except), so Sharkey gets a fetchland's or a utility land's non-mana
// abilities and no taps for mana. The rider covers only the rows this
// borrow lent. Sharkey has no ability of its own, so that is all of
// "Sharkey's abilities".
//
// DECLARED SIMPLIFICATION, weaker than printed: "any type" is the
// engine's any-colour spend, which widens a coloured symbol and leaves a
// colourless {C} alone (CR 106.1b). A lent ability whose cost prints {C}
// still needs colourless mana. Few lands print one.
func init() {
	const clause = "Sharkey has all activated abilities of lands your opponents control except mana abilities."
	Register(Spec{
		OracleID:     "93a476c0-c77e-43b0-8695-ada40368a002",
		Name:         "Sharkey, Tyrant of the Shire",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Mana of any type pays coloured costs in Sharkey's abilities, but a colourless {C} cost in one still needs colourless mana."},
		ActivationRestrictions: []game.ActivationRestriction{
			OpponentsSourcesCantActivate(
				"Sharkey, Tyrant of the Shire — activated abilities of lands your opponents control can't be activated unless they're mana abilities.",
				Land(), true),
		},
		Static: []game.StaticAbility{HasActivatedAbilitiesOf(game.AbilityBorrow{
			Label:   clause,
			Except:  game.BorrowExceptMana,
			Lenders: BattlefieldLenders(ByController(And(Land(), OpponentControls()))),
		})},
		AnyColorSpend: BorrowedAbilitiesSpendableAsAnyColor(
			"Mana of any type can be spent to activate Sharkey's abilities.", clause),
	})
}
