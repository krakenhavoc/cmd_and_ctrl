package effects

// Galvanizing Sawship — Artifact — Spacecraft, {5}{R}, 6/5:
//
//	Station (Tap another creature you control: Put charge counters
//	equal to its power on this Spacecraft. Station only as a sorcery.
//	It's an artifact creature at 3+.)
//	3+ | Flying, haste
//
// The smallest complete station card: the ability, one threshold, and
// the P/T box behind it. Three charge counters make it a 6/5 flier
// with haste — which is the interesting bit, because haste is what
// lets a Spacecraft that entered THIS turn swing the turn it is
// stationed. Summoning sickness asks how long the permanent has been
// under its controller's control (CR 302.6), and the layer-6 haste
// grant under the same gate answers it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dfe8f77a-cc26-438b-92ae-2ca7a91f813b",
		Name:         "Galvanizing Sawship",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Station()},
		Static: append(
			SpacecraftAt(3, 6, 5),
			ThresholdKeywords(3, "flying", "haste"),
		),
	})
}
