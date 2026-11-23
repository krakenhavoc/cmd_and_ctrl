package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Resonating Lute — Artifact {2}{U}{R}:
//
//	"Lands you control have '{T}: Add two mana of any one color.
//	 Spend this mana only to cast instant and sorcery spells.'
//	 {T}: Draw a card. Activate only if you have seven or more cards
//	 in your hand."
//
// The restricted-mana half is TWO granted mana abilities, not one.
// The engine's mana-restriction tags are ANDed within one ability
// (Cavern of Souls, Sliver Hive), and there is no "type:Instant OR
// type:Sorcery" tag — adding one would be an engine change outside
// this slice. Two abilities, each Tap-costed and each restricted to
// one of the two card types, are the same set of legal spends as the
// printed OR: a land can only tap once, so tapping it for the
// "instants only" ability or the "sorceries only" ability is exactly
// the choice the printed clause's single restricted token would leave
// to the SPEND side. Not a simplification — the observable behaviour
// matches the card.
//
// The draw ability is an ordinary Activated entry with a hand-size
// activation condition (HandSizeAtLeast, added alongside
// GraveyardAtLeast's threshold pattern).
//
// No simplification.
func init() {
	const grantInstants = "resonating-lute/instants"
	const grantSorceries = "resonating-lute/sorceries"
	effects.Register(effects.Spec{
		OracleID:     "9263287d-f782-4318-9ab8-e2e5f8107f5e",
		Name:         "Resonating Lute",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{
			{
				Key:  grantInstants,
				Text: "{T}: Add two mana of any one color. Spend this mana only to cast instant and sorcery spells.",
				Mana: []effects.ManaAbility{{
					Cost:     effects.ManaAbilityCost{Tap: true},
					Produced: effects.OneColorOfAmount(2),
					Label:    "Add two mana of any one color (instant spells only)",
					Restrictions: []string{
						effects.ManaRestrictCast,
						effects.ManaRestrictType("Instant"),
					},
				}},
			},
			{
				Key:  grantSorceries,
				Text: "{T}: Add two mana of any one color. Spend this mana only to cast instant and sorcery spells.",
				Mana: []effects.ManaAbility{{
					Cost:     effects.ManaAbilityCost{Tap: true},
					Produced: effects.OneColorOfAmount(2),
					Label:    "Add two mana of any one color (sorcery spells only)",
					Restrictions: []string{
						effects.ManaRestrictCast,
						effects.ManaRestrictType("Sorcery"),
					},
				}},
			},
		},
		Static: []game.StaticAbility{
			effects.GrantAbilities(effects.LandsYouControl, grantInstants, grantSorceries),
		},
		Activated: []effects.ActivatedAbility{{
			Label:     "{T}: Draw a card. Activate only if you have seven or more cards in your hand.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.TapCost(),
			Condition: effects.HandSizeAtLeast(7),
			Effect:    effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
