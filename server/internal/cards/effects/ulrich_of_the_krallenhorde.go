package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

func init() {
	const oracle = "0d397c05-a680-4274-972f-6a5f778b5133"
	pump := TriggerWithPurpose(OnAny([]game.EventKind{game.EventETB, game.EventTransform}, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return ev.Kind == game.EventETB && ev.CardID == source.InstanceID || ThisTransformedInto(0)(ev, source, lki, g)
	}, "Ulrich of the Krallenhorde — target creature gets +4/+4", func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := b16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return (BoostUntilEOT{Target: id, Power: 4, Toughness: 4, Label: "Ulrich of the Krallenhorde — +4/+4"}).Apply(ctx)
	}), ForTargets(BenefitUntilEOT(0, 4, 4)))
	pump.Targets = TargetCreature("target creature")
	Register(Spec{OracleID: oracle, Name: "Ulrich of the Krallenhorde", Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{pump, TransformAtUpkeepIfNoSpellsLastTurn("Ulrich of the Krallenhorde")},
	})
	fight := On(game.EventTransform, ThisTransformedInto(1), "Ulrich, Uncontested Alpha — fight a non-Werewolf", b30SourceFightsFirstLegalTarget)
	fight.Targets = TargetCreature("target non-Werewolf creature you don't control", Not(HasSubtype("Werewolf")), OpponentControls())
	Register(Spec{OracleID: oracle + "#1", Name: "Ulrich, Uncontested Alpha", Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{Optional(fight, "Ulrich, Uncontested Alpha — fight?"), TransformAtUpkeepIfAPlayerCastTwoSpellsLastTurn("Ulrich, Uncontested Alpha")},
	})
}
