package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crested Sunmare — Creature — Horse {3}{W}{W}, 5/5 (EDHREC rank
// 4278):
//
//	"Other Horses you control have indestructible.
//	 At the beginning of each end step, if you gained life this turn,
//	 create a 5/5 white Horse creature token."
//
// A five-mana 5/5 that makes another 5/5 every end step and makes them
// all unkillable — as long as you gained a single point of life that
// turn, which any lifelink creature, any Soul Warden or any gain land
// supplies. The reason it is a Commander card rather than a Standard
// rare is the word EACH: at a four-player table with a steady lifegain
// trickle it is four Horses a turn cycle, not one.
//
// The two clauses are deliberately lopsided and the card would be
// broken if they were not: the Sunmare grants indestructible to OTHER
// Horses and not to itself, so the horse army survives a wrath and the
// engine that built it does not.
//
// "If you gained life this turn" is an INTERVENING-IF (CR 603.4),
// checked when the end step begins and again on resolution. The tally
// counts positive life changes only, so gaining 2 and then losing 2
// still satisfies it — the card asks what you gained, not where your
// life total ended up.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "054fedbb-062e-491e-9a94-594fda33094f",
		Name:         "Crested Sunmare",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{"Horse"}, Others: true, YoursOnly: true}, "indestructible"),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B33EndStepAndYouGainedLifeThisTurn(ev, source, g)
			}, "Crested Sunmare — create a 5/5 white Horse",
				// CR 603.4: re-checked on resolution.
				effects.Do(effects.If{
					Cond: effects.CondLifeGained{},
					Then: []effects.Applier{effects.CreateToken{Template: effects.TokenCard("5/5 white Horse"), N: 1}},
				})),
		},
	})
}
