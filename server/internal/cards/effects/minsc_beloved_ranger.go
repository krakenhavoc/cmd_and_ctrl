package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Minsc, Beloved Ranger — Legendary Creature — Human Ranger {R}{G}{W}, 3/3
// (Adventures in the Forgotten Realms, S82 batch 15, #5499):
//
//	"When Minsc enters, create Boo, a legendary 1/1 red Hamster creature
//	 token with trample and haste.
//	 {X}: Until end of turn, target creature you control has base power and
//	 toughness X/X and becomes a Giant in addition to its other types.
//	 Activate only as a sorcery."
//
// The {X} ability is Mirror Entity's shape on one target: a layer 7b base
// P/T set (CR 613.4b) and a creature-type addition (CR 205.1b, layer 4),
// both locked to the target as the ability resolves (CR 611.2c). Counters
// and other pumps still apply on top of the base. A target that has left
// or become illegal does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2cd10c6c-f946-457e-9175-92501322f7a5",
		Name:         "Minsc, Beloved Ranger",
		Completeness: CompletenessFull,
		XMatters:     true,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Minsc — create Boo, a legendary 1/1 red Hamster creature token with trample and haste",
				Do(CreateToken{Template: TokenCard(afrB15BooToken), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:        "{X}: Until end of turn, target creature you control has base power and toughness X/X and becomes a Giant in addition to its other types. Activate only as a sorcery.",
			Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0)),
			Cost:         ManaCost("{X}"),
			SorcerySpeed: true,
			Targets:      TargetCreature("target creature you control", YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				t, ok := firstLegalTarget(ctx)
				if !ok {
					return nil
				}
				x := ctx.X()
				return untilEndOfTurn(ctx, t.ID, nil,
					"Minsc, Beloved Ranger — base power and toughness X/X and a Giant until end of turn",
					game.SetBasePowerMod(x), game.SetBaseToughnessMod(x), game.AddSubtypesMod(ctx.TextSubtype("Giant")))
			},
		}},
	})
}
