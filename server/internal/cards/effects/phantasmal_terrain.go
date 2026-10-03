package effects

// Phantasmal Terrain — Enchantment — Aura {U}{U}:
//
//	"Enchant land
//	 As this Aura enters, choose a basic land type.
//	 Enchanted land is the chosen type."
//
// ADR 0109 owner decision 4 (#1881). The choice is CR 614.12's, made as
// the Aura enters and stored on it (Card.ChosenOption), and each of the
// five types is its own gated static (ChosenIs), so until the choice is
// made the land is unchanged. Then it is effects.SetsBasicLandType, the
// static form of CR 305.7, attached to the enchanted land: in layer 4 its
// land types are replaced by the chosen one and every other subtype stays
// (CR 205.1a), it loses the abilities its rules text gives it, keeps any
// another effect granted it, and taps for the chosen colour (CR 305.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7dcbce46-2973-4a9f-93df-95ac41ce668a",
		Name:         "Phantasmal Terrain",
		Completeness: CompletenessFull,
		Targets:      EnchantLand(),
		AsEnters:     ChooseBasicLandTypeAsEnters("Phantasmal Terrain"),
		Static:       EnchantedLandIsTheChosenType(),
	})
}
