package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mole Module — {5} Artifact — Vehicle 6/6 (Teenage Mutant Ninja Turtles
// Commander, #5011):
//
//	"Menace
//	 Whenever this Vehicle deals combat damage to a player, mill four
//	 cards. You may put a permanent card from among them onto the
//	 battlefield.
//	 Crew 2"
//
// The trigger is Six's mill-then-pick (MillToZone with its tail), with
// Colossal Grave-Reaver's keyed reanimation as the pick: the controller
// mills four (not optional), then is offered the permanent cards among the
// milled ones that are still in the graveyard (a permanent type, CR 110.4)
// and may take at most one, which enters under their control. Nothing to
// offer asks nothing. Crew 2 is the ordinary crew ability (CR 702.122).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ba64d3a9-08b4-4b12-b011-e41b00815277",
		Name:            "Mole Module",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Activated: []ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Mole Module"),
		}},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Mole Module — mill four cards, then you may put a permanent card from among them onto the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return MillToZone{N: 4, ThenTail: ctx.ItemTail(moleModuleMilled)}.Apply(ctx)
				}),
		},
	})
}

// moleModuleMilled is the clause after the mill: the may-pick of one
// permanent card from among the milled cards still in the graveyard. An
// on-disk key: never renamed, never reused.
var moleModuleMilled = CardTail("mole-module/after-mill", func(ctx *Context, o game.TailOutcome) error {
	var candidates []uuid.UUID
	for _, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || !isPermanentCard(c) {
			continue
		}
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		candidates = append(candidates, id)
	}
	if len(candidates) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    ctx.Item.Controller,
		FromPlayer: ctx.Item.Controller,
		Source:     ctx.Item.SourceCardID,
		Question:   "Mole Module — you may put a permanent card from among them onto the battlefield",
		Cards:      candidates,
		Min:        0,
		Max:        1,
		ThenKey:    b17ReanimatePicked,
		LiveZone:   game.ZoneGraveyard,
		Asked:      game.NewAskedItem(ctx.Item),
	})
	return nil
})
