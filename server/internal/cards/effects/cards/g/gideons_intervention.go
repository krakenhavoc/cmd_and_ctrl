package cardsg

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gideon's Intervention — {2}{W}{W} Enchantment (Amonkhet, S82 old batch 248):
//
//	"As this enchantment enters, choose a card name.
//	 Your opponents can't cast spells with the chosen name.
//	 Prevent all damage that would be dealt to you and permanents you control by
//	 sources with the chosen name."
//
// The name is chosen as it enters (CR 614.12), per instance, through the shared
// card-name prompt. The cast ban binds every opponent of the controller and
// reads the name on the spell's own characteristics; the prevention (CR 615) is
// OldB248PreventDamageBySourcesNamedChosen.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "02fe8ed4-3c5d-4875-b584-b33b819c47be",
		Name:         "Gideon's Intervention",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseCardNameAsEnters("Gideon's Intervention — choose a card name"),
		CastRestrictions: []game.CastRestriction{{
			Label: "Your opponents can't cast spells with the chosen name.",
			Forbids: func(q game.CastQuery) bool {
				named := strings.TrimSpace(q.Source.ChosenName)
				return named != "" && q.Controller != q.Source.Controller && game.CardNameMatches(q.Card, named)
			},
		}},
		Replacements: []game.ReplacementEffect{effects.OldB248PreventDamageBySourcesNamedChosen()},
	})
}
