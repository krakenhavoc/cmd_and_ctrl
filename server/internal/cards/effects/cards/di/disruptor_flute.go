package cardsdi

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Disruptor Flute — Artifact {2} (Modern Horizons 3, #5948):
//
//	"Flash
//	 As this artifact enters, choose a card name.
//	 Spells with the chosen name cost {3} more to cast.
//	 Activated abilities of sources with the chosen name can't be activated
//	 unless they're mana abilities."
//
// Curse of Silence's tax joined to Pithing Needle's ban, with no enchanted
// player: the name is chosen as the Flute enters (CR 614.12, the card-name
// prompt, ChooseCardNameAsEnters) and stored on the permanent, so until it is
// answered nothing is named. Every player's spell with the name costs {3} more
// (CR 601.2f; a cost increase on the cast, not on an ability or a special
// action), matched on any face (CR 201.2b). The ban is the shared
// ChosenNameCantActivate restriction with the mana-ability exemption. Flash is
// the printed keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d6c46906-1339-4440-8cd0-2bb9890d4975",
		Name:            "Disruptor Flute",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		AsEnters:        effects.ChooseCardNameAsEnters("Disruptor Flute — choose a card name"),
		CostModifiers: []game.CostModifier{
			effects.CostsMore(3, "Spells with the chosen name cost {3} more to cast.",
				func(q game.CostQuery) bool {
					named := strings.TrimSpace(q.Source.ChosenName)
					return named != "" && q.Ability == nil && q.SpecialAction == nil && game.CardNameMatches(q.Card, named)
				}),
		},
		ActivationRestrictions: []game.ActivationRestriction{
			effects.ChosenNameCantActivate(
				"Disruptor Flute — activated abilities of sources with the chosen name can't be activated unless they're mana abilities.",
				true),
		},
	})
}
