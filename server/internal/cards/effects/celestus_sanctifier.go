package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Celestus Sanctifier — {2}{W} Creature — Human Cleric 3/2 (#2586, ADR
// 0132):
//
//	"If it's neither day nor night, it becomes day as this creature
//	 enters.
//	 Whenever day becomes night or night becomes day, look at the top two
//	 cards of your library. Put one of them into your graveyard."
//
// The look is private to the controller (CR 701.20). "Put one of them"
// is mandatory: with two cards the controller picks which goes to the
// graveyard and the other stays where it was, on top; with one card in
// the library that card goes without a prompt, and an empty library does
// nothing. The graveyard move goes through the same route every other
// library-to-graveyard card uses, so replacements and commander offers
// apply.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6885aa03-fbfe-4c2e-a785-0865788574b4",
		Name:         "Celestus Sanctifier",
		Completeness: CompletenessFull,
		AsEnters:     BecomesDayAsEnters(),
		Triggered: []game.TriggeredAbility{
			WheneverDayBecomesNightOrNightBecomesDay("Celestus Sanctifier — look at the top two cards; put one into your graveyard",
				celestusSanctifierLook),
		},
	})
}

func celestusSanctifierLook(g *game.Game, item *game.StackItem) error {
	player := item.Controller
	looked := g.LookAtTopOfLibraryForEffect(player, 2)
	switch len(looked) {
	case 0:
		return nil
	case 1:
		return restIntoGraveyard(g, looked)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   item.SourceCardID,
		Question: "Celestus Sanctifier — put one of them into your graveyard",
		Cards:    looked,
		Min:      1,
		Max:      1,
		ThenKey:  celestusSanctifierPut,
		LiveZone: game.ZoneLibrary,
	})
	return nil
}

// celestusSanctifierPut is the keyed answer to "which one goes into your
// graveyard" (#4253). Its key is an on-disk identity: never renamed, never
// reused.
var celestusSanctifierPut = game.RegisterChooseCardsThen("choose-cards/celestus-sanctifier-put", func(g *game.Game, r game.CardsChosen) error {
	return restIntoGraveyard(g, r.Picked)
})
