package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Burden of Proof — Enchantment — Aura {1}{U} (Murders at Karlov Manor,
// #4469):
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature gets +2/+2 as long as it's a Detective you
//	 control. Otherwise, it has base power and toughness 1/1 and can't
//	 block Detectives."
//
// Two halves of one condition, read live on every pass: the enchanted
// creature is a Detective and the Aura's controller controls it. True,
// a layer-7c +2/+2 (CR 613.4c). False, a layer-7b base 1/1 (CR 613.4b),
// so counters and pumps still apply on top, and a block rule on the
// blocker's side of the pair (Kargan Intimidator's CantBlockAttackers)
// that refuses a Detective attacker. "Detective" is a word of the Aura's
// text, read through it (CR 612.2), on both sides.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d1767573-713d-4f63-a09d-e94150655093",
		Name:            "Burden of Proof",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         EnchantCreature(),
		Static: []game.StaticAbility{
			{
				Label:    "Enchanted creature gets +2/+2 as long as it's a Detective you control.",
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return source.IsAttachedTo(target.InstanceID) && burdenOfProofYourDetective(target, source)
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power += 2
					c.Toughness += 2
				},
			},
			{
				Label:    "Otherwise, it has base power and toughness 1/1.",
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7B_Set,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return source.IsAttachedTo(target.InstanceID) && !burdenOfProofYourDetective(target, source)
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power = 1
					c.Toughness = 1
				},
			},
		},
		BlockRules: []game.BlockRule{
			CantBlockAttackers(
				func(_ *game.Game, c, source *game.Card) bool {
					return source != nil && source.IsAttachedTo(c.InstanceID) && !burdenOfProofYourDetective(c, source)
				},
				func(_ *game.Game, c, source *game.Card) bool {
					return source != nil && c.HasSubtype(source.TextSubtype("Detective"))
				},
				"Otherwise, it can't block Detectives"),
		},
	})
}

// burdenOfProofYourDetective is "it's a Detective you control", where
// "you" is the Aura's controller.
func burdenOfProofYourDetective(c, source *game.Card) bool {
	return c.Controller == source.Controller && c.HasSubtype(source.TextSubtype("Detective"))
}
