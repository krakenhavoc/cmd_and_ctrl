package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Phyrexian Revoker — Artifact Creature — Phyrexian Horror {2}, 2/1:
//
//	"As this creature enters, choose a nonland card name.
//	 Activated abilities of sources with the chosen name can't be
//	 activated."
//
// Pithing Needle's card on a body (#4092, Captain N'ghathrod's list),
// with the two differences the printed words make:
//
//   - NO MANA EXEMPTION. The Needle says "unless they're mana
//     abilities"; the Revoker does not, so naming a Sol Ring stops it
//     tapping for mana (exemptMana is false). Both go through the same
//     activation gate; the clause decides.
//   - "A NONLAND CARD NAME" is a restriction on the CHOICE (CR 201.4a),
//     and the engine takes a name as free text with no vocabulary to
//     check it against (game/choose_card_name.go), the same position
//     Alpine Moon's "nonbasic land card name" is in. So the restriction
//     is kept where the name is read: a source only matches through a
//     face whose own type line is not a land, which is the only kind of
//     name a legal choice could have made (CR 201.4d judges a face by
//     its own characteristics). A name no legal choice could make
//     reaches nothing, so the card is never stronger than printed.
//
// The name is chosen as the creature enters (CR 614.12) through the
// shared card-name prompt, per instance, so two Revokers name two
// different cards; the window before the answer names nothing (the
// as-enters family's declared simplification, game/choose_card_name.go).
// Loyalty abilities are activated abilities (CR 606.1) and go through
// the same gate.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "786b3657-b595-4b38-83f2-d40915e51e2b",
		Name:         "Phyrexian Revoker",
		Completeness: CompletenessFull,
		AsEnters:     ChooseCardNameAsEnters("Phyrexian Revoker — choose a nonland card name"),
		ActivationRestrictions: []game.ActivationRestriction{
			revokerChosenNameCantActivate(
				"Phyrexian Revoker — activated abilities of sources with the chosen name can't be activated."),
		},
	})
}

// revokerChosenNameCantActivate is ChosenNameCantActivate for a
// NONLAND name with no mana exemption: the restricting permanent's
// chosen name, matched against the activating object's non-land faces.
func revokerChosenNameCantActivate(label string) game.ActivationRestriction {
	return game.ActivationRestriction{
		Label: label,
		Forbids: func(q game.ActivationQuery) bool {
			named := strings.TrimSpace(q.Source.ChosenName)
			if named == "" {
				return false
			}
			return nonlandFaceNamed(q.Card, named)
		},
	}
}

// nonlandFaceNamed reports whether `c` has a face called `named`
// (case-insensitive, CR 201.2b) whose own type line is not a land's.
// The active face is judged by the card's printed type line, every
// other face by its own.
func nonlandFaceNamed(c game.Card, named string) bool {
	if c.FaceDown {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(c.Name), named) && !c.PrintedIsLand() {
		return true
	}
	for _, f := range c.Faces {
		if strings.EqualFold(strings.TrimSpace(f.Name), named) && !faceTypeLineIsLand(f.TypeLine) {
			return true
		}
	}
	return false
}

// faceTypeLineIsLand is whether a printed type line carries the land type
// as a whole word ("Land", "Basic Land — Forest"; not "Landfall").
func faceTypeLineIsLand(typeLine string) bool {
	for _, w := range strings.Fields(strings.ToLower(typeLine)) {
		if w == "land" {
			return true
		}
	}
	return false
}
