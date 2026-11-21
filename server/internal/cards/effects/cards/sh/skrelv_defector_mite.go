package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skrelv, Defector Mite — toxic 1, can't block, and a Phyrexian-white
// tap ability granting another creature toxic 1, hexproof from the
// chosen color and evasion against that color until end of turn.
// The target is announced first; color is chosen during resolution.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "20053847-6623-493c-8cdb-a69cda3b1577",
		Name:            "Skrelv, Defector Mite",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"toxic 1"},
		Static:          []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
		Activated: []effects.ActivatedAbility{{
			Label: "{W/P}, {T}: Choose a color. Another target creature you control gains toxic 1 and hexproof from that color until end of turn. It can't be blocked by creatures of that color this turn.",
			// #4734 slice 7: hexproof from a color (protect), and toxic and
			// evasion against that color (combat_grant): the union (ADR 0142
			// sweep ruling 4).
			Purpose: game.Purpose{Answers: game.AnswerProtect | game.AnswerCombatGrant, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0))},
			Cost:    effects.Plus(effects.ManaCost("{W/P}"), effects.TapCost()),
			Targets: effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				targets := ctx.LegalTargets()
				if len(targets) == 0 {
					return nil
				}
				effects.ChooseColorTail(game.ColorForProtection, g, item.Controller, item.SourceCardID,
					"Skrelv, Defector Mite — choose a color", effects.WithCarry(ctx.ItemTail(skrelvColorChosen), targets[0].ID))
				return nil
			},
		}},
	})
}

// skrelvColorChosen grants the target (the tail's one carried ID) toxic 1,
// hexproof from the colour (its Word) and evasion against it until end of
// turn. Run against the asking ability. An on-disk key.
var skrelvColorChosen = effects.CardTail("then/skrelv-defector-mite-color-chosen", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	color := o.Tail.Word
	return effects.ScopedEffectFor{
		Target: o.Tail.Carry[0],
		Mods: []game.Mod{
			game.AddKeywordsMod("toxic 1", game.HexproofFromColor(color)),
			game.CantBeBlockedByColorMod(color),
		},
		Duration: effects.DurationUntilEndOfTurn(ctx),
		Label:    "Skrelv, Defector Mite — toxic 1, hexproof and evasion from " + game.ColorName(color),
	}.Apply(ctx)
})
