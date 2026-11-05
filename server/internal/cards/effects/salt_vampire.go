package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Salt Vampire — Creature — Vampire Shapeshifter {1}{B}, 2/2 (Star Trek,
// #4973):
//
//	"Lifelink
//	 When this creature enters, look at target opponent's hand. You may
//	 choose a nonland card from it. If you do, that player exiles that
//	 card, then draws a card."
//
// "Look at" shows the hand to you alone (ChooseFromRevealedHand.Look);
// the choice is optional, and the draw follows only a card actually
// chosen and exiled. The exile is not a discard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "31c7453e-33b8-4eb6-abac-28cdb35fb993",
		Name:            "Salt Vampire",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Salt Vampire — look at target opponent's hand; you may exile a nonland card from it, then they draw",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						return ChooseFromRevealedHand{
							Player:   TargetedPlayer(ctx),
							Filter:   Nonland(),
							Label:    "nonland card",
							Look:     true,
							Optional: true,
							Exile:    true,
							Then:     saltVampireDraw,
						}.Apply(ctx)
					}),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}

// saltVampireDraw is "If you do, ... then draws a card": the player whose
// card was exiled draws one. An on-disk key.
var saltVampireDraw = RevealedPickThen("revealed-pick/salt-vampire-draw", func(ctx *Context, pick game.RevealedPick) error {
	if len(pick.Chosen) == 0 {
		return nil
	}
	return DrawCards{Player: pick.FromPlayer, N: 1}.Apply(ctx)
})
