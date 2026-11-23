package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Captain Kirk, Boldly Going — Legendary Creature — Human Officer
// {1}{G}{W}{U}, 3/4 (Star Trek Commander, #4974):
//
//	"Vigilance
//	 Spacecraft creatures you control get +3/+3.
//	 Whenever Captain Kirk enters or attacks, search your library and/or
//	 graveyard for a basic land card or a card named Starship Enterprise,
//	 reveal it, and put it into your hand. If you search your library
//	 this way, shuffle."
//
// The anthem is Assault Drone's layer 7c static over creatures you
// control with the Spacecraft subtype, Kirk's own controller only, read
// off each permanent's effective types. The trigger is one ability with
// two conditions (WhenThisEntersOrAttacks). The search is Tower Winder's
// library-and/or-graveyard search (#3217) with no "may": the player is
// asked which zones to search before anything is shown, picks at most
// one basic land card or card named Starship Enterprise, reveals it and
// puts it into their hand, and the library is shuffled only if it was
// searched (CR 701.23).
//
// No simplification.
func init() {
	anthem := effects.B16Anthem(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.Controller == source.Controller && target.IsCreature() && target.HasSubtype("Spacecraft")
	}, 3, 3)
	anthem.Label = "Spacecraft creatures you control get +3/+3."
	trigger := effects.WhenThisEntersOrAttacks("Captain Kirk, Boldly Going — search your library and/or graveyard for a basic land card or a card named Starship Enterprise",
		func(g *game.Game, item *game.StackItem) error {
			return effects.SearchLibrary{
				Player: item.Controller,
				Predicate: func(c game.Card) bool {
					return effects.IsBasicLand(c) || c.Name == "Starship Enterprise"
				},
				Dest:          game.ZoneHand,
				Limit:         1,
				Reveal:        true,
				Shuffle:       true,
				AlsoGraveyard: true,
				Reason:        "Captain Kirk, Boldly Going — a basic land card or a card named Starship Enterprise to put into your hand",
			}.Apply(effects.NewContext(g, item))
		})
	trigger.Purpose = game.Purpose{Tutors: 1}
	effects.Register(effects.Spec{
		OracleID:        "5228fffe-6347-4177-bccb-b15621c7159c",
		Name:            "Captain Kirk, Boldly Going",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static:          []game.StaticAbility{anthem},
		Triggered:       []game.TriggeredAbility{trigger},
	})
}
