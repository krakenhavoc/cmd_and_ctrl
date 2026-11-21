package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tales of Master Seshiro // Seshiro's Living Legacy — a transforming Saga
// (Kamigawa: Neon Dynasty, #5493, S82).
//
// Front face, Enchantment — Saga {4}{G}:
//
//	"I, II — Put a +1/+1 counter on target creature or Vehicle you control.
//	 It gains vigilance until end of turn.
//	 III — Exile this Saga, then return it to the battlefield transformed
//	       under your control."
//
// Back face, Enchantment Creature — Snake Warrior, 5/5: "Vigilance, haste".
//
// A Vehicle that isn't a creature is a legal target and takes the counter
// and the keyword. If the target is gone as the chapter resolves, neither
// half happens (CR 608.2b).
//
// No simplification.
const TalesOfMasterSeshiroOracleID = "096fbb51-be73-4edc-a051-770a12c09005"

func init() {
	const name = "Tales of Master Seshiro"
	body := func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		target := effects.FirstLegalBattlefieldTarget(ctx)
		if target == [16]byte{} {
			return nil
		}
		if err := (effects.AddCounter{Target: target, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
			return err
		}
		return effects.GrantKeywordUntilEOT{Target: target, Keywords: []string{"vigilance"}, Label: name + " — vigilance"}.Apply(ctx)
	}
	chapter := func(n int) game.TriggeredAbility {
		return effects.ChapterTriggerTargeting(n, effects.SagaChapterLabel(name, n, "a +1/+1 counter on a creature or Vehicle, and vigilance"),
			effects.TargetPermanent("target creature or Vehicle you control", effects.Or(effects.Creature(), effects.OfSubtype("Vehicle")), effects.YouControl()), body)
	}
	effects.Register(effects.Spec{
		OracleID:     TalesOfMasterSeshiroOracleID,
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			chapter(1), chapter(2),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "exile it, then return it transformed"),
				effects.ChapterExileAndReturnTransformed),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        TalesOfMasterSeshiroOracleID + "#1",
		Name:            "Seshiro's Living Legacy",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "haste"},
	})
}
