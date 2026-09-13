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
// Sandbox simplification, declared: "spell OR ABILITY" is spells
// only. An activated or triggered ability on the stack is a
// StackMeta item with no card in the stack zone, and the target
// clauses enumerate cards, so an ability that targets you cannot be
// chosen. Weaker than printed, never stronger.
func init() {
	Register(Spec{
		OracleID:        "59c7496b-19a2-478b-b28f-6f153c2458ae",
		Name:            "Siren Stormtamer",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"Only spells can be countered — an ability that targets you or your creature can't be picked."},
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{U}, Sacrifice this creature: Counter target spell that targets you or a creature you control.",
			Cost:    Plus(ManaCost("{U}"), SacrificeThis()),
			Targets: TargetSpell("target spell that targets you or a creature you control", b10SpellTargetsYouOrACreatureYouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				return CounterTarget{StackID: item.Targets[0].ID}.Apply(NewContext(g, item))
			},
		}},
	})
}
