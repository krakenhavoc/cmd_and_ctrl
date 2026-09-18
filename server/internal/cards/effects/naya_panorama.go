package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Naya Panorama — Land (EDHREC rank 4461):
//
//	"{T}: Add {C}.
//	 {1}, {T}, Sacrifice this land: Search your library for a basic
//	 Mountain, Forest, or Plains card, put it onto the battlefield
//	 tapped, then shuffle."
//
// The Alara wedge fetch: an Evolving Wilds that costs a mana and
// narrows to three basic types, and in exchange taps for {C} on the
// turn it lands. A Naya deck plays it for the extra land TYPE — the
// fetch is a shuffle for a Courser of Kruphix or a Sensei's Divining
// Top, and the colourless tap means the land is never blank.
//
// The three-way shape is b43PanoramaFetch in batch43_helpers.go; the
// other four Panoramas join that table as their batches reach them
// rather than opening four more files for the same card.
//
// The sacrifice is part of the COST, so it happens on activation and
// the search happens on resolution — a Stifle on the ability leaves
// the land already dead, as printed. The pick rides the catalog-wide
// deterministic search chooser: the engine offers the legal basics
// and the activating player picks one, and "fail to find" is offered
// because a search that may find is a search that may decline
// (CR 701.19c).
//
// No simplification.
func init() {
	Register(b43PanoramaFetch(
		"71e28800-c42c-48c0-95e5-0296be54a4e8", "Naya Panorama",
		"Mountain", "Forest", "Plains"))
}

// b43PanoramaFetch is the Alara Panorama cycle: "{T}: Add {C}" plus
// "{1}, {T}, Sacrifice this land: Search your library for a basic
// <A>, <B>, or <C> card, put it onto the battlefield tapped, then
// shuffle."
//
// Lives here rather than in batch43_helpers.go because it is a CYCLE
// table, not a shared body — the next Panorama is a row, the way the
// gain lands and the Tron lands are rows.
func b43PanoramaFetch(oracleID, name string, subtypes ...string) Spec {
	match := IsBasicLandOfAnySubtype(subtypes...)
	reason := "Choose a basic " + subtypes[0]
	for i := 1; i < len(subtypes); i++ {
		if i == len(subtypes)-1 {
			reason += ", or " + subtypes[i]
		} else {
			reason += ", " + subtypes[i]
		}
	}
	reason += " to put onto the battlefield tapped"
	label := "{1}, {T}, Sacrifice this land: Search your library for a basic " +
		subtypes[0] + ", " + subtypes[1] + ", or " + subtypes[2] +
		" card, put it onto the battlefield tapped, then shuffle."
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label: label,
			Cost:  Plus(ManaCost("{1}"), TapCost(), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:        item.Controller,
					Predicate:     match,
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					Reveal:        true,
					Shuffle:       true,
					TappedOnEntry: true,
					Reason:        reason,
				}.Apply(NewContext(g, item))
			},
		}},
	}
}
