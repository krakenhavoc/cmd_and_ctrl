package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Secret Identity — Instant {U} (Marvel's Spider-Man):
//
//	"Choose one —
//	 • Conceal — Until end of turn, target creature you control becomes a
//	   Citizen with base power and toughness 1/1 and gains hexproof.
//	 • Reveal — Until end of turn, target creature you control becomes a
//	   Hero with base power and toughness 3/4 and gains flying and
//	   vigilance."
//
// "Becomes a Citizen" names a creature type and no card type, so it sets the
// creature subtypes and keeps the card types (CR 205.1a): SetCreatureTypesMod
// (Reptil, Dinomorpher's shape). Base power and toughness are layer 7b, so
// counters and pumps still land on top. Each bullet carries its own target
// clause (CR 700.2c); the effect is a continuous effect until end of turn on
// the chosen creature as it resolves (CR 611.2a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e6637f9f-8293-43a5-a134-93799a016ecb",
		Name:         "Secret Identity",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeWithPurpose(ModeDoing("Conceal — Until end of turn, target creature you control becomes a Citizen with base power and toughness 1/1 and gains hexproof.",
				TargetCreature("target creature you control", YouControl()),
				secretIdentityBecomes("Citizen", 1, 1, "hexproof")),
				ForTargets(BenefitUntilEOT(0, 0, 0, "hexproof"))),
			ModeWithPurpose(ModeDoing("Reveal — Until end of turn, target creature you control becomes a Hero with base power and toughness 3/4 and gains flying and vigilance.",
				TargetCreature("target creature you control", YouControl()),
				secretIdentityBecomes("Hero", 3, 4, "flying", "vigilance")),
				ForTargets(BenefitUntilEOT(0, 0, 0, "flying", "vigilance"))),
		),
	})
}

func secretIdentityBecomes(subtype string, power, toughness int, keywords ...string) func(*game.StackItem, *Context, int) error {
	return func(_ *game.StackItem, ctx *Context, occ int) error {
		t, ok := ModeTarget(ctx, occ)
		if !ok || t.Kind != game.TargetCard {
			return nil
		}
		return ScopedEffectFor{
			Target: t.ID,
			Mods: []game.Mod{
				game.SetCreatureTypesMod(ctx.TextSubtypes(subtype)...),
				game.SetBasePowerMod(power),
				game.SetBaseToughnessMod(toughness),
				game.AddKeywordsMod(keywords...),
			},
			Duration: DurationUntilEndOfTurn(ctx),
			Label:    "Secret Identity — becomes a " + subtype + " until end of turn",
		}.Apply(ctx)
	}
}
