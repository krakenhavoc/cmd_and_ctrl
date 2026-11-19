package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// March from Velis Vel — Sorcery {2}{U} (Modern Horizons 3 Commander, #5948):
//
//	"Choose a nonbasic land type. Each land you control of that type
//	 becomes a copy of target creature you control until end of turn and
//	 gains haste until end of turn.
//	 Flashback {4}{U}"
//
// The target is a creature you control, chosen as the spell is cast. The
// type is chosen as it resolves (CR 608.2c) from CR 205.3i's land types other
// than the five basic ones (Vision Charm's option pick over game.LandTypes),
// and the lands that have it are fixed then (CR 611.2c). Each becomes a copy
// of the creature as it is now (CR 707.2, BecomeCopy with a duration; the
// copy replaces the land's own types, so it is a creature and no longer a
// land) and gains haste. A target that has left makes the spell do nothing
// (CR 608.2b). Flashback is the shared alternative cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "25ba6f57-3d50-4c3e-ab55-284e91d684a6",
		Name:             "March from Velis Vel",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Answers: game.AnswerValue},
		Targets:          TargetCreature("target creature you control", YouControl()),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{4}{U}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			var options []string
			for _, t := range game.LandTypes {
				if !game.IsBasicLandType(t) {
					options = append(options, t)
				}
			}
			return ChooseBasicLandTypeTail(ctx, "March from Velis Vel — choose a nonbasic land type", options, ctx.ItemTail(marchFromVelisVelChosen))
		},
	})
}

// marchFromVelisVelChosen is the type chosen (the tail's Word): each land its
// controller controls of that type becomes a copy of the target creature and
// gains haste, until end of turn. An on-disk key.
var marchFromVelisVelChosen = CardTail("then/march-from-velis-vel-type-chosen", func(ctx *Context, o game.TailOutcome) error {
	var creature uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			creature = t.ID
			break
		}
	}
	if creature == uuid.Nil || o.Tail.Word == "" {
		return nil
	}
	lands := permanentsControlledByMatching(ctx.Game, ctx.Controller(), And(Land(), HasSubtype(o.Tail.Word)))
	if len(lands) == 0 {
		return nil
	}
	if err := (BecomeCopy{
		Targets:  lands,
		Of:       creature,
		Duration: CopyUntilEndOfTurn,
		Label:    "March from Velis Vel — becomes a copy of the creature until end of turn",
	}).Apply(ctx); err != nil {
		return err
	}
	for _, id := range lands {
		if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "March from Velis Vel — haste"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
