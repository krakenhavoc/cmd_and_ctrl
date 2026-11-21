package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Greasefang, Okiba Boss — Legendary Creature — Rat Pilot {1}{W}{B}, 4/3
// (Neon Dynasty):
//
//	"At the beginning of combat on your turn, return target Vehicle card
//	 from your graveyard to the battlefield. It gains haste. Return it to
//	 its owner's hand at the beginning of your next end step."
//
// Admiral Brass's return-and-tail shape, mandatory. The haste has no
// printed duration, so it lasts as long as the object stays on the
// battlefield (CR 611.2a, CR 400.7: pinned to the entered object). The
// return is Zara's delayed bounce, scheduled for the controller's next end
// step and bound to the object that entered: a Vehicle that left and came
// back by then is a new object and stays (CR 400.7), and one that died does
// nothing (CR 603.7). A target that left the graveyard before resolution is
// not returned (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0129ffe2-389f-4e05-8a04-96461bac39a5",
		Name:         "Greasefang, Okiba Boss",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.AtBeginningOfYourCombat("Greasefang, Okiba Boss — return target Vehicle card from your graveyard to the battlefield; it gains haste; return it to its owner's hand at the beginning of your next end step",
				greasefangReturn),
				effects.TargetCardInGraveyard("target Vehicle card from your graveyard", effects.YouOwn(), effects.OfSubtype("Vehicle"))),
		},
	})
}

func greasefangReturn(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return g.ReturnFromGraveyardToBattlefieldKeyedForEffect(id, item.Controller, false, ctx.ItemTail(greasefangEntered))
}

// greasefangEntered is "It gains haste. Return it to its owner's hand at
// the beginning of your next end step." An on-disk key.
var greasefangEntered = effects.EnteredTail("greasefang-okiba-boss/haste-and-return", func(ctx *effects.Context, entered uuid.UUID, _ game.TailOutcome) error {
	g := ctx.Game
	if entered == uuid.Nil || !effects.OnBattlefield(g, entered) {
		return nil
	}
	c, ok := g.LookupCardForEffect(entered)
	if !ok {
		return nil
	}
	if err := (effects.ScopedEffectFor{
		Target:   entered,
		Mods:     []game.Mod{game.AddKeywordsMod("haste")},
		Duration: g.PinnedTo(game.IndefiniteDuration(), entered),
		Label:    "Greasefang, Okiba Boss — haste",
	}).Apply(ctx); err != nil {
		return err
	}
	return effects.ScheduleDelayedTrigger{
		At:                 game.StepEnd,
		Controller:         ctx.Item.Controller,
		ControllerTurnOnly: true,
		Label:              "Greasefang, Okiba Boss — return it to its owner's hand",
		Cards:              []uuid.UUID{entered},
		Body:               effects.ZaraReturnToHandBody,
		Params:             game.EffectParams{Object: game.ObjectRef{ID: entered, Epoch: c.ObjectEpoch}},
	}.Apply(ctx)
})
