package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// printedFeatherToken is Everquill Phoenix's Feather (ADR 0223, #5910):
// a red artifact token named Feather with "{1}, Sacrifice this token:
// Return target Phoenix card from your graveyard to the battlefield
// tapped." It has no subtype and is not a creature. The target is
// chosen as the ability is activated (CR 602.2b) and checked again as
// it resolves (CR 608.2b). Its slug is an on-disk identity: never
// renamed, never reused.
func printedFeatherToken() TokenTemplate {
	return TokenTemplate{
		Slug: "feather",
		Card: game.Card{
			Name:     "Feather",
			TypeLine: "Token Artifact",
			Colors:   []string{"R"},
		},
		Text: "{1}, Sacrifice this token: Return target Phoenix card from your graveyard to the battlefield tapped.",
		Activated: []game.ActivatedAbilityShape{{
			Label: "{1}, Sacrifice this token: Return target Phoenix card from your graveyard to the battlefield tapped",
			Cost: game.AbilityCost{
				SacrificeSelf: true,
				Mana:          "{1}",
			},
			Targets: TargetCardInGraveyard("target Phoenix card from your graveyard", Subtype("Phoenix"), YouOwn()),
			Effect:  ReturnFirstLegalGraveyardTargetToBattlefieldTapped,
		}},
	}
}

// FeatherToken is the Feather's card value for CreateToken.
func FeatherToken() game.Card { return TokenFromCatalog(printedFeatherToken) }
