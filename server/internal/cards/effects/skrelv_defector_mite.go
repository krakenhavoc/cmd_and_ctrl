package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skrelv, Defector Mite — toxic 1, can't block, and a Phyrexian-white
// tap ability granting another creature toxic 1, hexproof from the
// chosen color and evasion against that color until end of turn.
// The target is announced first; color is chosen during resolution.
func init() {
	Register(Spec{
		OracleID:        "20053847-6623-493c-8cdb-a69cda3b1577",
		Name:            "Skrelv, Defector Mite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"toxic 1"},
		Static:          []game.StaticAbility{RestrictSelf(game.CantBlock)},
		Activated: []ActivatedAbility{{
			Label:   "{W/P}, {T}: Choose a color. Another target creature you control gains toxic 1 and hexproof from that color until end of turn. It can't be blocked by creatures of that color this turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0)),
			Cost:    Plus(ManaCost("{W/P}"), TapCost()),
			Targets: Another(TargetCreature("another target creature you control", YouControl())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				targets := ctx.LegalTargets()
				if len(targets) == 0 {
					return nil
				}
				ChooseColorTail(game.ColorForProtection, g, item.Controller, item.SourceCardID,
					"Skrelv, Defector Mite — choose a color", withCarry(ctx.ItemTail(skrelvColorChosen), targets[0].ID))
				return nil
			},
		}},
	})
}

// skrelvColorChosen grants the target (the tail's one carried ID) toxic 1,
// hexproof from the colour (its Word) and evasion against it until end of
// turn. Run against the asking ability. An on-disk key.
var skrelvColorChosen = CardTail("then/skrelv-defector-mite-color-chosen", func(ctx *Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	color := o.Tail.Word
	return ScopedEffectFor{
		Target: o.Tail.Carry[0],
		Mods: []game.Mod{
			game.AddKeywordsMod("toxic 1", game.HexproofFromColor(color)),
			game.CantBeBlockedByColorMod(color),
		},
		Duration: DurationUntilEndOfTurn(ctx),
		Label:    "Skrelv, Defector Mite — toxic 1, hexproof and evasion from " + game.ColorName(color),
	}.Apply(ctx)
})
