package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nalia de'Arnise — {1}{W}{B} Legendary Creature — Human Rogue 3/3
//
//	"You may look at the top card of your library any time.
//	 You may cast Cleric, Rogue, Warrior, and Wizard spells from the top of
//	 your library.
//	 At the beginning of combat on your turn, if you have a full party, put a
//	 +1/+1 counter on each creature you control and those creatures gain
//	 deathtouch until end of turn."
//
// The top card is public to its owner (LibraryTopVisible). The cast
// permission is a standing one over the top of the library (ADR 0066),
// narrowed to spells with one of the four party creature types; the filter
// names two types at a time, so two permissions of the same shape cover the
// four. Only spells are cast this way, never lands. The combat trigger is an
// intervening "if" on a full party (CR 700.8c, 603.4): four creatures, one
// per party type, read as combat begins and again as the trigger resolves.
func init() {
	buff := effects.InterveningIf(effects.FinB9FullParty,
		effects.On(game.EventStepBegan,
			func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.StepBegan(game.StepBeginCombat, true)(ev, source, game.Characteristic{}, g) &&
					effects.FinB9HasFullParty(g, source.Controller)
			},
			"Nalia de'Arnise — put a +1/+1 counter on each creature you control; they gain deathtouch until end of turn",
			func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := effects.B11PutCounterOnEachCreatureYouControl(g, item); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{
					Match:    effects.And(effects.Creature(), effects.YouControl()),
					Keywords: []string{"deathtouch"},
					Label:    "Nalia de'Arnise — deathtouch",
				}.Apply(ctx)
			}))
	permission := func(a, b string) game.CastPermission {
		return game.CastPermission{
			Zone:             game.ZoneLibrary,
			TopOfLibraryOnly: true,
			CastOnly:         true,
			Filter:           game.PermissionFilter{CreatureTypesAny: [2]string{a, b}},
			Label:            "You may cast " + a + " and " + b + " spells from the top of your library (Nalia de'Arnise)",
		}
	}
	effects.Register(effects.Spec{
		OracleID:          "1bf6248c-658c-4540-8bee-81af79d349be",
		Name:              "Nalia de'Arnise",
		Completeness:      effects.CompletenessFull,
		LibraryTopVisible: game.LibraryTopOwner,
		CastPermissions:   []game.CastPermission{permission("Cleric", "Rogue"), permission("Warrior", "Wizard")},
		Triggered:         []game.TriggeredAbility{buff},
	})
}
