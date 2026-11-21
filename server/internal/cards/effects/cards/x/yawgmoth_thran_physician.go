package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Yawgmoth, Thran Physician — Legendary Creature — Human Cleric
// {2}{B}{B}, 2/4 (issue #1117):
//
//	"Protection from Humans
//	 Pay 1 life, Sacrifice another creature: Put a -1/-1 counter on
//	 up to one target creature and draw a card.
//	 {B}{B}, Discard a card: Proliferate."
//
// Every printed clause is expressible, so this ships
// CompletenessFull — the sacrifice-outlet aristocrats' commander of
// choice, in full.
//
//   - Protection from Humans is a SUBTYPE quality (CR 702.16b-f,
//     ADR 0072), the same grammar Baneslayer Angel's "protection from
//     Demons and from Dragons" uses. A Human source cannot target,
//     block, deal damage to or enchant/equip Yawgmoth.
//   - The first ability's cost is a life payment plus a sacrifice of
//     ANOTHER creature — excluded by name, the Erebos, Bleak-Hearted
//     shape (#350: a legendary creature never wrongly refuses a
//     second copy this way). Both components are paid at announce,
//     so the draw happens whether or not a legal target remains for
//     the -1/-1 counter; "up to one" is a Min-0 target clause, and a
//     target that left in response is simply skipped (CR 608.2b).
//   - The second ability is an ordinary mana-plus-discard cost
//     (#660) with Proliferate (CR 701.34) as its effect.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a1e232c0-dc38-47be-a5a0-f68bc1d86a29",
		Name:            "Yawgmoth, Thran Physician",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from Humans"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "Pay 1 life, Sacrifice another creature: Put a -1/-1 counter on up to one target creature and draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerSacOutlet | game.AnswerRemove},
				Cost:    effects.Plus(effects.PayLife(1), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
				Targets: effects.TargetCreature("up to one target creature").WithCount(0, 1),
				Effect: effects.Do(
					// The draw waits for a prompt the counter asks (CR 608.2c, #4155).
					effects.Instruction(func(ctx *effects.Context) error {
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return effects.AddCounter{Target: id, Kind: game.CounterMinusOne, N: 1}.Apply(ctx)
					}),
					effects.DrawCards{N: 1},
				),
			},
			{
				Label:   "{B}{B}, Discard a card: Proliferate.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.Plus(effects.ManaCost("{B}{B}"), effects.DiscardACard()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.Proliferate{}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
