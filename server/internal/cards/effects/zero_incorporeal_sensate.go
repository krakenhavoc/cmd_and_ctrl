package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zero, Incorporeal Sensate — Legendary Artifact Creature — Medusan
// Scientist {1}{W/B}{W/B}, 2/2 (Star Trek, #4973):
//
//	"Lifelink
//	 When Zero enters, look at target opponent's hand. You may exile a
//	 nonland card from it. For as long as that card remains exiled, its
//	 owner may play it. A spell cast this way costs {2} more to cast."
//
// Lifelink is the printed keyword. The enters trigger targets an
// opponent (CR 603.3d); the hand is LOOKED at, so only Zero's controller
// sees it (Word of Command's ChooseFromRevealedHand.Look), and the pick
// of one nonland card is optional. The chosen card stays put for the
// pick's continuation, which exiles it with a play permission for its
// OWNER (ExileWithPermission: GrantTo left nil is the owner) lasting
// while it stays in exile (CR 611.2b), and with
// game.CastPermission.CostMore 2: the spell's total cost is raised by
// {2} with the other increases (CR 601.2f), whatever else the cast pays.
// The permission is the card's, not Zero's, so it outlives Zero. A
// target that is no longer an opponent's, or an empty hand, does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "746e494e-fdc4-415e-8a82-d35e11ebeccb",
		Name:            "Zero, Incorporeal Sensate",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Zero, Incorporeal Sensate — look at target opponent's hand; you may exile a nonland card from it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return ChooseFromRevealedHand{
						Player:   TargetedPlayer(ctx),
						Filter:   Nonland(),
						Label:    "nonland card",
						Look:     true,
						Optional: true,
						Keep:     true,
						Then:     zeroIncorporealSensateExile,
					}.Apply(ctx)
				}),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}

// zeroIncorporealSensateExile is "exile [the chosen card]. For as long as
// that card remains exiled, its owner may play it. A spell cast this way
// costs {2} more to cast." An on-disk key: never renamed, never reused.
var zeroIncorporealSensateExile = RevealedPickThen("revealed-pick/zero-incorporeal-sensate-exile",
	func(ctx *Context, pick game.RevealedPick) error {
		if len(pick.Chosen) == 0 {
			return nil
		}
		return ExileWithPermission{
			Target:      pick.Chosen[0].InstanceID,
			WhileExiled: true,
			CostMore:    2,
		}.Apply(ctx)
	})
