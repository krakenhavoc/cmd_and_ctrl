package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Melek, Izzet Paragon — Legendary Creature — Weird Wizard {4}{U}{R}, 2/4
// (Commander Legends: Battle for Baldur's Gate):
//
//	"Play with the top card of your library revealed.
//	 You may cast instant and sorcery spells from the top of your library.
//	 Whenever you cast an instant or sorcery spell from your library, copy
//	 it. You may choose new targets for the copy."
//
// Magus of the Future's revealed top card, with the cast permission limited
// to instants and sorceries (lands cannot be played from the top). The copy
// trigger reads the zone the spell was cast from off the cast event (CR
// 601.2a) and is Codie's copy-the-cast-spell body, with new targets.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:          "095d9719-0db6-43de-8e4f-a0035a4c65ed",
		Name:              "Melek, Izzet Paragon",
		Completeness:      CompletenessFull,
		LibraryTopVisible: game.LibraryTopRevealed,
		CastPermissions: []game.CastPermission{
			PlayFromTopOfYourLibrary(game.PermissionFilter{InstantOrSorceryOnly: true},
				"Cast an instant or sorcery spell from the top of your library (Melek, Izzet Paragon)"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventCast || ev.Actor != source.Controller || ev.OldZone != game.ZoneLibrary {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && (c.IsInstant() || c.IsSorcery())
			}, "Melek, Izzet Paragon — copy that spell; you may choose new targets", gapB106CopyTheCastSpell),
		},
	})
}
