package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lobelia Sackville-Baggins — Legendary Creature — Halfling Citizen {2}{B},
// 2/3 (The Lord of the Rings: Tales of Middle-earth):
//
//	"Flash
//	 Menace
//	 When Lobelia enters, exile target creature card from an opponent's
//	 graveyard that was put there from the battlefield this turn, then
//	 create X Treasure tokens, where X is the exiled card's power."
//
// The target clause is Moira, Urborg Haunt's (PutThereFromTheBattlefieldThisTurn
// reads the turn's record of cards that moved from the battlefield to a
// graveyard) over a creature card an opponent owns. The card's power is read
// as it sits in the graveyard, before it moves (CR 608.2h), and carried to
// the tail, which creates the Treasure tokens only if the card was really
// exiled ("then"). A card that is no longer a legal target as the ability
// resolves is not exiled and makes no Treasure (CR 608.2b).
//
// No simplification.
func init() {
	enters := effects.WhenThisEnters("Lobelia Sackville-Baggins — exile target creature card from an opponent's graveyard that was put there from the battlefield this turn, then create X Treasure tokens",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			t, ok := effects.FirstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			c, found := g.LookupCardForEffect(t.ID)
			if !found {
				return nil
			}
			return effects.ExileTarget{Target: t.ID, ThenTail: ctx.TailN(lobeliaTreasures, c.CurrentPower())}.Apply(ctx)
		})
	enters.Targets = effects.TargetCardInGraveyard("target creature card from an opponent's graveyard that was put there from the battlefield this turn",
		effects.Creature(), effects.Not(effects.YouOwn()), effects.PutThereFromTheBattlefieldThisTurn)
	effects.Register(effects.Spec{
		OracleID:        "19f5a25a-97fc-41e6-bb26-7c35f4b4e37c",
		Name:            "Lobelia Sackville-Baggins",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "menace"},
		Triggered:       []game.TriggeredAbility{enters},
	})
}

// lobeliaTreasures is "then create X Treasure tokens, where X is the exiled
// card's power", X carried on the tail's N and told only when the card was
// exiled. An on-disk key: never renamed, never reused.
var lobeliaTreasures = effects.CardTail("then/lobelia-treasures", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 || o.Tail.N <= 0 {
		return nil
	}
	return effects.CreateToken{Template: effects.TreasureToken(), N: o.Tail.N}.Apply(ctx)
})
