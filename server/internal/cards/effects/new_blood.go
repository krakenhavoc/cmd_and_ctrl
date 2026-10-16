package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// New Blood — Sorcery {2}{B}{B} (#3868):
//
//	"As an additional cost to cast this spell, tap an untapped Vampire
//	 you control.
//	 Gain control of target creature. Change the text of that creature by
//	 replacing all instances of one creature type with Vampire."
//
// The additional cost is TapUntappedCost (AdditionalCost.TapOthers, the
// activated abilities' TapOthersCost on a cast's mandatory slot): the
// caster names the Vampire on cast_spell as teamwork_ids, it is tapped
// with the spell on the stack, and with no untapped Vampire the spell
// cannot be cast (CR 118.3). Tapping is not the {T} symbol, so a Vampire
// that entered this turn may pay (CR 302.6), and paying does not target,
// so a hexproof one may too.
//
// The control change has no duration (CR 611.2a's "indefinitely"). The
// text change is CR 612.2's creature-type class (ADR 0160's 2026-10-15
// amendment): the caster chooses the replaced type as the spell resolves,
// and the new word is fixed, so Vampire itself is not offered. It reaches
// the type line, a target clause that names a type, and a lord's printed
// tribes, and lasts indefinitely like Artificial Evolution's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "021d654e-b731-40d0-a18e-eaff43f9fdb9",
		Name:           "New Blood",
		Completeness:   CompletenessFull,
		AdditionalCost: TapUntappedCost(1, "an untapped Vampire you control", Creature(), Subtype("Vampire")),
		Targets:        TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (GainControl{
					Target:   t.ID,
					Duration: game.IndefiniteDuration(),
					Label:    "New Blood — gain control",
				}).Apply(ctx); err != nil {
					return err
				}
				return ChangeCreatureTypeWord{Target: t.ID, To: "Vampire"}.Apply(ctx)
			}
			return nil
		},
	})
}
