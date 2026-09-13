package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1ff25f67-36a7-4cfa-a2b1-2135b5b6fb67",
		Name:         "Scavenging Ooze",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{G}: Exile target card from a graveyard. If it was a creature card, put a +1/+1 counter on this creature and you gain 1 life.",
			Cost:    ManaCost("{G}"),
			Targets: TargetCardInGraveyard("target card from a graveyard"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					c, ok := g.LookupCardForEffect(t.ID)
					if !ok {
						continue
					}
					wasCreature := c.IsCreature()
					if err := (ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
					if !wasCreature {
						continue
					}
					if z := g.FindCardZoneForEffect(item.SourceCardID); z != nil && z.Kind == game.ZoneBattlefield {
						if err := (AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: 1}).Apply(ctx); err != nil {
							return err
						}
					}
					if err := (GainLife{Player: item.Controller, Amount: 1}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
