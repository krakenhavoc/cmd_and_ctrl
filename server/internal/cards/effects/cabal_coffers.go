package effects

// Cabal Coffers — Land:
//
//	"{2}, {T}: Add {B} for each Swamp you control."
//
// Rank 196, and the card that needs two of #352's four seams at once:
// a mana component in a mana ability's cost ({2}) and a scaled
// produced string (one {B} per Swamp). Neither existed before this
// pass, which is why a card this played was missing.
//
// The Swamp count reads EFFECTIVE subtypes, so Urborg, Tomb of
// Yawgmoth — which turns every land on the battlefield into a Swamp —
// really does turn this into "add {B} for each land you control".
// That pairing is most of why either card gets played, and it works
// here for free because the layer engine maintains effective type
// lines for battlefield permanents.
//
// The count happens AFTER the {2} is paid and the land is tapped, per
// CR 605.3a (a mana ability resolves immediately, all of it, with no
// window in between). With no Swamps out the ability is still
// activatable: it costs {2}, taps the land, and adds nothing. That is
// the printed card.
//
// Not auto-tappable, like every ability with a mana component — the
// player floats the {2} and clicks. See signets.go for the reasoning.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7358e164-5704-4e78-9b21-6a9bf2a968ce",
		Name:         "Cabal Coffers",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true, Mana: "{2}"},
			ProducedFunc: ProducedPerPermanent("B", MatchLandSubtype("Swamp")),
			Label:        "{2}, {T}: Add {B} for each Swamp you control",
		}},
	})
}
