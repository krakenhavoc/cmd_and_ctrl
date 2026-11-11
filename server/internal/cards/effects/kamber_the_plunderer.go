package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kamber, the Plunderer — Legendary Creature — Vampire Rogue {3}{B}, 3/4
// (Outlaws of Thunder Junction Commander, #5368):
//
//	"Partner with Laurine, the Diversion (When this creature enters,
//	 target player may put Laurine into their hand from their library,
//	 then shuffle.)
//	 Lifelink
//	 Whenever a creature an opponent controls dies, you gain 1 life and
//	 create a Blood token."
//
// Partner with is Toothy, Imaginary Friend's pair of abilities. The death
// trigger is Moonstone Eulogist's condition (the creature left the
// battlefield for a graveyard under an opponent's control, last-known,
// CR 603.10a): one trigger per creature, the life gained first, then the
// Blood token.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0b620775-70a5-4423-8453-e403a5b79554",
		Name:            "Kamber, the Plunderer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			PartnerWith("Kamber, the Plunderer", "Laurine, the Diversion"),
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b18OpponentsCreatureDied(ev, source, g)
			}, "Kamber, the Plunderer — you gain 1 life and create a Blood token",
				Do(GainLife{Amount: 1}, CreateToken{Template: BloodToken(), N: 1})),
		},
	})
}
