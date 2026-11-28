package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hakim, Loreweaver — Legendary Creature — Human Wizard {3}{U}{U}, 2/4 (Mirage,
// S82 OLD batch 812, #5633):
//
//	"Flying
//	 {U}{U}: Return target Aura card from your graveyard to the battlefield
//	 attached to Hakim. Activate only during your upkeep and only if Hakim isn't
//	 enchanted.
//	 {U}{U}, {T}: Destroy all Auras attached to Hakim."
//
// The first ability's two gates are both announce-time conditions (upkeep of
// its controller's turn; no Aura attached to Hakim). The Aura card enters
// attached to Hakim (CR 303.4f), only if it can enchant him (Bruna's
// CanBeAttachedTo). The second destroys every Aura attached to Hakim as it
// resolves; none is nothing to do.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "43e99cba-a6e3-4abc-a92b-aaf5839c72d2",
		Name:            "Hakim, Loreweaver",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{
			{
				Label: "{U}{U}: Return target Aura card from your graveyard to the battlefield attached to Hakim. Activate only during your upkeep and only if Hakim isn't enchanted.",
				Cost:  effects.ManaCost("{U}{U}"),
				Condition: effects.AllConditions(effects.DuringYourUpkeep(), func(g *game.Game, _, source uuid.UUID) bool {
					return !effects.S82OldB812HasAnAuraAttached(g, source)
				}),
				Targets: effects.TargetCardInGraveyard("target Aura card from your graveyard", effects.OfSubtype("Aura"), effects.YouOwn()),
				Effect:  effects.S82OldB812HakimReturnsAnAura,
			},
			{
				Label:   "{U}{U}, {T}: Destroy all Auras attached to Hakim.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{U}{U}"), effects.TapCost()),
				Effect:  effects.S82OldB812DestroyAurasAttachedToThis,
			},
		},
	})
}
