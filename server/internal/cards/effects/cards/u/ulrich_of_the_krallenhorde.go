package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

func init() {
	const oracle = "0d397c05-a680-4274-972f-6a5f778b5133"
	pump := effects.TriggerWithPurpose(effects.OnAny([]game.EventKind{game.EventETB, game.EventTransform}, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return ev.Kind == game.EventETB && ev.CardID == source.InstanceID || effects.ThisTransformedInto(0)(ev, source, lki, g)
	}, "Ulrich of the Krallenhorde — target creature gets +4/+4", func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		id, ok := effects.B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return (effects.BoostUntilEOT{Target: id, Power: 4, Toughness: 4, Label: "Ulrich of the Krallenhorde — +4/+4"}).Apply(ctx)
	}), effects.ForTargets(effects.BenefitUntilEOT(0, 4, 4)))
	pump.Targets = effects.TargetCreature("target creature")
	effects.Register(effects.Spec{OracleID: oracle, Name: "Ulrich of the Krallenhorde", Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{pump, effects.TransformAtUpkeepIfNoSpellsLastTurn("Ulrich of the Krallenhorde")},
	})
	fight := effects.On(game.EventTransform, effects.ThisTransformedInto(1), "Ulrich, Uncontested Alpha — fight a non-Werewolf", effects.B30SourceFightsFirstLegalTarget)
	fight.Targets = effects.TargetCreature("target non-Werewolf creature you don't control", effects.Not(effects.HasSubtype("Werewolf")), effects.OpponentControls())
	effects.Register(effects.Spec{OracleID: oracle + "#1", Name: "Ulrich, Uncontested Alpha", Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{effects.Optional(fight, "Ulrich, Uncontested Alpha — fight?"), effects.TransformAtUpkeepIfAPlayerCastTwoSpellsLastTurn("Ulrich, Uncontested Alpha")},
	})
}
