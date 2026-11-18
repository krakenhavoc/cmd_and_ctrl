package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dawnhart Disciple — Creature — Human Warlock {1}{G}, 2/2 (Innistrad:
// Crimson Vow, #5495):
//
//	"Whenever another Human you control enters, this creature gets +1/+1
//	 until end of turn."
//
// Youthful Valkyrie's "another <type> you control enters" condition for
// Humans (post-layer subtypes, so a changeling and a token count), with
// Slickshot Show-Off's self pump.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e86279ba-cb22-4754-90f4-38354657b8fa",
		Name:         "Dawnhart Disciple",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev, source, g, source.TextSubtype("Human"))
			}, "Dawnhart Disciple — gets +1/+1 until end of turn",
				thisGetsUntilEndOfTurn(1, 1, "Dawnhart Disciple — +1/+1")),
		},
	})
}
