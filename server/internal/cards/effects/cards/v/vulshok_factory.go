package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vulshok Factory — Artifact {2}{R}:
//
//	"{T}: Add {R}. Put a charge counter on this artifact.
//	 {2}{R}, {T}, Sacrifice this artifact: Create an X/X colorless Golem
//	 artifact creature token with haste, where X is the number of charge
//	 counters on Vulshok Factory. Activate only as a sorcery."
//
// Glittering Stockpile's mana ability with a Rider (the counter follows the
// mana, CR 605.3b). The sacrifice is a cost, so X is the Factory's last-known
// charge counters (CR 608.2h); with none the Golem is 0/0 and dies to
// state-based actions.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c5b0d98b-9743-488e-a3be-07043d6ad78a",
		Name:         "Vulshok Factory",
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Rider:    s82OneB12PutAChargeCounter,
			Label:    "{T}: Add {R}. Put a charge counter on this artifact",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:        "{2}{R}, {T}, Sacrifice this artifact: Create an X/X colorless Golem artifact creature token with haste, where X is the number of charge counters on Vulshok Factory. Activate only as a sorcery.",
			Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
			Cost:         effects.Plus(effects.ManaCost("{2}{R}"), effects.TapCost(), effects.SacrificeThis()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				x := 0
				if info, ok := ctx.SourcePermanent(); ok {
					x = info.Counters[game.CounterCharge]
				}
				tmpl := effects.S82OneB12Token("Golem", "Token Artifact Creature — Golem", x, x, nil, []string{"haste"})
				return g.CreateTokenForEffect(item.Controller, ctx.TextToken(tmpl), 1)
			},
		}},
	})
}

// s82OneB12PutAChargeCounter is "Put a charge counter on this artifact", the
// rider of Vulshok Factory's mana ability.
func s82OneB12PutAChargeCounter(g *game.Game, _, source uuid.UUID) error {
	_, err := g.AddCounterMustSettleNowForEffect(source, game.CounterCharge, 1)
	return err
}
