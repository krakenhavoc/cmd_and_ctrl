package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Slicer, Hired Muscle // Slicer, High-Speed Antagonist (#3275, #3195,
// ADR 0152) — a More Than Meets the Eye card.
//
// Front face, Legendary Artifact Creature — Robot {4}{R}, 3/4:
//
//	"More Than Meets the Eye {2}{R}
//	 Double strike, haste
//	 At the beginning of each opponent's upkeep, you may have that player
//	 gain control of Slicer until end of turn. If you do, untap Slicer,
//	 goad it, and it can't be sacrificed this turn. If you don't, convert
//	 it."
//
// Back face, Legendary Artifact — Vehicle, 3/2:
//
//	"Living metal
//	 First strike, haste
//	 Whenever Slicer deals combat damage to a player, convert it at end
//	 of combat."
//
// "That player" is the upkeep's player, read off the trigger's event.
// The control change lasts until cleanup; the goad is the trigger
// controller's, so it lasts until their next turn (CR 701.15a) and
// Slicer must attack someone other than them. "Can't be sacrificed this
// turn" is CantBeSacrificedThisTurn (#3275), so the borrowing player
// can't feed it to a sacrifice cost or an edict before it goes home.
// "If you do" is the yes branch: with Slicer gone by then, neither the
// control change nor the convert finds it (CR 400.7).
//
// The back face's convert is a delayed trigger at end of combat with
// the turn-over count recorded when it is created (CR 701.28e), so two
// first-strike and regular hits turn it over once (ADR 0152 §3).
//
// No simplification.
const SlicerOracleID = "fa716557-8e75-4617-8310-ab23a9ca1af1"

func init() {
	effects.Register(effects.Spec{
		OracleID:         SlicerOracleID,
		Name:             "Slicer, Hired Muscle",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"double strike", "haste"},
		AlternativeCosts: []game.AlternativeCost{effects.MoreThanMeetsTheEye("{2}{R}")},
		Triggered: []game.TriggeredAbility{
			effects.AtEachOpponentsUpkeep("Slicer, Hired Muscle — you may have that player gain control of Slicer; if you don't, convert it",
				slicerUpkeep),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        SlicerOracleID + "#1",
		Name:            "Slicer, High-Speed Antagonist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", "haste"},
		Static:          []game.StaticAbility{effects.LivingMetal()},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Slicer, High-Speed Antagonist — convert it at end of combat",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ScheduleDelayedTrigger{
						At:    game.StepEndCombat,
						Label: "Slicer, High-Speed Antagonist — convert Slicer at end of combat",
						Body:  effects.TransformersConvertThisBody,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

func slicerUpkeep(g *game.Game, item *game.StackItem) error {
	player := item.Trigger.Event.Actor
	return effects.MayChoice{
		Question: "Slicer, Hired Muscle — have " + effects.AttackTargetName(g, player) + " gain control of Slicer until end of turn? (If you don't, convert it.)",
		YesLabel: "Hand it over",
		NoLabel:  "Convert it",
		Then:     slicerAnswer,
		Carry:    []uuid.UUID{player},
	}.Apply(effects.NewContext(g, item))
}

// slicerAnswer is the controller's choice: the player attacked (the carried
// ID) gains control of Slicer until end of turn, or it converts. Keyed
// (#4291).
var slicerAnswer = effects.MayYesNo("confirm/slicer-hired-muscle-hand-over-or-convert",
	func(ctx *effects.Context, r game.ConfirmAnswer) error {
		slicer, player := ctx.Item.SourceCardID, effects.CarryAt(r, 0)
		if ctx.IsNewSourceObject(slicer) || !effects.OnBattlefield(ctx.Game, slicer) {
			return nil
		}
		if err := (effects.GainControl{
			Target:     slicer,
			Controller: player,
			Label:      "Slicer, Hired Muscle — gain control until end of turn",
		}).Apply(ctx); err != nil {
			return err
		}
		if err := (effects.UntapTarget{Target: slicer}).Apply(ctx); err != nil {
			return err
		}
		if err := effects.GoadAndScheduleClear(ctx, []uuid.UUID{slicer}); err != nil {
			return err
		}
		return effects.CantBeSacrificedThisTurn{
			Target: slicer,
			Label:  "Slicer, Hired Muscle — can't be sacrificed this turn",
		}.Apply(ctx)
	},
	func(ctx *effects.Context, _ game.ConfirmAnswer) error { return effects.ConvertThis{}.Apply(ctx) })
