package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hedge Whisperer — Creature — Elf Druid Detective {G}, 0/3 (Murders at
// Karlov Manor, #5193):
//
//	"You may choose not to untap this creature during your untap step.
//	 {3}{G}, {T}, Collect evidence 4: Target land you control becomes a 5/5
//	 green Plant Boar creature with haste for as long as this creature
//	 remains tapped. It's still a land. Activate only as a sorcery. (To
//	 collect evidence 4, exile cards with total mana value 4 or greater
//	 from your graveyard.)"
//
// Endoskeleton's shape: the untap opt-out, and a continuous effect from a
// resolving ability timed by "for as long as this creature remains tapped"
// (CR 611.2b, DurationWhileThisRemainsTapped). The land gains the creature
// type and the Plant Boar subtypes at layer 4, becomes green at layer 5
// (its only colour), gains haste at layer 6 and has base power and
// toughness 5/5 at layer 7b; it keeps its land types and abilities. The
// cost's collect evidence is CollectEvidenceN. If the Whisperer is already
// untapped when the ability resolves, nothing happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "923b2a1b-d7b6-439d-9681-c7ba88165c31",
		Name:         "Hedge Whisperer",
		Completeness: CompletenessFull,
		UntapOptOuts: []game.UntapOptOut{mayChooseNotToUntapSelf("Hedge Whisperer — you may choose not to untap this creature")},
		Activated: []ActivatedAbility{{
			Label: "{3}{G}, {T}, Collect evidence 4: Target land you control becomes a 5/5 green Plant Boar creature with haste " +
				"for as long as this creature remains tapped. It's still a land. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{3}{G}"), TapCost(), CollectEvidenceN(4)),
			Targets:      TargetPermanent("target land you control", Land(), YouControl()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				target := FirstLegalBattlefieldTarget(ctx)
				if target == uuid.Nil {
					return nil
				}
				d, ok := DurationWhileThisRemainsTapped(ctx)
				if !ok {
					return nil
				}
				return ScopedEffectFor{
					Target: target,
					Mods: []game.Mod{
						game.AddTypesMod("Creature"),
						game.AddSubtypesMod(ctx.TextSubtypes("Plant", "Boar")...),
						game.SetColorsMod(ctx.TextColor("G")),
						game.AddKeywordsMod("haste"),
						game.SetBasePowerMod(5),
						game.SetBaseToughnessMod(5),
					},
					Duration: d,
					Label:    "Hedge Whisperer — a 5/5 green Plant Boar creature with haste while this creature remains tapped",
				}.Apply(ctx)
			},
		}},
	})
}
