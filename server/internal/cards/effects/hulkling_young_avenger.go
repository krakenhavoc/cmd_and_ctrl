package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hulkling, Young Avenger — Legendary Creature — Kree Skrull Hero {2}{R}{R},
// 4/4 (Marvel Super Heroes Commander):
//
//	"Flying
//	 Whenever you cast a noncreature spell, Hulkling becomes a copy of up to
//	 one other target creature until end of turn, except his name is
//	 Hulkling, Young Avenger, he's 4/4, and he has flying and this ability."
//
// Unstable Shapeshifter's shape: a duration copy (BecomeCopy, CR 707.2) whose
// "except" clause is a GRANT of the trigger itself (CR 707.9a), so the ability
// survives the copy. The clause also sets the name, the 4/4 body and flying.
// "Up to one" target: with none, the trigger does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c92f20c4-838d-43a1-8a71-dae205789d94",
		Name:            "Hulkling, Young Avenger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Grants: []AbilityGrant{{
			Key:       hulklingGrant,
			Text:      hulklingText,
			Triggered: []game.TriggeredAbility{hulklingTrigger()},
		}},
		Triggered: []game.TriggeredAbility{hulklingTrigger()},
	})
}

const (
	hulklingGrant = "hulkling-young-avenger/this-ability"
	hulklingText  = "Whenever you cast a noncreature spell, Hulkling becomes a copy of up to one other target creature until end of turn, except his name is Hulkling, Young Avenger, he's 4/4, and he has flying and this ability."
)

// hulklingTrigger is the printed trigger and, byte for byte, the one the copy
// grants itself.
func hulklingTrigger() game.TriggeredAbility {
	return Targeting(WheneverYouCast(Noncreature(), "Hulkling — becomes a copy of up to one other target creature until end of turn", hulklingCopy),
		Another(TargetCreature("up to one other target creature").WithCount(0, 1)))
}

func hulklingCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 {
		return nil
	}
	return BecomeCopy{
		Targets: []uuid.UUID{ctx.Source()},
		Of:      targets[0].ID,
		Except: func(v *game.PrintedValues) {
			v.SetName("Hulkling, Young Avenger")
			v.Power, v.Toughness = 4, 4
			v.VariableToughness = false
			v.PrintedPTKnown = true
			v.AddKeyword("flying")
			v.GrantAbility(hulklingGrant)
		},
		Label: "Hulkling — becomes a copy until end of turn",
	}.Apply(ctx)
}
