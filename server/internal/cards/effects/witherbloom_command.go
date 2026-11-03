package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Witherbloom Command — Sorcery {B}{G} (Secrets of Strixhaven Commander,
// #4834):
//
//	"Choose two —
//	 • Target player mills three cards, then you return a land card from
//	   your graveyard to your hand.
//	 • Destroy target noncreature, nonland permanent with mana value 2 or
//	   less.
//	 • Target creature gets -3/-1 until end of turn.
//	 • Target opponent loses 2 life and you gain 2 life."
//
// A "choose two" command (Prismari Command's shape): each chosen bullet
// reads its own target group and is re-checked against its own clause on
// resolution (CR 608.2b). The first bullet's return is the CASTER's choice
// from the caster's own graveyard, asked after the mill (CR 608.2c) so a
// land the mill put there can be the one that comes back; it is not "may",
// so a graveyard holding a land owes one. It is not a target, so it still
// happens when the milled player is gone. The last bullet's gain waits for
// a prompt the life loss may ask (CR 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7cc81fb3-495c-47cd-8c56-96197d87e9e0",
		Name:         "Witherbloom Command",
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose two", 2, 2,
			witherbloomCommandMillMode(),
			ModeDoing("Destroy target noncreature, nonland permanent with mana value 2 or less.",
				TargetPermanent("target noncreature, nonland permanent with mana value 2 or less",
					Noncreature(), Nonland(), ManaValueLE(2)),
				DestroyTheModesTarget),
			ModeWithPurpose(ModeDoing("Target creature gets -3/-1 until end of turn.",
				TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					return BoostUntilEOT{Target: t.ID, Power: -3, Toughness: -1, Label: "Witherbloom Command"}.Apply(ctx)
				}), ForTargets(HarmUntilEOT(0, -3, -1))),
			// The gain waits for a prompt the life loss asks (a CR 616 order; CR 608.2c, #4155).
			ModeSteps("Target opponent loses 2 life and you gain 2 life.", TargetPlayer("target opponent", Opponent()),
				Instruction(func(ctx *Context) error {
					t, ok := ModeTarget(ctx, ctx.Occurrence())
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), t.ID, -2)
				}),
				socB15ModeYouGainLife(2)),
		),
	})
}

// witherbloomCommandMillMode is the first bullet, in its own function so the
// exit-payout guard reads its mill apart from the other bullets' targets: the
// land return is not gated on what the mill moved, it is "a land card from
// your graveyard" after the mill, whatever the mill did.
func witherbloomCommandMillMode() game.ModeOption {
	return ModeSteps("Target player mills three cards, then you return a land card from your graveyard to your hand.",
		TargetPlayer("target player"),
		Instruction(func(ctx *Context) error {
			t, ok := ModeTarget(ctx, ctx.Occurrence())
			if !ok || t.Kind != game.TargetPlayer {
				return nil
			}
			return MillCards{Player: t.ID, N: 3}.Apply(ctx)
		}),
		Instruction(func(ctx *Context) error {
			controller := ctx.Controller()
			candidates := graveyardCardIDs(ctx, controller, game.Card.IsLand)
			if len(candidates) == 0 {
				return nil
			}
			ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:    controller,
				FromPlayer: controller,
				Source:     ctx.Source(),
				Question:   "Witherbloom Command — return a land card from your graveyard to your hand",
				Cards:      candidates,
				Min:        1,
				Max:        1,
				ThenKey:    witherbloomCommandReturnLand,
				LiveZone:   game.ZoneGraveyard,
				Asked:      game.NewAskedItem(ctx.Item),
			})
			return nil
		}))
}

// witherbloomCommandReturnLand is the land card chosen returning to its
// owner's hand, keyed (#4253): the asking item rides the prompt as data.
var witherbloomCommandReturnLand = askedPick("choose-cards/witherbloom-command-land-to-hand", func(ctx *Context, r game.CardsChosen) error {
	for _, id := range r.Picked {
		if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
