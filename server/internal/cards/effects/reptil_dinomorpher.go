package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reptil, Dinomorpher — Legendary Creature — Human Hero {G}, 1/2
// (Marvel Super Heroes, #4455):
//
//	"Brontosaurus — {3}: Until end of turn, Reptil becomes a Dinosaur Hero
//	 with base power and toughness 3/5 and gains reach and vigilance.
//	 Tyrannosaurus Rex — {6}: Until end of turn, Reptil becomes a Dinosaur
//	 Hero with base power and toughness 6/6 and gains trample."
//
// Two activated rows over one body: a layer-4 type change that replaces his
// creature types (SetCreatureTypesMod, CR 205.1a: not "in addition"), a
// layer-7b base P/T, and a layer-6 keyword grant, all until end of turn
// (CR 611.2a). The later activation's timestamp wins where the two overlap
// (CR 613.7), so paying {6} after {3} leaves him 6/6 with every keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c6284597-1983-4a21-a461-724135220e0d",
		Name:         "Reptil, Dinomorpher",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			mshB7ReptilForm("Brontosaurus — {3}: Until end of turn, Reptil becomes a Dinosaur Hero with base power and toughness 3/5 and gains reach and vigilance.",
				"{3}", 3, 5, "reach", "vigilance"),
			mshB7ReptilForm("Tyrannosaurus Rex — {6}: Until end of turn, Reptil becomes a Dinosaur Hero with base power and toughness 6/6 and gains trample.",
				"{6}", 6, 6, "trample"),
		},
	})
}

func mshB7ReptilForm(label, cost string, power, toughness int, keywords ...string) ActivatedAbility {
	return ActivatedAbility{
		Label:   label,
		Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerAnimate},
		Cost:    ManaCost(cost),
		Effect: func(g *game.Game, item *game.StackItem) error {
			if !sourceIsStillThisPermanent(g, item) {
				return nil
			}
			ctx := NewContext(g, item)
			return ScopedEffectFor{
				Target: item.SourceCardID,
				Mods: []game.Mod{
					game.SetCreatureTypesMod(ctx.TextSubtypes("Dinosaur", "Hero")...),
					game.SetBasePowerMod(power),
					game.SetBaseToughnessMod(toughness),
					game.AddKeywordsMod(keywords...),
				},
				Duration: DurationUntilEndOfTurn(ctx),
				Label:    "Reptil, Dinomorpher — a Dinosaur Hero until end of turn",
			}.Apply(ctx)
		},
	}
}
