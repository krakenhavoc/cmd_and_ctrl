package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

func init() {
	pump := TriggerWithPurpose(AtBeginningOfYourCombat("Howling Moon — a Wolf or Werewolf gets +2/+2", func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := b16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return (BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Howling Moon — +2/+2"}).Apply(ctx)
	}), ForTargets(BenefitUntilEOT(0, 2, 2)))
	pump.Targets = TargetCreature("target Wolf or Werewolf you control", YouControl(), Or(HasSubtype("Wolf"), HasSubtype("Werewolf")))
	Register(Spec{
		OracleID: "a814d4c7-2ec0-4278-8704-23b39dcead4e", Name: "Howling Moon", Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{pump,
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return opponentCastTheirSecondSpellThisTurn(ev, source, g)
			}, "Howling Moon — create a Wolf", Do(CreateToken{Template: TokenCard("2/2 green Wolf"), N: 1})),
		},
	})
}
