package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Powder Keg — Artifact {2} (Urza's Destiny, S82 #5621):
//
//	"At the beginning of your upkeep, you may put a fuse counter on this artifact.
//	 {T}, Sacrifice this artifact: Destroy each artifact and creature with mana
//	 value equal to the number of fuse counters on this artifact."
//
// Ratchet Bomb's sweep over artifacts and creatures: the sacrifice is a cost, so
// the count is the Keg's last-known information (CR 608.2h), read through the
// source permanent rather than the graveyard card. The Keg is already gone and
// is not destroyed again; a token's mana value is 0, so a Keg with no counters
// destroys every artifact and creature token and every permanent of mana value
// 0, and an indestructible permanent survives it. The upkeep counter is optional.
//
// No simplification.
func init() {
	upkeep := effects.AtYourUpkeep("Powder Keg — you may put a fuse counter on it", func(g *game.Game, item *game.StackItem) error {
		return effects.AddCounter{Target: item.SourceCardID, Kind: "fuse", N: 1}.Apply(effects.NewContext(g, item))
	})
	upkeep.OptionalPrompt = &game.TriggerOptionalPrompt{Question: "Powder Keg — put a fuse counter on it?"}
	effects.Register(effects.Spec{
		OracleID:     "ca1144c2-49a3-49b4-9085-3094141769ea",
		Name:         "Powder Keg",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{upkeep},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice this artifact: Destroy each artifact and creature with mana value equal to the number of fuse counters on this artifact.",
			Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy, Partial: true}},
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				info, ok := ctx.SourcePermanent()
				if !ok {
					return nil
				}
				return effects.DestroyAllMatching{Match: effects.And(effects.Or(effects.Artifact(), effects.Creature()), effects.B03ManaValueIs(info.Counters["fuse"]))}.Apply(ctx)
			},
		}},
	})
}
