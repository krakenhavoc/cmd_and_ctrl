package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tovolar's Huntmaster // Tovolar's Packleader — {4}{G}{G} Creature —
// Human Werewolf 6/6 // Creature — Werewolf 7/7 (#2586, ADR 0132):
//
//	Front: "When this creature enters, create two 2/2 green Wolf creature
//	        tokens.
//	        Daybound"
//	Back:  "Whenever this creature enters or attacks, create two 2/2
//	        green Wolf creature tokens.
//	        {2}{G}{G}: Another target Wolf or Werewolf you control fights
//	        target creature you don't control.
//	        Nightbound"
//
// The back face's "enters" matters: cast at night the permanent enters on
// Tovolar's Packleader (CR 702.145b), so it is that face's trigger that
// sees it enter. The fight is a two-clause activated ability, each clause
// carrying its own predicate (Bite Down's shape), and b10Fight does
// nothing if either creature has left (CR 701.14b).
//
// No simplification.
func init() {
	const oracle = "18563bc9-6090-4629-b304-89a67d93f635"
	wolves := effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 green Wolf"), N: 2})
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Tovolar's Huntmaster",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"daybound"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Tovolar's Huntmaster — create two 2/2 Wolf tokens", wolves),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Tovolar's Packleader",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"nightbound"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrAttacks("Tovolar's Packleader — create two 2/2 Wolf tokens", wolves),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{2}{G}{G}: Another target Wolf or Werewolf you control fights target creature you don't control.",
			Cost:  effects.ManaCost("{2}{G}{G}"),
			Targets: effects.Clauses(
				effects.Another(effects.TargetCreature("another target Wolf or Werewolf you control", effects.YouControl(),
					effects.Or(effects.OfCreatureType("Wolf"), effects.OfCreatureType("Werewolf")))),
				effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
			),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				fighter, ok := ctx.ClauseTarget(0)
				if !ok || fighter.Kind != game.TargetCard {
					return nil
				}
				victim, ok := ctx.ClauseTarget(1)
				if !ok || victim.Kind != game.TargetCard {
					return nil
				}
				return effects.B10Fight(ctx, fighter.ID, victim.ID)
			},
		}},
	})
}
