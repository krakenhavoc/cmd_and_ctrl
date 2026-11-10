package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gravestone Strider — Artifact Creature — Golem {2}, 1/3 (Murders at
// Karlov Manor, #4469):
//
//	"{1}: Add one mana of any color. Activate only once each turn.
//	 {2}, Exile this card from your graveyard: Exile target card from a
//	 graveyard."
//
// The filter is Barrels of Blasting Jelly's mana ability with the
// engine's OncePerTurn bit, so every path refuses a second activation
// in a turn. The second ability works only from the graveyard
// (CR 113.6, Bonebind Orator's ExileThis cost), and its target is any
// card in any graveyard, chosen before the cost is paid (CR 601.2c); a
// target gone by resolution is skipped (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9ea58a77-9b2e-4ae2-8a31-a51804900a3b",
		Name:         "Gravestone Strider",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:        ManaAbilityCost{Mana: "{1}"},
			Produced:    "{W|U|B|R|G}",
			Label:       "{1}: Add one mana of any color. Activate only once each turn.",
			OncePerTurn: true,
			Answers:     game.AnswerValue,
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}, Exile this card from your graveyard: Exile target card from a graveyard.",
			Cost:    Plus(ManaCost("{2}"), ExileThis()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Targets: TargetCardInGraveyard("target card from a graveyard"),
			Effect:  ExileTargetCards,
		}},
	})
}
