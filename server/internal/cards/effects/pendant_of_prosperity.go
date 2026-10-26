package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pendant of Prosperity — Artifact {3}:
//
//	"This artifact enters under the control of an opponent of your
//	 choice.
//	 {2}, {T}: Draw a card, then you may put a land card from your hand
//	 onto the battlefield. This artifact's owner draws a card, then
//	 that player may put a land card from their hand onto the
//	 battlefield."
//
// A group-hug gift: the first line is ADR 0102's
// EntersUnderTheControlOfAnOpponentOfYourChoice, declared as a BENEFIT
// so a bot hands it to its weakest opponent. The activation is then the
// controller's ("you"), and the second half is the OWNER's — the player
// who cast it, whatever has happened to control since (CR 108.3).
//
// The order is printed and it is load-bearing: each "you may put a
// land" is a prompt, so the owner's half runs from the controller's
// land prompt's continuation rather than on the next line — otherwise
// the owner would draw before the controller had decided. The 2019
// ruling is the case where one player is both: "If you control your
// own Pendant of Prosperity … you'll draw a card, then you may put a
// land card onto the battlefield, then you'll draw a second card, then
// you may once again put a land card onto the battlefield." None of it
// is a land PLAY, so neither put spends a land drop.
//
// The owner is read when the ability resolves, from the card wherever
// it is — ownership never changes, and a Pendant that has left the
// battlefield in response still names the same player.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d958321b-789d-4f9a-bdbe-f907166a0216",
		Name:         "Pendant of Prosperity",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			EntersUnderTheControlOfAnOpponentOfYourChoice("Pendant of Prosperity", game.ControlForBenefit),
		},
		Activated: []ActivatedAbility{{
			Label: "{2}, {T}: Draw a card, then you may put a land card from your hand onto the battlefield. " +
				"This artifact's owner draws a card, then that player may put a land card from their hand onto the battlefield.",
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			// The land prompt waits for a prompt the draw asks (a dredge offer; CR
			// 608.2c, #4155).
			Effect: Do(
				DrawCards{N: 1},
				Instruction(func(ctx *Context) error {
					item := ctx.Item
					put := MayPutALandFromHand("Pendant of Prosperity")
					put.Player = item.Controller
					put.ThenTail = pendantOwnersHalf
					put.ThenItem = true
					return put.Apply(ctx)
				}),
			),
		}},
	})
}

// pendantOwnersHalf is the ability's second sentence, run once the
// controller has answered their land prompt: the owner draws, then may
// put a land. A keyed tail (#4253 slice 6), told the asking item; the owner
// is read off the artifact again, as the ability read it. An on-disk key.
var pendantOwnersHalf = PutFromHandTail("pendant-of-prosperity/owners-half", func(ctx *Context, _ PutFromHandResult) error {
	g, item := ctx.Game, ctx.Item
	owner := item.Controller
	if c, ok := g.LookupCardForEffect(item.SourceCardID); ok && c.Owner != uuid.Nil {
		owner = c.Owner
	}
	if p := g.PlayerByIDForEffect(owner); p == nil || p.Eliminated {
		return nil
	}
	ctx.RecordStep(pendantOwnerRecord, []uuid.UUID{owner})
	return pendantOwnersDrawThenPut.Run(ctx)
})

// pendantOwnerRecord is the artifact's owner.
const pendantOwnerRecord = "pendant-of-prosperity/owner"

// pendantOwnersDrawThenPut is the owner's draw and then their land prompt: the
// prompt waits for a prompt the draw asks (a dredge offer; CR 608.2c, #4155).
var pendantOwnersDrawThenPut = RegisterThenSteps("then-steps/pendant-of-prosperity-owner",
	Instruction(func(ctx *Context) error {
		if ids := ctx.StepRecord(pendantOwnerRecord).IDs; len(ids) > 0 {
			return DrawCards{Player: ids[0], N: 1}.Apply(ctx)
		}
		return nil
	}),
	Instruction(func(ctx *Context) error {
		ids := ctx.StepRecord(pendantOwnerRecord).IDs
		if len(ids) == 0 {
			return nil
		}
		put := MayPutALandFromHand("Pendant of Prosperity")
		put.Player = ids[0]
		// #1997: "that player may put a land card from their hand" —
		// the owner puts it (CR 110.2a), which Arboria reads.
		return ctx.Game.PutByEnteringPlayerForEffect(func() error { return put.Apply(ctx) })
	}),
)
