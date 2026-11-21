package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ashad, the Lone Cyberman — Legendary Artifact Creature — Cyberman
// {1}{U}{B}{R}, 3/3 (Doctor Who, #5948):
//
//	"The first nonlegendary artifact spell you cast each turn has casualty 2.
//	 (As you cast it, you may sacrifice a creature with power 2 or greater.
//	 When you do, copy it. A copy of an artifact spell becomes a token.)
//	 Whenever you sacrifice another creature, put a +1/+1 counter on Ashad."
//
// The first is Anhelo's granted casualty with a different "first" (the helper
// reads this turn's casts for an earlier nonlegendary artifact spell); the
// copy of an artifact spell becoming a token is the engine's. The second is
// Ashling's sacrifice read, for any creature other than Ashad, token or not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f3902dfd-6d09-425f-8446-a68cae2eea50",
		Name:         "Ashad, the Lone Cyberman",
		Completeness: CompletenessFull,
		GrantedOptionalCosts: []game.GrantedOptionalCost{
			SpellsYouCastHave("The first nonlegendary artifact spell you cast each turn has casualty 2.",
				gapB84FirstNonlegendaryArtifactSpell, Casualty(2)),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventSacrifice, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventSacrifice || ev.CardID == uuid.Nil || ev.CardID == source.InstanceID || ev.Actor != source.Controller {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature()
			}, "Ashad — put a +1/+1 counter on Ashad", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
