package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mikey & Don, Party Planners — Legendary Creature — Mutant Ninja Turtle
// {2}{G/U}{G/U}, 3/3 (TMT, #4457):
//
//	"Ward {2}
//	 You may look at the top card of your library any time.
//	 You may play lands and cast Mutant, Ninja, or Turtle spells from the
//	 top of your library. If you cast a creature spell this way, that
//	 creature enters with an additional +1/+1 counter on it."
//
// The look is the private strength (LibraryTopOwner), and the play is
// standing permissions over the top of the library: Courser of Kruphix's
// lands, and Conspicuous Snoop's typed spells. "Mutant, Ninja, or Turtle"
// is a union of three creature types, judged on the spell as cast, so a
// changeling qualifies. PermissionFilter.CreatureTypesAny holds two, so
// the union is two permissions side by side, {Mutant, Ninja} and
// Turtle, with the same rider; a spell that matches both is cast through
// whichever is found first, which is the same cast.
//
// "That creature enters with an additional +1/+1 counter" is the
// permissions' own rider (CastPermission.EntersWithCounters, #4679): it
// marks the creature spell cast through one, and the counter is put on
// as the creature enters, through the CR 614 pipeline (CR 614.1c,
// CR 614.12), so Doubling Season doubles it. The same creature cast from
// hand enters with none, and a countered spell takes the mark with it.
//
// No simplification.
func init() {
	const name = "Mikey & Don, Party Planners"
	typed := func(filter game.PermissionFilter, label string) game.CastPermission {
		p := PlayFromTopOfYourLibrary(filter, label)
		p.EntersWithCounterKind = game.CounterPlusOne
		p.EntersWithCounters = 1
		return p
	}
	Register(Spec{
		OracleID:          "529178e8-425e-4d59-948c-ca7d0663f6ff",
		Name:              name,
		Completeness:      CompletenessFull,
		LibraryTopVisible: game.LibraryTopOwner,
		CastPermissions: []game.CastPermission{
			PlayFromTopOfYourLibrary(
				game.PermissionFilter{LandsOnly: true},
				"Play a land from the top of your library ("+name+")"),
			typed(game.PermissionFilter{NonLandOnly: true, CreatureTypesAny: [2]string{"Mutant", "Ninja"}},
				"Cast a Mutant or Ninja spell from the top of your library; a creature cast this way enters with an additional +1/+1 counter ("+name+")"),
			typed(game.PermissionFilter{NonLandOnly: true, CreatureType: "Turtle"},
				"Cast a Turtle spell from the top of your library; a creature cast this way enters with an additional +1/+1 counter ("+name+")"),
		},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), name+" — ward {2}"),
		},
	})
}
