package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Queen's Diamond — Legendary Artifact — Spacecraft {4}, 5/5 (Star Trek
// Commander, #4974):
//
//	"{T}: Choose a color. Add one mana of that color for each charge
//	 counter on Queen's Diamond. Spend this mana only to cast your
//	 commander.
//	 Station (Tap another creature you control: Put charge counters
//	 equal to its power on this Spacecraft. Station only as a sorcery.
//	 It's an artifact creature at 8+.)
//	 8+ | Flying, vigilance, haste"
//
// Astral Cornucopia's mana ability (ProducedOneColor over the charge
// counters on the source, read as the ability resolves), with the CR 106.6
// spend restriction game.ManaRestrictYourCommander: the mana pays for a
// spell that is a commander its caster owns (CR 903.3d), cast from the
// command zone or from anywhere else, and for nothing else. With no
// charge counters it adds nothing. Station is the ordinary keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "44837c9b-971a-44b6-8e55-301cebc3a361",
		Name:         "Queen's Diamond",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:           ManaAbilityCost{Tap: true},
			ProducedAmount: ProducedOneColor(chargeCountersOn),
			Restrictions:   []string{ManaRestrictCast, game.ManaRestrictYourCommander},
			Label:          "Choose a color. Add one mana of that color for each charge counter on Queen's Diamond. Spend this mana only to cast your commander",
		}},
		Activated: []ActivatedAbility{Station()},
		Static: append(
			SpacecraftAt(8, 5, 5),
			ThresholdKeywords(8, "flying", "vigilance", "haste"),
		),
	})
}
