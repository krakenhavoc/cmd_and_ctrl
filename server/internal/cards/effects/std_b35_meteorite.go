package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// std_b35_meteorite.go — Outlaws of Thunder Junction's Meteorite token
// (#4468), made by Roxanne, Starfall Savant: "a tapped colorless artifact
// token named Meteorite with 'When this token enters, it deals 2 damage to
// any target' and '{T}: Add one mana of any color.'"
//
// Both abilities are printed on the token, so it is a token template (ADR
// 0083), registered in token_catalog.go's list. The enters trigger's
// damage is dealt by the token as it last existed there.

// MeteoriteToken is the Meteorite template.
func MeteoriteToken() game.Card { return tokenFromCatalog(printedMeteoriteToken) }

func printedMeteoriteToken() tokenTemplate {
	return tokenTemplate{
		Slug: "meteorite",
		Card: game.Card{
			Name:     "Meteorite",
			TypeLine: "Token Artifact",
		},
		Text: "When this token enters, it deals 2 damage to any target.\n{T}: Add one mana of any color.",
		Mana: []game.ManaAbilityShape{{
			TapCost:  true,
			Produced: "{W|U|B|R|G}",
			Label:    "{T}: Add one mana of any color",
		}},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters("Meteorite — 2 damage to any target",
				sourceDealsDamageToEachLegalTarget(2)), TargetAny()), ForTargets(DamageToTarget(0, 2))),
		},
	}
}
