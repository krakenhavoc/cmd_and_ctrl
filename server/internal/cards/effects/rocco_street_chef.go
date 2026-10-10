package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rocco, Street Chef — Legendary Creature — Elf Druid {R}{G}{W}, 2/4
// (#2559):
//
//	"At the beginning of your end step, each player exiles the top card
//	 of their library. Until your next end step, each player may play
//	 the card they exiled this way.
//	 Whenever a player plays a land from exile or casts a spell from
//	 exile, you put a +1/+1 counter on target creature and create a Food
//	 token."
//
// The first ability is Memory Vessel's permission without its hand ban
// (EachPlayerExilesTopAndMayPlay): each player holds a permission over
// the card they exiled, stamped "until your next end step" against
// Rocco's controller (game.UntilYourNextEndStep), which from this end
// step is the end step of their next turn. A land is played and spends
// that player's land drop.
//
// The second watches every player, any grant: a card a Rocco handed
// out, an opponent's impulse, a foretold or warped card. The cast reads
// EventCast's origin and the land play Event.Played with an exile
// origin (aCardWasPlayedFromExile, Prosper's read for any player).
// "You" put the counter and make the Food, so both are Rocco's
// controller's; with no creature to target the ability is removed from
// the stack and makes no Food (CR 603.3d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "071f5b4b-6a1e-4dc3-9c08-b79713191811",
		Name:         "Rocco, Street Chef",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Rocco, Street Chef — each player exiles the top card of their library and may play it until your next end step",
				func(g *game.Game, item *game.StackItem) error {
					return EachPlayerExilesTopAndMayPlay(g, 1, g.UntilYourNextEndStepDuration(item.Controller))
				}),
			Targeting(OnAny([]game.EventKind{game.EventCast, game.EventZoneMove}, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return aCardWasPlayedFromExile(ev, g)
			}, "Rocco, Street Chef — put a +1/+1 counter on target creature and create a Food", roccoCounterAndFood),
				TargetCreature("target creature")),
		},
	})
}

// roccoCounterAndFood is the second ability's resolution.
func roccoCounterAndFood(g *game.Game, item *game.StackItem) error {
	if err := putPlusOneCounterOnEachLegalTarget(g, item); err != nil {
		return err
	}
	return CreateToken{Template: FoodToken(), N: 1}.Apply(NewContext(g, item))
}
