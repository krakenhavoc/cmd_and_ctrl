package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Siren Stormtamer — Creature — Siren Pirate Wizard {U}, 1/1 (EDHREC
// rank 1186):
//
//	"Flying
//	 {U}, Sacrifice this creature: Counter target spell or ability that
//	 targets you or a creature you control."
//
// The one-mana bodyguard. Flying rides PrintedKeywords; the ability
// is a mana-plus-sacrifice cost with a stack target clause whose
// predicate reads the candidate's announced targets: legal when any
// slot is the Stormtamer's controller or a creature they control ON
// the battlefield. The zone check is what makes the printed ruling
// fall out on its own — sacrificing the Stormtamer to counter a
// spell aimed only at the Stormtamer leaves that spell targeting a
// creature in a graveyard, the clause fails the CR 608.2b re-check,
// and the ability fizzles.
//
// "Spell OR ability" is TargetSpellOrAbility (#1211): an activated or
// triggered ability on the stack is a legal pick when it targets you
// or a creature you control, and a countered ability simply goes away.
func init() {
	Register(Spec{
		OracleID:        "59c7496b-19a2-478b-b28f-6f153c2458ae",
		Name:            "Siren Stormtamer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{U}, Sacrifice this creature: Counter target spell or ability that targets you or a creature you control.",
			Cost:    Plus(ManaCost("{U}"), SacrificeThis()),
			Targets: TargetSpellOrAbility("target spell or ability that targets you or a creature you control", b10ItemTargetsYouOrACreatureYouControl),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				return CounterTarget{StackID: item.Targets[0].ID}.Apply(NewContext(g, item))
			},
		}},
	})
}
