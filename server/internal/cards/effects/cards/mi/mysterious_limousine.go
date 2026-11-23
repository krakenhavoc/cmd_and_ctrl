package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const mysteriousLimousineKey = "Mysterious Limousine — exile up to one other target creature until this Vehicle leaves the battlefield; if a creature is exiled this way, return each other card exiled with it"

// Mysterious Limousine — Artifact — Vehicle {3}{W}{W}, 4/4 (Streets of New
// Capenna, #5491):
//
//	"Whenever this Vehicle enters or attacks, exile up to one other target
//	 creature until this Vehicle leaves the battlefield. If a creature is put
//	 into exile this way, return each other card exiled with this Vehicle to
//	 the battlefield under its owner's control.
//	 Crew 2"
//
// One trigger over two events. The exile is ExileUntil{ThisLeaves} with the CR
// 607.2a link stamped (Assimilation Aegis' shape), so each card comes back
// when this Vehicle leaves and is "exiled with" it by this row until then. The
// "if a creature is put into exile this way" rider is the exile's keyed tail,
// told the card only when it reached exile (CR 610.3b: a Vehicle already gone
// exiles nothing, so nothing returns): it puts every OTHER card linked to this
// Vehicle by this row back onto the battlefield together under its owner's
// control (CR 610.3c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "864a9276-2fec-49b3-bc53-008608de5ab6",
		Name:         "Mysterious Limousine",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEntersOrAttacks(mysteriousLimousineKey, mysteriousLimousineExile),
				effects.Another(effects.UpToOneTargetCreature("up to one other target creature"))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    effects.CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Mysterious Limousine"),
		}},
	})
}

func mysteriousLimousineExile(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return effects.ExileUntil{
				Target:     t.ID,
				ThisLeaves: true,
				Linked:     true,
				Label:      "Mysterious Limousine — the exiled card returns when Mysterious Limousine leaves the battlefield",
				ThenTail:   ctx.Tail(mysteriousLimousineReturnOthers),
			}.Apply(ctx)
		}
	}
	return nil
}

// mysteriousLimousineReturnOthers is "if a creature is put into exile this
// way, return each other card exiled with this Vehicle to the battlefield
// under its owner's control". An on-disk key: never renamed, never reused.
var mysteriousLimousineReturnOthers = effects.CardTail("exile-then/mysterious-limousine-return-others", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	src, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var others []uuid.UUID
	for _, c := range ctx.Game.ExiledWithRowForEffect(game.PermissionCardRef(src), game.AbilitySlotTriggered, mysteriousLimousineKey) {
		if c.InstanceID != o.IDs[0] {
			others = append(others, c.InstanceID)
		}
	}
	if len(others) == 0 {
		return nil
	}
	return effects.ReturnFromExileTogether{Targets: others}.Apply(ctx)
})
