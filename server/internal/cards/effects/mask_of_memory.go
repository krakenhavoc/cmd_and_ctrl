package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mask of Memory — Artifact — Equipment for {2} (EDHREC rank 990):
//
//	"Whenever equipped creature deals combat damage to a player, you
//	 may draw two cards. If you do, discard a card.
//	 Equip {1}"
//
// The third user of attachedCreatureDealtCombatDamageToPlayer, after
// the two Swords — the shape that makes an Equipment a card-advantage
// engine rather than a stat stick, and the reason evasive
// one-drops are worth equipping.
//
// "You may draw two cards. If you do, discard a card" is a single
// linked clause: the discard is the price of the draw, not a separate
// instruction, so declining the draw declines the discard. The
// engine has no yes/no prompt at trigger resolution, so the "may" is
// resolved as YES — which is the choice a player makes essentially
// every time (two-for-one into a pitch of the worst card in hand)
// and, critically, is the choice that keeps the linked discard
// attached to its draw. Taking the draw and skipping the discard
// would be stronger than printed; skipping both would be weaker.
// Declared as a caveat because an empty-library or hellbent corner
// exists where a player would genuinely decline.
//
// The discard is random rather than chosen, which is the standing
// DiscardCards simplification across the catalog (a hand picker is
// not in the engine) and is strictly weaker than the printed "discard
// a card", where you choose.
func init() {
	Register(Spec{
		OracleID:     "d6b2c998-a226-426c-a40d-6e6007041bfe",
		Name:         "Mask of Memory",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The optional draw is always taken — you aren't asked, and the linked discard follows.",
			"The discarded card is chosen at random instead of by you.",
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			}, "Mask of Memory — draw two, discard one", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (DrawCards{Player: item.Controller, N: 2}.Apply(ctx)); err != nil {
					return err
				}
				return DiscardCards{Player: item.Controller, N: 1}.Apply(ctx)
			}),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{1}"),
		},
	})
}
