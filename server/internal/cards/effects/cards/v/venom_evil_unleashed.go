package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Venom, Evil Unleashed — Legendary Creature — Symbiote Villain {4}{B},
// 4/5 (Marvel Super Heroes Commander, #4455):
//
//	"Deathtouch
//	 {2}{B}, Exile this card from your graveyard: Put two +1/+1 counters on
//	 target creature. It gains deathtouch until end of turn. Activate only
//	 as a sorcery."
//
// Scavenge's graveyard ability (the card exiles itself as a cost, so it is
// gone before the ability is even on the stack, and the ability works
// whether or not Venom could still be returned) with fixed counters and a
// deathtouch rider. Both halves act on the one target; a target that left
// in response gets neither (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6731d861-cc52-4a2e-8c33-154681fd50e4",
		Name:            "Venom, Evil Unleashed",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Activated: []effects.ActivatedAbility{{
			Label: "{2}{B}, Exile this card from your graveyard: Put two +1/+1 counters on target creature. " +
				"It gains deathtouch until end of turn. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{2}{B}"), effects.ExileThis()),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Targets:      effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"deathtouch"}, Label: "Venom — deathtouch"}.Apply(ctx)
			},
		}},
	})
}
