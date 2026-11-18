package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Park Heights Maverick — Creature — Human Soldier {2}{G}, 2/2 (Streets of
// New Capenna Commander):
//
//	"Dethrone (Whenever this creature attacks the player with the most
//	 life or tied for most life, put a +1/+1 counter on it.)
//	 This creature can't be blocked by creatures with power 2 or less.
//	 Whenever this creature deals combat damage to a player or dies,
//	 proliferate."
//
// Dethrone is Scourge of the Throne's trigger (CR 702.105a): the defender
// is checked once, as the attack is declared, and an attack at a
// planeswalker or battle never triggers it. The block restriction is
// Bristlebane Outrider's, judged as blockers are declared on each
// would-be blocker's power. The last ability is one trigger on two
// conditions, combat damage to a player and dying; proliferate asks its
// controller which permanents and players to add a counter to.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5dd54b2a-7d6a-44f5-8827-ff3660a4b12b",
		Name:         "Park Heights Maverick",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnSelf(), PowerLE(2), "creatures with power 2 or less"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, preacherAttackedTheLifeLeader,
				"Park Heights Maverick — dethrone: put a +1/+1 counter on it",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			OnAny([]game.EventKind{game.EventDealDamage, game.EventLTB}, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return ThisDealtCombatDamageToAPlayer(ev, source, lki, g) || ThisDied(ev, source, lki, g)
			}, "Park Heights Maverick — proliferate", Do(Proliferate{})),
		},
	})
}
