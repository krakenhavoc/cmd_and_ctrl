package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mister Gutsy — Artifact Creature — Robot Soldier {2}, 1/1 (Fallout,
// #5948):
//
//	"Whenever you cast an Aura or Equipment spell, put a +1/+1 counter
//	 on this creature.
//	 When this creature dies, create X Junk tokens, where X is the
//	 number of +1/+1 counters on it. (They're artifacts with "{T},
//	 Sacrifice this token: Exile the top card of your library. You may
//	 play that card this turn. Activate only as a sorcery.")"
//
// The cast trigger is WheneverYouCast over an Aura or an Equipment (the
// spell's subtypes). X is read from the departure record (the counters
// it had as it last existed, CR 608.2h), Tangled Colony's shape: a
// Gutsy that died with no +1/+1 counters makes nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "09763d9a-056f-49db-a72b-1c7325075789",
		Name:         "Mister Gutsy",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Or(effects.HasSubtype("Aura"), effects.HasSubtype("Equipment")),
				"Mister Gutsy — put a +1/+1 counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			effects.WhenThisDies("Mister Gutsy — create X Junk tokens, where X is the number of +1/+1 counters on it",
				func(g *game.Game, item *game.StackItem) error {
					info, ok := g.LastKnownPermanentForEffect(item.SourceCardID)
					if !ok || info.Counters[game.CounterPlusOne] <= 0 {
						return nil
					}
					return effects.CreateToken{Template: effects.JunkToken(), N: info.Counters[game.CounterPlusOne]}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
