package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jaxis, the Troublemaker — Legendary Creature — Human Warrior {3}{R}, 2/3:
//
//	"{R}, {T}, Discard a card: Create a token that's a copy of another
//	 target creature you control. It gains haste and "When this token
//	 dies, draw a card." Sacrifice it at the beginning of the next end
//	 step. Activate only as a sorcery.
//	 Blitz {1}{R}"
//
// Blitz (CR 702.152, #3772), and the same three riders printed on an
// activated ability: the token gets exactly what a blitzed creature
// gets, through game.ApplyBlitzRidersForEffect — haste and the shared
// draw bundle for as long as it is on the battlefield, and a delayed
// sacrifice at the next end step. The token's ID is read back off
// EventTokenCreated, Reflection of Kiki-Jiki's way.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "764121bf-535d-40d0-ab9d-65486da31734",
		Name:             "Jaxis, the Troublemaker",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Blitz("{1}{R}")},
		Activated: []ActivatedAbility{{
			Label:        "{R}, {T}, Discard a card: Create a token that's a copy of another target creature you control. It gains haste and \"When this token dies, draw a card.\" Sacrifice it at the beginning of the next end step. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{R}"), TapCost(), DiscardACard()),
			Targets:      Another(TargetCreature("another target creature you control", YouControl())),
			SorcerySpeed: true,
			Effect:       jaxisCopy,
		}},
	})
}

// jaxisCopy makes the copy and gives it the blitz riders.
func jaxisCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		cursor := b25LastEventSeq(g)
		if err := (CreateTokenCopy{Controller: item.Controller, Copy: t.ID, N: 1}).Apply(ctx); err != nil {
			return err
		}
		for _, tok := range b27TokensCreatedByAfter(g, item.Controller, cursor) {
			g.ApplyBlitzRidersForEffect(tok, item.Controller, "Jaxis, the Troublemaker — sacrifice the token")
		}
		return nil
	}
	return nil
}
