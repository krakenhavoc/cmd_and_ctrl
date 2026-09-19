package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rise of the Witch-king — Sorcery {2}{B}{G} (EDHREC rank 1924):
//
//	"Each player sacrifices a creature of their choice. If you
//	 sacrificed a creature this way, you may return another permanent
//	 card from your graveyard to the battlefield."
//
// An edict that pays you back. The sacrifice is EachPlayerSacrifices
// with the controller included — every player picks their own, one
// prompt each, and a player with no creature is skipped (CR 701.21a).
//
// "IF YOU SACRIFICED A CREATURE THIS WAY" is the run's answer for the
// controller's seat (#1019, ADR 0013 §5x). It used to be "was the
// controller handed a prompt", which is a question about the
// QUESTION: the permanent came back before anybody had chosen
// anything, and a controller whose only creature left while another
// seat was being asked was paid out for a sacrifice that never
// happened. EachPlayerSacrificesThenForEffect runs the clause once
// every asked seat has answered AND the permanents they named have
// finished moving — so a sacrificed commander's CR 903.9 prompt holds
// the payout too, and a commander that takes the command zone still
// counts, because CR 701.17a's sacrifice is the move off the
// battlefield.
//
// Sandbox simplification, declared (the Mount Doom / Time Wipe
// posture): "you may return another permanent card from your
// graveyard" is a resolution-time choice, and the engine has no
// pick-from-graveyard prompt with a continuation for a spell. So the
// card to return is picked when the spell is cast, as an optional
// target ("up to one"). Three consequences, all weaker than printed:
// opponents see the pick before the spell resolves; the creature you
// sacrifice cannot be the card that comes back (it is not in the
// graveyard yet when you pick — "another" for free, but also no
// choosing it deliberately); and if the picked card left the
// graveyard in response the spell is countered by game rules (CR
// 608.2b) and the edict does not happen either, where the printed
// card would still make everyone sacrifice.
func init() {
	Register(Spec{
		OracleID:     "3c86541c-3601-4a38-8872-39705e41303a",
		Name:         "Rise of the Witch-king",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The permanent card to return is picked when you cast the spell rather than after the sacrifices, so opponents can respond to the choice, and the creature you sacrifice to it can't be the one that comes back.",
		},
		Targets: TargetCardInGraveyard("up to one permanent card in your graveyard to return", YouOwn(), Permanent()).WithCount(0, 1),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			controller := ctx.Controller()
			return ctx.Game.EachPlayerSacrificesThenForEffect(
				ctx.Source(), uuid.Nil,
				sacrificeSpec("a creature", Creature()),
				"Rise of the Witch-king — sacrifice a creature",
				func(g *game.Game, sacrificed game.PromptedSacrifices) error {
					if !sacrificed.Sacrificed(controller) {
						return nil
					}
					// A fresh Context bound to the same stack item:
					// an undo restores the game's fields in place, so
					// the *Game captured when the prompts went up can
					// be the wrong object by the time they are
					// answered (resumeClause's contract).
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield}.Apply(ctx)
					}
					return nil
				})
		},
	})
}
