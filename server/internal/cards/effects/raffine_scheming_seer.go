package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Raffine, Scheming Seer — Legendary Creature — Sphinx Demon {W}{U}{B}, 1/4:
//
//	"Flying, ward {1}
//	 Whenever you attack, target attacking creature connives X, where X
//	 is the number of attacking creatures. (Draw X cards, then discard X
//	 cards. Put a +1/+1 counter on that creature for each nonland card
//	 discarded this way.)"
//
// "Whenever you attack" is one trigger per declaration (OncePerBatch, the
// Kiora of Salt and Sand shape), with the target chosen as the trigger goes
// on the stack. The target is any attacking creature, yours or not, as
// printed. X is every attacking creature, counted as the ability resolves
// (not when it was put on the stack), so an attacker removed in response
// lowers it. If the target has left combat or the battlefield, the ability
// does nothing (CR 608.2b). Connive itself is ConniveX in connive.go.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9307f7e9-bcd9-41f9-88f8-0fb9700a6b12",
		Name:            "Raffine, Scheming Seer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{1}"), "Raffine, Scheming Seer — ward {1}"),
			OncePerBatch(Targeting(
				On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclaredByYou(ev, source.Controller)
				}, "Raffine, Scheming Seer — target attacking creature connives X", raffineConnive),
				TargetCreature("target attacking creature", AttackingCreature()))),
		},
	})
}

func raffineConnive(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	x := len(AttackingCreatures(g))
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (ConniveX{Controller: item.Controller, Creature: t.ID, X: x}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
