package effects

// Mox Opal — Legendary Artifact {0}:
//
//	"Metalcraft — {T}: Add one mana of any color. Activate only if you
//	 control three or more artifacts."
//
// Rank 242. The gate again, on a free artifact, and the same #259
// hazard: without the condition this is a Mox that taps for any colour
// on turn one, which is one of the most powerful cards ever printed
// rather than the metalcraft-dependent one that actually exists.
//
// Mox Opal counts ITSELF toward metalcraft — it is an artifact — so
// the real threshold is two other artifacts. ControlsAtLeast(3,
// MatchArtifact) walks the whole battlefield including this one.
//
// "Metalcraft —" is an ability word: reminder text with no rules
// meaning of its own (CR 207.2c). The condition that follows is the
// whole ability, and it is what is modelled here.
//
// "One mana of any color" is a five-colour pipe, and the printed text
// does not mention the commander's identity — so
// IgnoreCommanderIdentity, exactly as City of Brass and Mana
// Confluence set it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "de2440de-e948-4811-903c-0bbe376ff64d",
		Name:     "Mox Opal",
		ManaAbilities: []ManaAbility{{
			Cost:                    ManaAbilityCost{Tap: true},
			Produced:                "{W|U|B|R|G}",
			Label:                   "Add one mana of any color (metalcraft)",
			Condition:               ControlsAtLeast(3, MatchArtifact),
			IgnoreCommanderIdentity: true,
		}},
	})
}
