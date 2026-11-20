package effects

// Reviving Melody — Sorcery {2}{G} (Journey into Nyx, #5558):
//
//	"Choose one or both —
//	 • Return target creature card from your graveyard to your hand.
//	 • Return target enchantment card from your graveyard to your hand."
//
// Season of Renewal's "choose one or both": each bullet returns its own
// graveyard target, in printed order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4834e6f2-1544-4e58-9092-ecc136cc0dee",
		Name:         "Reviving Melody",
		Completeness: CompletenessFull,
		Modes: ChooseOneOrMore(
			ModeDoing("Return target creature card from your graveyard to your hand.",
				TargetCardInGraveyard("target creature card in your graveyard", YouOwn(), Creature()),
				ReturnTheModesGraveyardTargetToHand),
			ModeDoing("Return target enchantment card from your graveyard to your hand.",
				TargetCardInGraveyard("target enchantment card in your graveyard", YouOwn(), Enchantment()),
				ReturnTheModesGraveyardTargetToHand),
		),
	})
}
