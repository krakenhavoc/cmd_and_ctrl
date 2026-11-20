package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scavenging Ooze — Creature — Ooze {1}{G}, 2/2 (EDHREC rank 1597):
//
//	"{G}: Exile target card from a graveyard. If it was a creature
//	 card, put a +1/+1 counter on this creature and you gain 1 life."
//
// The graveyard hate that grows. Any graveyard, any card, as
// printed; the "if it was a creature card" reads the card's type
// BEFORE the exile, and the counter and the life follow only then.
// No tap in the cost, so a fresh Ooze can eat the turn it lands and
// can eat as many times as there is green mana. The Ooze itself has
// to still be on the battlefield for the counter to land — an Ooze
// that died in response exiles the card and grows nothing.
//
// #911, ADR 0013 §5t: the clause also waits for the exile and is
// gated on it. An Ooze whose meal was kept in its graveyard by the
// CR 614 window — or whose owner sent a commander card to the command
// zone instead — ate nothing, so it grows nothing and gains nothing.
// ExileThenIfItWas holds both halves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1ff25f67-36a7-4cfa-a2b1-2135b5b6fb67",
		Name:         "Scavenging Ooze",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}: Exile target card from a graveyard. If it was a creature card, put a +1/+1 counter on this creature and you gain 1 life.",
			Cost:    effects.ManaCost("{G}"),
			Targets: effects.TargetCardInGraveyard("target card from a graveyard"),
			// #4559: what the exile takes away and the counter it grows.
			// The life it gains is left unpriced.
			Purpose: effects.ForTargets(effects.ExilesFromGraveyard(0, game.ExileGrowsCreature)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if err := (effects.ExileThenIfItWas{
						Target: t.ID,
						Was:    effects.WasCreatureCard,
						// The life gain waits for a prompt the counter
						// asks (a CR 616 order; #4162, CR 608.2c).
						Keyed: scavengingOozeIfCreature,
					}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}

// scavengingOozeIfCreature is the exile's keyed "if it was a creature
// card" (#4253 slice 10a).
var scavengingOozeIfCreature = effects.IfItWasTail("if-it-was/scavenging-ooze-grow", func(ctx *effects.Context) error {
	return scavengingOozeGrow.Run(ctx)
}, nil)

// scavengingOozeGrow is "put a +1/+1 counter on this creature and you
// gain 1 life".
var scavengingOozeGrow = effects.RegisterThenSteps("then-steps/scavenging-ooze-grow",
	effects.Instruction(func(ctx *effects.Context) error {
		// Re-read here rather than before the exile: the Ooze has to
		// be on the battlefield when the counter lands, and with a
		// CR 903.9 prompt in between that is not the same moment as
		// the cast.
		self := ctx.Item.SourceCardID
		if z := ctx.Game.FindCardZoneForEffect(self); z == nil || z.Kind != game.ZoneBattlefield {
			return nil
		}
		return effects.AddCounter{Target: self, Kind: "+1/+1", N: 1}.Apply(ctx)
	}),
	effects.GainLife{Amount: 1},
)
