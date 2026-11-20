package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Knight-Captain of Eos — Creature — Human Knight {4}{W}, 2/2 (Shards of
// Alara, #5588):
//
//	"When this creature enters, create two 1/1 white Soldier creature
//	 tokens.
//	 {W}, Sacrifice a Soldier: Prevent all combat damage that would be
//	 dealt this turn."
//
// Siege-Gang Commander's shape in white: the enters trigger makes the two
// Soldiers, and the sacrifice is a cost (CR 701.21a) of any Soldier its
// controller controls, by effective subtype, so a changeling pays it. The
// effect is Fog's prevention for the rest of the turn (CR 615.1).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "611f1714-a8bb-4ce4-a810-4d26fe64e358",
		Name:         "Knight-Captain of Eos",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Knight-Captain of Eos — create two 1/1 white Soldiers",
				Do(CreateToken{Template: TokenCard("1/1 white Soldier"), N: 2})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{W}, Sacrifice a Soldier: Prevent all combat damage that would be dealt this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent | game.AnswerSacOutlet},
			Cost:    Plus(ManaCost("{W}"), game.AbilityCost{SacrificeOther: sacrificeSpec("a Soldier", Subtype("Soldier"))}),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return PreventAllCombatDamageThisTurn{Label: "Knight-Captain of Eos: prevent combat damage"}.Apply(NewContext(g, item))
			},
		}},
	})
}
