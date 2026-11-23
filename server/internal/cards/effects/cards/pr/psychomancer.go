package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Psychomancer — Artifact Creature — Necron Wizard {1}{B}, 1/1 (Warhammer
// 40,000 Commander, S82 GAP batch 188, #5948):
//
//	"Flying
//	 Harbinger of Despair — Whenever this creature or another nontoken
//	 artifact you control is put into a graveyard from the battlefield or
//	 is put into exile from the battlefield, target opponent loses 1 life
//	 and you gain 1 life."
//
// A leaves-the-battlefield trigger (CR 603.10a) that reads the permanent as
// it last existed: its controller and its artifact type. The destination is
// the graveyard or exile, never the hand or library; a token is not a
// nontoken artifact (the Psychomancer's own death counts even if it is a
// token, since "this creature" is named).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "589b742c-fee5-413a-a56c-4d4ffaf0f73d",
		Name:            "Psychomancer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventLTB, gapB188HarbingerLeft,
				"Psychomancer — target opponent loses 1 life and you gain 1 life",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					t, ok := effects.FirstLegalTarget(ctx)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					if err := (effects.GainLife{Player: t.ID, Amount: -1}).Apply(ctx); err != nil {
						return err
					}
					return effects.GainLife{Amount: 1}.Apply(ctx)
				}), effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}

// gapB188HarbingerLeft is Psychomancer's condition.
func gapB188HarbingerLeft(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventLTB || (ev.NewZone != game.ZoneGraveyard && ev.NewZone != game.ZoneExile) {
		return false
	}
	if ev.CardID == source.InstanceID {
		return true
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || effects.IsToken(c) || effects.LeftUnderControlOf(ev, c) != source.Controller {
		return false
	}
	return effects.LeftAsType(ev, c, "artifact")
}
