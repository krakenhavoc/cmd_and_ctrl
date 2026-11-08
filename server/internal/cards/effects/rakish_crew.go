package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rakish Crew — Enchantment {2}{B} (Outlaws of Thunder Junction, #4468):
//
//	"When this enchantment enters, create a 1/1 red Mercenary creature
//	 token with "{T}: Target creature you control gets +1/+0 until end
//	 of turn. Activate only as a sorcery."
//	 Whenever an outlaw you control dies, each opponent loses 1 life and
//	 you gain 1 life. (Assassins, Mercenaries, Pirates, Rogues, and
//	 Warlocks are outlaws.)"
//
// The token is the set's catalog Mercenary (createAMercenary). An outlaw
// is a batch, not a creature type (b40Outlaw): the dead creature counts
// if it was an Assassin, Mercenary, Pirate, Rogue or Warlock as it last
// existed on the battlefield (leftAsSubtype, CR 603.10a), so a changeling
// counts and the Mercenary token the Crew makes counts. "You control" is
// its controller as it left (leftUnderControlOf). The drain is Ayara's
// (drainEachOpponent): life loss, and a gain of 1.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ddf34404-39dd-4dcf-a706-ec78d47167e9",
		Name:         "Rakish Crew",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Rakish Crew — create a 1/1 red Mercenary", createAMercenary()),
			On(game.EventLTB, rakishCrewOutlawYouControlDied,
				"Rakish Crew — an outlaw died: each opponent loses 1 life and you gain 1 life", drainEachOpponent),
		},
	})
}

// rakishCrewOutlawYouControlDied is "whenever an outlaw you control dies".
// The five types are words of the card's text (CR 612.2).
func rakishCrewOutlawYouControlDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := diedCreature(ev, g)
	if !ok || leftUnderControlOf(ev, dead) != source.Controller {
		return false
	}
	for _, t := range []string{"Assassin", "Mercenary", "Pirate", "Rogue", "Warlock"} {
		if leftAsSubtype(ev, dead, source.TextSubtype(t)) {
			return true
		}
	}
	return false
}
