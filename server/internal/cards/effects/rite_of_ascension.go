package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rite of Ascension — Enchantment {2}{R}{W} (Star Trek Commander, #4974):
//
//	"Battlefront — Whenever you attack, for each opponent you attacked
//	 this turn, create a tapped and attacking 2/1 red Klingon Warrior
//	 creature token with haste."
//
// "Battlefront" is an ability word (CR 207.2c). One trigger per attack
// declaration; as it resolves it counts the opponents still in the game
// a creature of yours was declared attacking this turn, this combat's
// and any earlier one's (an attack on a planeswalker or battle is not an
// attack on its player). That many printed Klingon Warriors enter tapped
// and attacking; the effect names nothing for them to attack, so their
// controller chooses for each (CR 508.4). They were never declared, so
// no "whenever a creature attacks" trigger sees them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "08c6ca45-09ff-4fe4-bc9b-45876e94f430",
		Name:         "Rite of Ascension",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			trekB40YouAttacked("Rite of Ascension — battlefront: a tapped and attacking 2/1 Klingon Warrior for each opponent you attacked this turn",
				func(g *game.Game, item *game.StackItem) error {
					return CreateTokensAttackingYourChoice{
						Template: TokenCard("2/1 red Klingon Warrior with haste"),
						N:        OpponentsYouAttackedThisTurn(g, item.Controller),
						Tapped:   true,
						Label:    "Rite of Ascension",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
