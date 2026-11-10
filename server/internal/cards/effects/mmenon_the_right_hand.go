package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const mmenonArtifactMana = "mmenon-the-right-hand/artifact-mana"

// Mm'menon, the Right Hand — Legendary Creature — Jellyfish Advisor
// {3}{U}{U}, 3/4 (Edge of Eternities):
//
//	"Flying
//	 You may look at the top card of your library any time.
//	 You may cast artifact spells from the top of your library.
//	 Artifacts you control have "{T}: Add {U}. Spend this mana only to
//	 cast a spell from anywhere other than your hand.""
//
// Augur of Autumn's owner-visible top card (CR 401.5) with a
// cast-only artifact permission over it, and Galazeth Prismari's grant
// of a mana ability to every artifact you control. The mana is
// restricted twice, as printed: it can pay only for casting a spell
// (ManaRestrictCast) and only for one cast from a zone other than the
// hand (ManaRestrictNotFromHand, the spend context's cast zone), so it
// can't pay an ability's cost and can't pay for a spell from hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:          "de58fa4b-eeeb-4dc8-aab9-6521b15678cc",
		Name:              "Mm'menon, the Right Hand",
		Completeness:      CompletenessFull,
		PrintedKeywords:   []string{"flying"},
		LibraryTopVisible: game.LibraryTopOwner,
		CastPermissions:   []game.CastPermission{mmenonCastFromTop()},
		Grants: []AbilityGrant{{
			Key: mmenonArtifactMana,
			Mana: []ManaAbility{{
				Cost:         ManaAbilityCost{Tap: true},
				Produced:     "{U}",
				Label:        "Add {U}. Spend this mana only to cast a spell from anywhere other than your hand",
				Restrictions: []string{ManaRestrictCast, ManaRestrictNotFromHand},
			}},
			Text: "{T}: Add {U}. Spend this mana only to cast a spell from anywhere other than your hand.",
		}},
		Static: []game.StaticAbility{Labelled(GrantAbilities(func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.IsArtifact() && target.Controller == source.Controller
		}, mmenonArtifactMana), "Artifacts you control have \"{T}: Add {U}. Spend this mana only to cast a spell from anywhere other than your hand.\"")},
	})
}

// mmenonCastFromTop is the cast-only (never "play", so no artifact land)
// artifact permission over the top card of the library.
func mmenonCastFromTop() game.CastPermission {
	p := PlayFromTopOfYourLibrary(
		game.PermissionFilter{CardType: "artifact"},
		"Cast an artifact spell from the top of your library (Mm'menon, the Right Hand)")
	p.CastOnly = true
	return p
}
