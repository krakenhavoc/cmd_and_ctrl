package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dropkick Bomber — Creature — Goblin Warrior {2}{R}, 2/3 (Foundations,
// #4465):
//
//	"Other Goblins you control get +1/+1.
//	 {R}: Until end of turn, another target Goblin you control gains
//	 flying and "When this creature deals combat damage, sacrifice it.""
//
// The anthem is a layer 7c tribal lord over the effective view
// (TribeFilter), so a changeling Goblin gets it. The activation is a duration
// grant (ADR 0093): the flying and the granted trigger are one effect at
// one timestamp (GrantAbilitiesFor.Also), and the trigger is the
// Goblin's own: once per batch of combat damage (a double striker that
// hits in both steps has already been sacrificed after the first), and it
// stays after the Bomber leaves, as the ruling says.
//
// No simplification.
const dropkickBomberSacrifice = "dropkick-bomber/sacrifice"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "c394d4b6-cfb8-4639-90cb-a322a25c0a2f",
		Name:         "Dropkick Bomber",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Goblin"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Grants: []effects.AbilityGrant{{
			Key: dropkickBomberSacrifice,
			Triggered: []game.TriggeredAbility{
				effects.OncePerBatch(effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Kind == game.EventDealDamage && ev.Combat && ev.Amount > 0 && ev.Source == source.InstanceID
				}, "Dropkick Bomber — sacrifice this creature", func(g *game.Game, item *game.StackItem) error {
					return effects.SacrificePermanent{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
				})),
			},
			Text: "When this creature deals combat damage, sacrifice it.",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: Until end of turn, another target Goblin you control gains flying and \"When this creature deals combat damage, sacrifice it.\"",
			Cost:    effects.ManaCost("{R}"),
			Targets: effects.Another(effects.TargetCreature("another target Goblin you control", effects.YouControl(), effects.OfSubtype("Goblin"))),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				target, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.GrantAbilitiesFor{
					Target: target,
					Keys:   []string{dropkickBomberSacrifice},
					Also:   []game.Mod{game.AddKeywordsMod("flying")},
					Label:  "Dropkick Bomber — flying, and it is sacrificed when it deals combat damage",
				}.Apply(ctx)
			},
		}},
	})
}
