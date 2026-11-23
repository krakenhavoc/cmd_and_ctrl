package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Magmatic Hellkite — Creature — Dragon {2}{R}{R}, 4/5 (Tarkir:
// Dragonstorm, #4463):
//
//	"Flying
//	 When this creature enters, destroy target nonbasic land an opponent
//	 controls. Its controller searches their library for a basic land card,
//	 puts it onto the battlefield tapped with a stun counter on it, then
//	 shuffles."
//
// Ghost Quarter's body with a mandatory search: the land's controller is
// read before the destroy and searches either way (a land that is
// indestructible still gets its controller a basic, the ruling), the land
// entering tapped with a stun counter, put on as it enters. A target that is
// gone or no longer a nonbasic land an opponent controls at resolution is
// illegal and the whole ability does nothing, no search included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2a3b29c3-f69d-4671-8907-600d21528153",
		Name:            "Magmatic Hellkite",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Magmatic Hellkite — destroy target nonbasic land an opponent controls; its controller searches for a basic land tapped with a stun counter",
					magmaticHellkiteDestroy),
				effects.TargetPermanent("target nonbasic land an opponent controls", effects.Land(), effects.B03Nonbasic(), effects.OpponentControls())),
		},
	})
}

func magmaticHellkiteDestroy(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	victim, ok := effects.ControllerOfTarget(ctx, id)
	if !ok {
		return nil
	}
	if err := (effects.DestroyTarget{Target: id}).Apply(ctx); err != nil {
		return err
	}
	return effects.SearchLibrary{
		Player:             victim,
		Predicate:          effects.IsBasicLand,
		Dest:               game.ZoneBattlefield,
		Limit:              1,
		Shuffle:            true,
		TappedOnEntry:      true,
		EntersWithCounters: map[string]int{game.CounterStun: 1},
		Reason:             "Magmatic Hellkite — a basic land card, tapped with a stun counter",
	}.Apply(ctx)
}
