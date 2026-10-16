package effects

// Magewright's Stone — Artifact {2} (EDHREC rank 2890):
//
//	"{1}, {T}: Untap target creature that has an activated ability
//	 with {T} in its cost."
//
// The untapper for tap-ability creatures. The target clause is
// b27HasTapAbility, and it answers the printed question in the two
// places a creature's abilities can be written down:
//
//   - the CATALOG: an activated or mana ability the engine would offer
//     the creature's controller with a tap in the cost — a catalog
//     creature's declared abilities, a token's intrinsic ones, a land
//     creature's intrinsic mana, and anything a catalogued permanent
//     granted it;
//   - the PRINTED TEXT, for a creature the catalog has no entry for
//     (#3860). game.Card carries no oracle text, so the deck importer
//     reads it once (game.PrintsTapAbilityText: a "<cost>: <effect>"
//     line whose cost names {T}, outside quotation marks and reminder
//     text) and stamps Card.PrintsTapAbility; a multi-faced card is
//     judged on the face that is up. A copy takes it with the copied
//     values (CR 707.2).
//
// The clause is about what the creature HAS, not about whether the
// engine runs it, so an uncatalogued Llanowar-alike is a legal target
// even though activating its ability stays manual. Untapping it is the
// whole of this card's job and the engine does that.
//
// What is still not seen: an ability granted to a creature by a
// permanent the catalog does not run is not an ability the engine
// gave it, so it is not one the engine can see — but then the granting
// effect is not happening in this game either.
func init() {
	Register(Spec{
		OracleID:     "9e500094-c430-4c10-89cc-97f4f18b08bc",
		Name:         "Magewright's Stone",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Untap target creature that has an activated ability with {T} in its cost.",
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Targets: TargetCreature("target creature that has an activated ability with {T} in its cost", b27HasTapAbility()),
			Effect:  untapTheTarget,
		}},
	})
}
