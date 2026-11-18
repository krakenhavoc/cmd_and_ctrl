package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_snc_b18_helpers.go — shared pieces for Streets of New Capenna batch 18
// (#5491).

// MagesAttendantWizardToken is Mage's Attendant's 1/1 blue Wizard with "{1},
// Sacrifice this token: Counter target noncreature spell unless its controller
// pays {1}." The ability is printed on the token, so it is a token template
// (ADR 0083), not a tokens_table.go row.
func MagesAttendantWizardToken() game.Card { return tokenFromCatalog(printedS82SncB18WizardToken) }

// printedS82SncB18WizardToken is that Wizard as PRINTED.
func printedS82SncB18WizardToken() tokenTemplate {
	const text = "{1}, Sacrifice this token: Counter target noncreature spell unless its controller pays {1}."
	return tokenTemplate{
		Slug: "wizard-sacrifice-counter",
		Card: game.Card{
			Name:      "Wizard",
			TypeLine:  "Token Creature — Wizard",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"U"},
		},
		Text: text,
		Activated: []game.ActivatedAbilityShape{{
			Label:   text,
			Cost:    Plus(ManaCost("{1}"), SacrificeThis()),
			Targets: TargetSpell("target noncreature spell", Noncreature()),
			Effect:  counterTheTargetSpellUnlessPaid("Wizard", "{1}"),
		}},
	}
}
