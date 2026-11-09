package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eloren Wilds — Plane — Shandalar (#4870):
//
//	"Whenever a player taps a permanent for mana, that player adds one mana of
//	 any type that permanent produced.
//	 Whenever chaos ensues, target player can't cast spells until a player
//	 planeswalks."
//
// The first ability is Mana Flare's triggered mana ability over every
// permanent, not just lands (CR 605.1b, 605.4a): it is read from the face-up
// plane (game.manaTriggerFiringsLocked), fires only for a mana ability with a
// {T} cost (the engine fires triggered mana abilities for no other), and gives
// the mana to whoever tapped. The chaos ability puts a cast ban on the target
// player that ends when the plane stops being face up (UntilYouPlaneswalkAway).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "23fe1c21-c396-4ebf-a8b0-22d697952b34",
		Name:         "Eloren Wilds",
		Completeness: CompletenessFull,
		ManaTriggers: []game.ManaTrigger{{
			Label: "Eloren Wilds — add one mana of any type that permanent produced",
			AppliesTo: func(game.ManaProduced, *game.Card, *game.Game) bool {
				return true
			},
			Produced: AddsOneManaOfAnyTypeProduced(),
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverChaosEnsues("Eloren Wilds — target player can't cast spells until a player planeswalks",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					t, ok := ctx.ClauseTarget(0)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					d, ok := UntilYouPlaneswalkAway(g, item.SourceCardID)
					if !ok {
						return nil
					}
					g.GrantCastBanForEffect(t.ID, game.CastBanRule{Kind: game.CastBanOutright},
						"Eloren Wilds — can't cast spells until a player planeswalks", item.SourceCardID, d)
					return nil
				}),
				TargetPlayer("target player")),
		},
	})
}
