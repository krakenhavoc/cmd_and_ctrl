package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vampiric Dragon — Creature — Vampire Dragon {6}{B}{R}, 5/5 (Innistrad:
// Crimson Vow Commander, S82):
//
//	"Flying
//	 Whenever a creature dealt damage by this creature this turn dies, put a
//	 +1/+1 counter on this creature.
//	 {1}{R}: This creature deals 1 damage to target creature."
//
// The dies trigger is Zurgo Helmsmasher's: the turn's record of damaged
// objects (dealtDamageByThisTurn, #1946) read for the creature that died,
// each side a new object once it re-enters (CR 400.7). The Dragon must
// still be on the battlefield for its counter. Its own death is not
// counted. The ping is a plain activated ability: it deals the damage as
// the Dragon, so the victim lands in that record.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e8941499-2b31-417b-b2d2-10a144826703",
		Name:            "Vampiric Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, died := diedCreature(ev, g)
				return died && ev.CardID != source.InstanceID && dealtDamageByThisTurn(g, source.InstanceID, ev.CardID)
			}, "Vampiric Dragon — put a +1/+1 counter on it", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}: This creature deals 1 damage to target creature.",
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Cost:    ManaCost("{1}{R}"),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				target, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return DealDamage{Source: item.SourceCardID, Target: target, Amount: 1}.Apply(ctx)
			},
		}},
	})
}
