package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sudden Substitution — Instant {2}{U}{U}:
//
//	"Split second
//	 Exchange control of target noncreature spell and target creature.
//	 Then the spell's controller may choose new targets for it."
//
// Two target clauses (ADR 0065) and ADR 0104's exchange of a spell and
// a permanent. The exchange is all or nothing (CR 701.12a): if either
// target is illegal at resolution, nothing is exchanged. If one player
// controls both, it does nothing (CR 701.12b).
//
// "Then the spell's controller may choose new targets for it" is a
// sentence of its own, so it is offered whether or not the exchange
// happened, to whoever controls the spell by then — the creature's
// former controller after an exchange, the caster of the spell
// otherwise — as long as the spell is still a legal target.
//
// Split second is the printed keyword (CR 702.61): while this is on
// the stack nobody can cast a spell or activate a non-mana ability.
func init() {
	Register(Spec{
		OracleID:        "e5ddacfb-e5a8-4885-a4c9-fa117f321293",
		Name:            "Sudden Substitution",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"split second"},
		Targets: Clauses(
			TargetSpell("target noncreature spell", Noncreature()),
			TargetCreature("target creature"),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			spell, spellOK := ctx.ClauseTarget(0)
			creature, creatureOK := ctx.ClauseTarget(1)
			if spellOK && creatureOK {
				if err := (ExchangeControlOfSpellAnd{Spell: spell.ID, Permanent: creature.ID,
					Label: "Sudden Substitution — exchange control"}).Apply(ctx); err != nil {
					return err
				}
			}
			if !spellOK {
				return nil
			}
			controller, ok := ctx.Game.SpellControllerForEffect(spell.ID)
			if !ok {
				return nil
			}
			return ChangeTargets{StackID: spell.ID, Chooser: controller,
				Policy: game.RetargetChooseNew, Optional: true}.Apply(ctx)
		},
	})
}
