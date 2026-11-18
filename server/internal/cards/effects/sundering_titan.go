package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sundering Titan — Artifact Creature — Golem {8}, 7/10:
//
//	"When this creature enters or leaves the battlefield, choose a land
//	 of each basic land type, then destroy those lands."
//
// One enters-or-leaves trigger (CR 603.6a, 603.6c). The choices are made
// by the ability's controller from every land on the battlefield, one
// leg per basic land type (CR 305.6) in the order the rule lists them,
// each from the lands that have that type when the ability resolves. A
// type with no land is skipped. The legs are independent, so a land with
// several basic land types may be chosen for more than one of them; it
// is destroyed once. Everything chosen is destroyed as one event.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d17d071d-dbee-4d65-8435-137f35cd89ca",
		Name:         "Sundering Titan",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrLeaves("Sundering Titan — choose a land of each basic land type, then destroy those lands", sunderingTitanChoose),
		},
	})
}

func sunderingTitanChoose(g *game.Game, item *game.StackItem) error {
	var legs []game.ChooseCardsPrompt
	for _, typ := range game.BasicLandTypes {
		var cands []uuid.UUID
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.IsLand() && c.HasSubtype(typ) {
				cands = append(cands, c.InstanceID)
			}
		}
		if len(cands) == 0 {
			continue
		}
		legs = append(legs, game.ChooseCardsPrompt{
			Chooser:    item.Controller,
			FromPlayer: item.Controller,
			Source:     item.SourceCardID,
			Question:   "Sundering Titan — choose a land of type " + typ + "; it will be destroyed",
			Cards:      cands,
			Min:        1,
			Max:        1,
			Zone:       game.ZoneBattlefield,
		})
	}
	if len(legs) == 0 {
		return nil
	}
	return g.ChoosePermanentsInOrderKeyedForEffect(legs, sunderingTitanDestroy, item.SourceCardID, game.RunData{})
}

// sunderingTitanDestroy is "then destroy those lands": every land chosen,
// once, as one simultaneous event. An on-disk key.
var sunderingTitanDestroy = game.RegisterPromptRunThen("prompt-run/sundering-titan-destroy", func(g *game.Game, r game.RunDone) error {
	seen := map[uuid.UUID]bool{}
	var ids []uuid.UUID
	for _, id := range r.Picks().Cards() {
		if !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return g.DestroyPermanentsKeyedForEffect(ids, game.KeyedTail{})
})
