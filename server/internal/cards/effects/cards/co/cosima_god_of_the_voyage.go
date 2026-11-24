package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cosima, God of the Voyage // The Omenkeel — a modal double-faced card (oracle
// 648100cd…), one file for both faces; the back is the "<oracle>#1" row.
//
// Front, Legendary Creature — God {2}{U}, 2/4:
//
//	"At the beginning of your upkeep, you may exile Cosima. If you do, it gains
//	 'Whenever a land you control enters, if Cosima is exiled, you may put a
//	 voyage counter on it. If you don't, return Cosima to the battlefield with X
//	 +1/+1 counters on it and draw X cards, where X is the number of voyage
//	 counters on it.'"
//
// Back, Legendary Artifact — Vehicle {1}{U}, 3/3:
//
//	"Whenever a Vehicle you control deals combat damage to a player, that
//	 player exiles that many cards from the top of their library. You may play
//	 lands from among those cards for as long as they remain exiled.
//	 Crew 1"
//
// The upkeep ability exiles Cosima linked to herself by that very row (CR
// 607.2a); that link is what "it gains" means here, because it is stamped only
// by this exile and cleared when the card leaves exile (CR 400.7), so a Cosima
// exiled any other way has no such ability. The gained ability is an exile-zone
// trigger over a land entering under her owner's control, which asks the one
// question the text asks: put a voyage counter on her, or (declining) return her
// as a new object with a +1/+1 counter for each voyage counter and draw that many
// cards. The Omenkeel's trigger reads any Vehicle its controller controls (itself
// included, crewed) dealing combat damage to a player: that player exiles that
// many cards from the top of their library and the Omenkeel's controller may
// play the lands among them while they stay exiled.
//
// No simplification.
func init() {
	const oracle = "648100cd-d44a-48f8-87e4-a1d5467f348c"
	upkeep := effects.Optional(effects.AtYourUpkeep(effects.OldB071CosimaUpkeepKey,
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			if !effects.OnBattlefield(g, item.SourceCardID) || ctx.IsNewSourceObject(item.SourceCardID) {
				return nil
			}
			return effects.ExileTarget{Target: item.SourceCardID, Linked: true}.Apply(ctx)
		}), "Cosima, God of the Voyage — exile Cosima? (she gains the voyage ability)")
	gained := effects.InExile(effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
		return ok && c.IsLand() && effects.OldB071CosimaIsGainedExile(*source, g)
	}, "Cosima, God of the Voyage — a land entered: put a voyage counter on Cosima, or return her and draw", effects.OldB071CosimaLandEntered))
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Cosima, God of the Voyage",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{upkeep, gained},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "The Omenkeel",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.OldB071OmenkeelTriggered,
				"The Omenkeel — that player exiles that many cards from the top of their library; you may play lands from among them", effects.OldB071OmenkeelExile),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    effects.CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("The Omenkeel"),
		}},
	})
}
