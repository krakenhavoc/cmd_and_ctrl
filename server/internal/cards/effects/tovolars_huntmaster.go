package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
// nothing if either creature has left (CR 701.12b).
//
// No simplification.
func init() {
	const oracle = "18563bc9-6090-4629-b304-89a67d93f635"
	wolves := Do(CreateToken{Template: TokenCard("2/2 green Wolf"), N: 2})
	Register(Spec{
		OracleID:        oracle,
		Name:            "Tovolar's Huntmaster",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"daybound"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Tovolar's Huntmaster — create two 2/2 Wolf tokens", wolves),
		},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Tovolar's Packleader",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"nightbound"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Tovolar's Packleader — create two 2/2 Wolf tokens", wolves),
		},
		Activated: []ActivatedAbility{{
			Label: "{2}{G}{G}: Another target Wolf or Werewolf you control fights target creature you don't control.",
			Cost:  ManaCost("{2}{G}{G}"),
			Targets: Clauses(
				Another(TargetCreature("another target Wolf or Werewolf you control", YouControl(),
					Or(OfCreatureType("Wolf"), OfCreatureType("Werewolf")))),
				TargetCreature("target creature you don't control", OpponentControls()),
			),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				fighter, ok := ctx.ClauseTarget(0)
				if !ok || fighter.Kind != game.TargetCard {
					return nil
				}
				victim, ok := ctx.ClauseTarget(1)
				if !ok || victim.Kind != game.TargetCard {
					return nil
				}
				return b10Fight(ctx, fighter.ID, victim.ID)
			},
		}},
	})
}
