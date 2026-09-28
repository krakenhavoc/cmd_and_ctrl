package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Queen Marchesa — Legendary Creature — Human Assassin {1}{R}{W}{B}, 3/3:
//
//	"Deathtouch, haste
//	 When Queen Marchesa enters, you become the monarch.
//	 At the beginning of your upkeep, if an opponent is the monarch,
//	 create a 1/1 black Assassin creature token with deathtouch and
//	 haste."
//
// The monarch commander (#1722). The upkeep ability is the one monarch
// line in this batch with an INTERVENING "if" (CR 603.4): it is checked
// as the upkeep begins — no trigger at all while you hold the crown or
// nobody does — and again as the trigger resolves, so an Assassin is
// not made if you took the crown back in response. The Assassin is a
// hasty deathtouch attacker whose job is exactly that.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d7ac4be1-dcca-49b6-8ddb-d1b0e6cf2dcf",
		Name:            "Queen Marchesa",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch", "haste"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouBecomeTheMonarch("Queen Marchesa"),
			On(game.EventBeginUpkeep, AllOf(ByYou, anOpponentIsTheMonarchNow),
				"Queen Marchesa — create a 1/1 Assassin with deathtouch and haste", queenMarchesaUpkeep),
		},
	})
}

// anOpponentIsTheMonarchNow is the intervening "if" as a trigger
// condition.
func anOpponentIsTheMonarchNow(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return AnOpponentIsTheMonarch(g, source.Controller)
}

// queenMarchesaUpkeep re-checks the intervening "if" and makes the
// Assassin.
func queenMarchesaUpkeep(g *game.Game, item *game.StackItem) error {
	if !AnOpponentIsTheMonarch(g, item.Controller) {
		return nil
	}
	return CreateToken{Template: TokenCard("1/1 black Assassin with deathtouch and haste"), N: 1}.Apply(NewContext(g, item))
}
