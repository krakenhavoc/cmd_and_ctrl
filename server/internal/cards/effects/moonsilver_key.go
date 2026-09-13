package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Moonsilver Key — Artifact {2} (EDHREC rank 1666):
//
//	"{1}, {T}, Sacrifice this artifact: Search your library for an
//	 artifact card with a mana ability or a basic land card, reveal
//	 it, put it into your hand, then shuffle."
//
// Expedition Map for mana rocks: a two-mana artifact that becomes
// the Sol Ring or the basic you needed. Same three-component cost,
// same search to hand; the clause is
// b15ArtifactWithManaAbilityOrBasicLand. The searcher chooses (the
// S22 chooser), so with more than one match the prompt opens.
//
// Sandbox simplification: "an artifact card with a mana ability" is
// answered by the catalog — the same lookup the battlefield uses —
// so an artifact whose mana ability the game does not know about is
// not offered. Weaker than printed, never stronger; the basic-land
// half is complete.
func init() {
	Register(Spec{
		OracleID:     "373395b2-f04b-46c9-b9f4-4f6815b71009",
		Name:         "Moonsilver Key",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Only artifacts whose mana ability the game already knows about can be found; any basic land can."},
		Activated: []ActivatedAbility{{
			Label: "{1}, {T}, Sacrifice this artifact: Search your library for an artifact card with a mana ability or a basic land card, reveal it, put it into your hand, then shuffle.",
			Cost:  Plus(ManaCost("{1}"), TapCost(), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: b15ArtifactWithManaAbilityOrBasicLand,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Moonsilver Key — an artifact with a mana ability or a basic land, to your hand",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
