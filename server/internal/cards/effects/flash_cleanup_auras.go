package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// flash_cleanup_auras.go — the Visions-era commons and uncommons that
// print the same rider as Necromancy (#3712, on #3156):
//
//	"You may cast this spell as though it had flash. If you cast it any
//	 time a sorcery couldn't have been cast, the controller of the
//	 permanent it becomes sacrifices it at the beginning of the next
//	 cleanup step."
//
// Each sets Spec.FlashWithCleanupSacrifice (game/flash_cleanup_sacrifice.go)
// and otherwise reads like any other Aura of its kind:
//
//	Lightning Reflexes {1}{R}  Enchanted creature gets +1/+0 and has first strike.
//	Spider Climb       {G}     Enchanted creature gets +0/+3 and has reach.
//	Soar               {1}{U}  Enchanted creature gets +0/+1 and has flying.
//	Armor of Thorns    {1}{G}  Enchant nonblack creature; gets +2/+2.
//	Grave Servitude    {1}{B}  Enchanted creature gets +3/-1 and is black.
//	Mystic Veil        {1}{U}  Enchanted creature has shroud.
//	Relic Ward         {1}{W}  Enchant artifact; it has shroud.
//	Ward of Lights     {W}{W}  As this Aura enters, choose a color; protection from it.
//	Parapet            {1}{W}  Creatures you control get +0/+1 (not an Aura).
//
// Shroud and protection are the engine's own keywords and tokens, so
// the targeting and damage gates read them with no card-side code.
// No simplification.
func init() {
	aura := func(oracleID, name string, target *game.TargetSpec, statics ...game.StaticAbility) Spec {
		return Spec{
			OracleID:                  oracleID,
			Name:                      name,
			Completeness:              CompletenessFull,
			FlashWithCleanupSacrifice: true,
			Targets:                   target,
			Purpose:                   game.Purpose{Aura: game.AuraBenefit},
			Static:                    statics,
		}
	}

	Register(aura("2b0c82f1-078a-4e8b-8e43-407f5c5d8549", "Lightning Reflexes", EnchantCreature(),
		PumpAttached(1, 0), GrantToAttached("first strike")))
	Register(aura("2636aafa-83f9-4b72-b729-04e1c3aa0eec", "Spider Climb", EnchantCreature(),
		PumpAttached(0, 3), GrantToAttached("reach")))
	Register(aura("acee6f4f-8b66-495d-bb0e-71c339e7dc49", "Soar", EnchantCreature(),
		PumpAttached(0, 1), GrantToAttached("flying")))
	Register(aura("31ead5a5-aa01-4601-8d3e-267b65fce3bb", "Armor of Thorns",
		TargetCreature("enchant nonblack creature", NonBlack()),
		PumpAttached(2, 2)))
	Register(aura("6f54e7ea-bb87-4ce0-96b6-4fc82605ed2a", "Grave Servitude", EnchantCreature(),
		PumpAttached(3, -1), SetAttachedColors("B")))
	Register(aura("43c88fcd-aac9-4741-8fd6-31ac81c5bb6f", "Mystic Veil", EnchantCreature(),
		GrantToAttached("shroud")))
	Register(aura("c6321de2-80b0-4679-93cf-97df75344e97", "Relic Ward",
		TargetPermanent("enchant artifact", Artifact()),
		GrantToAttached("shroud")))

	ward := aura("edc05b33-2bfa-4692-bb3b-26b7cf111ce8", "Ward of Lights", EnchantCreature(),
		game.StaticAbility{
			Layer:     game.Layer6Ability,
			Label:     "Enchanted creature has protection from the chosen color.",
			AppliesTo: AttachedToSource,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
				tok := game.ProtectionFromColor(source.ChosenColor)
				if tok != "" && !slices.Contains(c.Abilities, tok) {
					c.Abilities = append(c.Abilities, tok)
				}
			},
		})
	ward.AsEnters = ChooseColorAsEnters(game.ColorForProtection, "Ward of Lights")
	Register(ward)

	Register(Spec{
		OracleID:                  "4e73634a-9df8-4c36-b1d2-f366eae902fc",
		Name:                      "Parapet",
		Completeness:              CompletenessFull,
		FlashWithCleanupSacrifice: true,
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			Label:    "Creatures you control get +0/+1.",
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Toughness++
			},
		}},
	})
}
