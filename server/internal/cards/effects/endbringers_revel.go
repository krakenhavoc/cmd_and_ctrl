package effects

// Endbringer's Revel — Enchantment {2}{B}:
//
//	"{4}: Return target creature card from a graveyard to its owner's
//	 hand. Any player may activate this ability but only as a sorcery."
//
// An any-player row (CR 602.2, 602.1b) with a timing instruction: "only
// as a sorcery" (CR 602.5d) is the ACTIVATOR's sorcery timing, so a
// non-controller may use it in their own main phase with an empty stack.
// Whoever activates it pays the {4} out of their own pool (CR 602.1a)
// and chooses the target, a creature card in ANY graveyard. The card
// goes to its owner's hand, not the activator's (Memorial to Folly's
// body, b27ReturnChosenGraveyardCardToHand). A target that has left
// the graveyard by resolution is illegal and nothing moves (CR 608.2b).
//
// No Purpose: ActivationPurpose has no "return a card" field, so the bot
// does not reach across the table for it (ADR 0106 owner decision 2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4ca2511d-76f0-470c-b522-b75fb08932d9",
		Name:         "Endbringer's Revel",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{4}: Return target creature card from a graveyard to its owner's hand. Any player may activate this ability but only as a sorcery.",
			Cost:         ManaCost("{4}"),
			Targets:      TargetCardInGraveyard("target creature card from a graveyard", Creature()),
			SorcerySpeed: true,
			AnyPlayer:    true,
			Effect:       b27ReturnChosenGraveyardCardToHand,
		}},
	})
}
