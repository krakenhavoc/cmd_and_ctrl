package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Teysa, Envoy of Ghosts — Legendary Creature — Human Advisor {5}{W}{B}, 4/4
// (Commander 2021, S82 OLD b058, #5504):
//
//	"Vigilance, protection from creatures
//	 Whenever a creature deals combat damage to you, destroy that creature.
//	 Create a 1/1 white and black Spirit creature token with flying."
//
// One trigger per creature that dealt combat damage to its controller (any
// creature, whoever controls it; damage of 0 is none). The creature is pinned
// to the object it was when it dealt the damage (BuildPinnedToEventSource), so
// one that left or came back as a new object by resolution is not destroyed
// (CR 400.7); the Spirit is created either way. Protection from creatures
// stops damage to Teysa herself, not to her controller.
//
// No simplification.
func init() {
	const label = "Teysa, Envoy of Ghosts — destroy that creature and create a 1/1 white and black Spirit creature token with flying"
	effects.Register(effects.Spec{
		OracleID:        "98b283dc-96a6-45a3-8a3f-11458497f358",
		Name:            "Teysa, Envoy of Ghosts",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "protection from creatures"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventDealDamage},
			AppliesTo: effects.StrixhavenCreatureDealtCombatDamageToYou,
			Key:       label,
			Build:     effects.BuildPinnedToEventSource(label),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if ref := item.Params.Object; ref.ID != uuid.Nil {
					if c, ok := effects.BattlefieldCardForEffect(g, ref.ID); ok && c.ObjectEpoch == ref.Epoch {
						if err := (effects.DestroyTarget{Target: ref.ID}).Apply(ctx); err != nil {
							return err
						}
					}
				}
				return effects.CreateToken{Template: effects.TokenCard("1/1 white and black Spirit with flying"), N: 1}.Apply(ctx)
			},
		}},
	})
}
