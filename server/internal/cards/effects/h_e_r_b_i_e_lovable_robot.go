package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// H.E.R.B.I.E., Lovable Robot — Legendary Artifact Creature — Robot Scout
// {2}, 1/1 (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 At the beginning of combat on your turn, if you've cast a noncreature
//	 spell this turn, surveil 1.
//	 {T}: Add {C}.
//	 {1}, {T}: Add one mana of any color."
//
// The trigger is Franklin Richards' (an intervening if, CR 603.4, checked as
// combat begins; a spell cast this turn cannot stop having been cast, so the
// resolution-time recheck could never change the answer and the surveil is a
// plain Do step). The two mana abilities are Conduit
// Pylons': a plain {C} and a {1} filter producing the five-colour pipe.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "064a2f4a-ff3f-4622-8727-24b22965416d",
		Name:            "H.E.R.B.I.E., Lovable Robot",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		ManaAbilities: []ManaAbility{
			painlessColorless(),
			{
				Cost:     ManaAbilityCost{Tap: true, Mana: "{1}"},
				Produced: "{W|U|B|R|G}",
				Label:    "{1}, {T}: Add one mana of any color",
			},
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventStepBegan, AllOf(StepBegan(game.StepBeginCombat, true), franklinCastNoncreature),
				"H.E.R.B.I.E., Lovable Robot — surveil 1",
				Do(Surveil{N: 1})),
		},
	})
}
