package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// energy_resolution_cards_test.go — ADR 0129 PR 3: the cards that pay
// energy as a spell or ability resolves (CR 118.12, 118.12a).

const (
	ErThrivingRhino  = "d8841f3a-f3ff-42ca-89a8-0cc1c3ca6a6c"
	ErAetherChaser   = "8aab0b55-2779-43c7-a3ee-151b8e5c72f3"
	ErRiparianTiger  = "dbb6b7f0-33fd-4530-a621-114148fb3929"
	ErLathnuHellion  = "6dd71453-63ec-4f4c-87d2-34d207398b9a"
	ErHarnessed      = "d8dd7e7f-0053-44bb-bac0-1a08843d4c49"
	ErGalvanic       = "ef0f06f1-3991-4897-bf62-d4cab7ec79a0"
	ErDieYoung       = "4bcc54d1-8ab6-4ad3-91b6-97fb02b88cc0"
	ErGreenbelt      = "c63a17b8-d183-48c7-bb04-edbfffba1e03"
	ErHexgoldSlith   = "0a467e75-f68d-4ad0-85ff-b8516de8d5bc"
	ErVoltaicBrawler = "ab7609dd-f4c0-4636-8177-7a926c01e470"
)

// energyPrompt is the open energy pay-unless owed by `p`, or nil.
func EnergyPrompt(g *game.Game, p uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == p {
			if a := c.PayAction(); a != nil && a.Kind == game.PayActionEnergy {
				return c
			}
		}
	}
	return nil
}

// payAmountPrompt is the open pay_amount owed by `p`, or nil.
func PayAmountPrompt(g *game.Game, p uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePayAmount && c.Chooser == p {
			return c
		}
	}
	return nil
}
