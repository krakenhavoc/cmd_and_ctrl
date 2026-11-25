package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Covenant of Minds — Sorcery {4}{U} (Conspiracy: Take the Crown, S82 older
// sets batch 271, #5540):
//
//	"Reveal the top three cards of your library. Target opponent may choose
//	 to put those cards into your hand. If they don't, put those cards into
//	 your graveyard and draw five cards."
//
// Combustible Gearhulk's question put to the chosen opponent, asked as the
// spell resolves, with the three revealed cards riding the prompt as data
// (MayChoice.Carry) so "those cards" is exactly the revealed ones whenever
// the answer comes. Yes puts all three into the caster's hand (no prompt,
// the caster takes them, CR 701.20 reveals them first); no puts them into
// the graveyard, not milled, and then the caster draws five. A library of
// fewer than three reveals what it has. A target that is gone by resolution
// makes the whole spell do nothing, not the "doesn't" branch (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3edac7df-2a88-4718-865e-703dff9c972e",
		Name:         "Covenant of Minds",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 5},
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			opp := uuid.Nil
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetPlayer {
					opp = t.ID
					break
				}
			}
			if opp == uuid.Nil {
				return nil
			}
			var revealed []uuid.UUID
			if err := (effects.RevealTopOfLibrary{N: 3, Reason: "Covenant of Minds — the top three cards of your library", Revealed: &revealed}).Apply(ctx); err != nil {
				return err
			}
			return effects.MayChoice{
				Player:   opp,
				Question: "Covenant of Minds — put the revealed cards into its controller's hand? If you don't, they go to the graveyard and its controller draws five cards.",
				YesLabel: "They put the cards into their hand",
				NoLabel:  "They put the cards into their graveyard and draw five",
				Carry:    revealed,
				Then:     covenantOfMindsAnswer,
			}.Apply(ctx)
		},
	})
}

// covenantOfMindsTake is the yes branch's take: every revealed card to the
// hand, no prompt and nothing left over. An on-disk key.
var covenantOfMindsTake = effects.RegisterTake("choose-cards/old-b271-covenant-of-minds-all-to-hand", effects.TakeDef{})

// covenantOfMindsAnswer is the chosen opponent's answer. Keyed.
var covenantOfMindsAnswer = effects.MayYesNo("confirm/old-b271-covenant-of-minds",
	func(ctx *effects.Context, r game.ConfirmAnswer) error {
		return effects.TakeFromLibraryToHand{
			Player: ctx.Controller(),
			Cards:  r.Carry,
			All:    true,
			Label:  "Covenant of Minds — put the revealed cards into your hand",
			Key:    covenantOfMindsTake,
		}.Apply(ctx)
	},
	func(ctx *effects.Context, r game.ConfirmAnswer) error {
		if err := effects.RestIntoGraveyard(ctx.Game, r.Carry); err != nil {
			return err
		}
		return effects.DrawCards{Player: ctx.Controller(), N: 5}.Apply(ctx)
	})
