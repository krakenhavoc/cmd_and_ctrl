package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aerial Extortionist — Creature — Bird Soldier {3}{W}{W}, 4/3
// (EDHREC rank 2762):
//
//	"Flying
//	 Whenever this creature enters or deals combat damage to a
//	 player, exile up to one target nonland permanent. For as long as
//	 that card remains exiled, its owner may cast it.
//	 Whenever another player casts a spell from anywhere other than
//	 their hand, draw a card."
//
// The white "politics" removal: the permanent leaves, but its owner
// can buy it back — and buying it back is a cast from exile, which
// is the second ability's trigger, so the Extortionist's controller
// draws for it. Two abilities:
//
//   - The exile is one printed ability with two conditions (ETB and
//     combat damage to a player), so one declaration watches both
//     kinds. "Up to one" is a Min-0 target; with nothing chosen the
//     ability does nothing. The exiled card carries an unbounded,
//     cast-only grant for its OWNER (game.CastPermission with
//     WhileExiled), which the cast path clears as the card leaves
//     exile, so a card exiled again later by something else does not
//     inherit it.
//   - The draw watches every cast whose announce record says the
//     spell came from somewhere other than the caster's hand — the
//     command zone (a commander cast), a graveyard (flashback), or
//     exile (an impulse grant, or the Extortionist's own) — by a
//     player other than the controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7e7d7f4d-2c22-4803-ba5e-5eb653c25b5e",
		Name:            "Aerial Extortionist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventETB, game.EventDealDamage},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b26SelfEnteredOrDealtCombatDamageToPlayer(ev, source, g)
				},
				Targets: TargetPermanent("up to one target nonland permanent", Nonland()).WithCount(0, 1),
				Key:     "Aerial Extortionist — exile a nonland permanent; its owner may cast it",
				Effect:  b26ExileFirstLegalTargetOwnerMayCast,
			},
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b26AnotherPlayerCastFromOutsideHand(ev, source, g)
			}, "Aerial Extortionist — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
