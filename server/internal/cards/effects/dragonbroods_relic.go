package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dragonbroods' Relic — Artifact {1}{G} (Tarkir: Dragonstorm, #4463):
//
//	"{T}, Tap an untapped creature you control: Add one mana of any color.
//	 {3}{W}{U}{B}{R}{G}, Sacrifice this artifact: Create a 4/4 Dragon
//	 creature token named Reliquary Dragon that's all colors. It has
//	 flying, lifelink, and "When this token enters, it deals 3 damage to
//	 any target." Activate only as a sorcery."
//
// The mana ability is Scene of the Crime's: the artifact's {T} plus the
// tap of an untapped creature, which is not the {T} symbol, so a creature
// that came in this turn may pay it. The second ability makes the token
// from the catalog (ReliquaryDragonToken, ADR 0083): the token's abilities
// are its own printed ones, so the damage is dealt BY the token (CR 120.3)
// to a target chosen as the trigger goes on the stack, and the token is all
// five colors with flying and lifelink. The Relic is sacrificed as a cost,
// so the token exists whether or not the Relic is still around.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "61822bbe-0407-4425-b8d7-0263a8c521ec",
		Name:         "Dragonbroods' Relic",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost: ManaAbilityCost{
				Tap: true,
				TapOthers: &game.TapOthersCost{
					Count:  1,
					Filter: TargetPermanent("an untapped creature you control", Creature()),
					Label:  "an untapped creature you control",
				},
			},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []ActivatedAbility{{
			Label:        "{3}{W}{U}{B}{R}{G}, Sacrifice this artifact: Create a 4/4 Dragon creature token named Reliquary Dragon that's all colors. It has flying, lifelink, and \"When this token enters, it deals 3 damage to any target.\" Activate only as a sorcery.",
			Purpose:      game.Purpose{Tokens: 1},
			Cost:         Plus(ManaCost("{3}{W}{U}{B}{R}{G}"), SacrificeThis()),
			SorcerySpeed: true,
			Effect:       Do(CreateToken{Template: ReliquaryDragonToken(), N: 1}),
		}},
	})
}

// ReliquaryDragonToken is Dragonbroods' Relic's 4/4 Dragon named
// Reliquary Dragon: all five colors, flying, lifelink, and "When this
// token enters, it deals 3 damage to any target."
func ReliquaryDragonToken() game.Card { return tokenFromCatalog(printedReliquaryDragonToken) }

// printedReliquaryDragonToken is the Reliquary Dragon as PRINTED.
func printedReliquaryDragonToken() tokenTemplate {
	return tokenTemplate{
		Slug: "reliquary-dragon",
		Card: game.Card{
			Name:      "Reliquary Dragon",
			TypeLine:  "Token Creature — Dragon",
			Power:     4,
			Toughness: 4,
			Colors:    []string{"W", "U", "B", "R", "G"},
			Keywords:  []string{"flying", "lifelink"},
		},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Reliquary Dragon — 3 damage to any target", sourceDealsDamageToEachLegalTarget(3)),
				TargetAny()),
		},
		Text: "This token is all colors.\nFlying, lifelink\nWhen this token enters, it deals 3 damage to any target.",
	}
}
