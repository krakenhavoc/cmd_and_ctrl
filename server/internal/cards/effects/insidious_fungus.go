package effects

// Insidious Fungus — Creature — Fungus {G}, 1/2 (EDHREC rank 1900):
//
//	"{2}, Sacrifice this creature: Choose one —
//	 • Destroy target artifact.
//	 • Destroy target enchantment.
//	 • Draw a card. Then you may put a land card from your hand onto
//	   the battlefield tapped."
//
// The one-mana modal utility creature. A modal ACTIVATED ability has
// no shape in the catalog (Spec.Modes is cast-time only), so the
// three modes are three activated abilities with the same cost, one
// per bullet — the mode is chosen at activation either way (CR
// 700.2), and the client's ability menu is the mode picker. The two
// removal modes are targeted; the third is not.
//
// The third mode's "then you may put a land card from your hand onto
// the battlefield TAPPED" is the shared clause from #654, in its
// tapped form: the land arrives tapped out of the CR 614 entry
// pipeline rather than being tapped afterwards, so nothing that
// watches for a tap sees one. It is a put, not a play (CR 305.4), so
// the turn's land drop is untouched.
//
// Until #654 the clause was omitted and declared: the pick-from-hand
// prompt existed (#552) but the hand-to-battlefield move did not.
func init() {
	fungusCost := Plus(ManaCost("{2}"), SacrificeThis())
	Register(Spec{
		OracleID:     "0a8d0217-ff24-4177-b6be-707eb2b6b9e9",
		Name:         "Insidious Fungus",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{2}, Sacrifice Insidious Fungus: Destroy target artifact",
				Cost:    fungusCost,
				Targets: TargetPermanent("target artifact", Artifact()),
				Effect:  destroyFirstLegalTarget,
			},
			{
				Label:   "{2}, Sacrifice Insidious Fungus: Destroy target enchantment",
				Cost:    fungusCost,
				Targets: TargetPermanent("target enchantment", Enchantment()),
				Effect:  destroyFirstLegalTarget,
			},
			{
				Label: "{2}, Sacrifice Insidious Fungus: Draw a card, then you may put a land from your hand onto the battlefield tapped",
				Cost:  fungusCost,
				Effect: Do(
					DrawCards{N: 1},
					MayPutALandFromHandTapped("Insidious Fungus"),
				),
			},
		},
	})
}
