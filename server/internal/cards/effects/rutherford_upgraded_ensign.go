package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rutherford, Upgraded Ensign — Legendary Creature — Human Artificer
// {2}{W}, 2/2 (Star Trek, #4973):
//
//	"When Rutherford enters, look at the top six cards of your library.
//	 You may reveal an artifact card from among them and put it into your
//	 hand. Put the rest on the bottom of your library in a random order.
//	 Whenever you cast a Spacecraft spell, put a +1/+1 counter on target
//	 creature."
//
// The entry is Horn of the Mark's look-and-take over six cards; the cast
// trigger is a targeted YouCast over the Spacecraft subtype.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "42258216-fe2b-42d0-a6ce-76ecf8104c10",
		Name:         "Rutherford, Upgraded Ensign",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Rutherford — look at the top six cards of your library",
				LookAtTopThenMayTakeToHand(6, Artifact(), 1,
					"Rutherford — you may reveal an artifact card from among them and put it into your hand")),
			Targeting(
				WheneverYouCast(HasSubtype("Spacecraft"), "Rutherford — put a +1/+1 counter on target creature",
					putPlusOneCounterOnEachLegalTarget),
				TargetCreature("target creature")),
		},
	})
}
