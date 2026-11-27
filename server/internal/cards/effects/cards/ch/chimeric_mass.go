package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chimeric Mass — Artifact {X} (Scars of Mirrodin, S82 OLD batch 470, #5577):
//
//	"This artifact enters with X charge counters on it.
//	 {1}: Until end of turn, this artifact becomes a Construct artifact
//	 creature with "This creature's power and toughness are each equal to the
//	 number of charge counters on it.""
//
// Walking Ballista's X counters as it enters, and Myth Realized's animation:
// until end of turn it is a Construct creature (it stays an artifact) with a
// granted power-and-toughness static in layer 7b that counts the charge
// counters on it as they are now. With no counters it is 0/0 and dies to the
// state-based actions (CR 704.5f).
//
// No simplification.
func init() {
	const grantKey = "chimeric-mass/power-toughness-charge-counters"
	const text = "This creature's power and toughness are each equal to the number of charge counters on it."
	effects.Register(effects.Spec{
		OracleID:                   "98066238-28f8-41e3-903a-d367927f23ca",
		Name:                       "Chimeric Mass",
		Completeness:               effects.CompletenessFull,
		XMatters:                   true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.XCounters(game.CounterCharge)},
		Grants: []effects.AbilityGrant{{
			Key: grantKey,
			Static: []game.StaticAbility{{
				Label:    text,
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7B_Set,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return source != nil && target.InstanceID == source.InstanceID
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
					n := source.Counters[game.CounterCharge]
					c.Power, c.Toughness = n, n
				},
			}},
			Text: text,
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}: Until end of turn, this artifact becomes a Construct artifact creature with \"" + text + "\"",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{1}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				self := item.SourceCardID
				if !effects.OnBattlefield(g, self) || effects.SourceIsNewObject(g, item) {
					return nil
				}
				return effects.GrantAbilitiesFor{
					Target: self,
					Keys:   []string{grantKey},
					Also: []game.Mod{
						game.AddTypesMod("Creature"),
						game.AddSubtypesMod(g.TextReaderForEffect(item).TextSubtypes("Construct")...),
					},
					Label: "Chimeric Mass — becomes a Construct artifact creature until end of turn",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
