package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Forger's Foundry — Artifact {2}{U} (Outlaws of Thunder Junction
// Commander):
//
//	"{T}: Add {U}. When you spend this mana to cast an instant or sorcery
//	 spell with mana value 3 or less, you may exile that spell instead of
//	 putting it into its owner's graveyard as it resolves.
//	 {3}{U}{U}, {T}: You may cast any number of spells from among cards
//	 exiled with this artifact without paying their mana costs. Activate
//	 only as a sorcery."
//
// The mana ability carries a spend trigger (WhenManaSpent, Study Hall's
// seam) that fires only when the mana pays for an instant or sorcery with
// mana value 3 or less. As it resolves it asks the "you may" (CR 608.2d) and,
// on a yes, marks the spell on the stack (Game.MarkSpellExilesLinkedForEffect):
// the spell's own resolution exit (CR 608.2n) then exiles it, "exiled with"
// the Foundry (CR 607.2a), instead of putting it into the graveyard. A spell
// countered or fizzled first goes to the graveyard as usual, and a Foundry
// that has left the battlefield leaves no link to make. The second ability
// grants a free cast over each nonland card exiled with this Foundry and
// offers them together, any number of them, at sorcery speed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "118da256-d1ea-44e8-9026-317e49694d29",
		Name:           "Forger's Foundry",
		ReadsManaSpent: game.ReadsPoolManaOnly,
		Completeness:   CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
			SpendRiders: []game.ManaSpendRider{WhenManaSpent("Forger's Foundry", game.ManaSpendTrigger{
				Label:     "Forger's Foundry — you may exile that spell instead of putting it into its owner's graveyard as it resolves",
				Condition: cmdr2S41FoundryCheapInstantOrSorcery,
				Effect:    cmdr2S41FoundryExileTrigger,
			}, game.ManaRestrictCast)},
		}},
		Activated: []ActivatedAbility{{
			Label:        "{3}{U}{U}, {T}: You may cast any number of spells from among cards exiled with this artifact without paying their mana costs. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{3}{U}{U}"), TapCost()),
			SorcerySpeed: true,
			Purpose:      game.Purpose{Answers: game.AnswerValue},
			Effect:       cmdr2S41FoundryCastExiled,
		}},
	})
}
