package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Attended Healer — Creature — Kor Cleric {3}{W}, 2/3 (Jumpstart 2022, #5948):
//
//	"Whenever you gain life for the first time each turn, create a 1/1
//	 white Cat creature token.
//	 {2}{W}: Another target Cleric gains lifelink until end of turn."
//
// The first-life-gain trigger is Cat Collector's shared watcher without its
// "your turn" rider: it fires on the first life gain of any turn and not on
// later ones. The second ability gives another Cleric (any controller's)
// lifelink until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9442a18b-9093-4a10-ab61-6962f5af3bb6",
		Name:         "Attended Healer",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventChangeLife, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.SocB28FirstLifeGainThisTurn(ev, source, g)
			}, "Attended Healer — create a 1/1 white Cat creature token",
				effects.Do(effects.CreateToken{Template: effects.B36WhiteCatToken(), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}: Another target Cleric gains lifelink until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "lifelink")),
			Cost:    effects.ManaCost("{2}{W}"),
			Targets: effects.Another(effects.TargetCreature("another target Cleric", effects.HasSubtype("Cleric"))),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"lifelink"}, Label: "Attended Healer — lifelink"}.Apply(ctx)
			},
		}},
	})
}
