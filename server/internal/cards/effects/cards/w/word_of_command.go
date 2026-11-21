package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Word of Command — Instant {B}{B}:
//
//	"Look at target opponent's hand and choose a card from it. You
//	 control that player until Word of Command finishes resolving. The
//	 player plays that card if able. While doing so, the player can
//	 activate mana abilities only if they're from lands that player
//	 controls and only if mana they produce is spent to activate other
//	 mana abilities of lands the player controls and/or to play that
//	 card. If the chosen card is cast as a spell, you control the player
//	 while that spell is resolving."
//
// CR 723.2's limited control of a player (ADR 0166, #2164). You look at
// the hand (nobody else sees it) and choose any card, which stays where
// it is. Then you control that player while their forced play happens,
// still inside Word of Command's resolution: a controlled_play prompt
// that you answer by casting the card as them (CR 608.2g, so its type's
// timing doesn't matter and nobody gets priority after it), choosing
// its targets, modes, X and costs and which of their lands pay. Only
// their lands' mana abilities can be activated. A land is played instead,
// if it's their turn and they have a land play left (CR 305.2b, 305.3).
// The play can't be declined while it is possible (2011-01-01 ruling);
// when it isn't, nothing happens. A spell cast this way is theirs, and
// you control them again while it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e8ad3a77-b293-4d69-b080-27ca9f95d443",
		Name:         "Word of Command",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ChooseFromRevealedHand{
				Player: effects.TargetedPlayer(ctx),
				Label:  "card",
				Look:   true,
				Keep:   true,
				Then:   wordOfCommandPlay,
			}.Apply(ctx)
		},
	})
}

// wordOfCommandPlay is "You control that player until Word of Command
// finishes resolving. The player plays that card if able."
var wordOfCommandPlay = effects.RevealedPickThen("revealed-pick/word-of-command-play",
	func(ctx *effects.Context, pick game.RevealedPick) error {
		if len(pick.Chosen) == 0 {
			return nil
		}
		ctx.Game.BeginControlledPlayForEffect(game.ControlledPlay{
			Controller: pick.Chooser,
			Controlled: pick.FromPlayer,
			Source:     pick.Source,
			SourceName: "Word of Command",
			Card:       pick.Chosen[0].InstanceID,
		})
		return nil
	})
