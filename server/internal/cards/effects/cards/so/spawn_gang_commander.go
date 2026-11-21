package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spawn-Gang Commander — Creature — Eldrazi Goblin {3}{R}{R}, 2/2 (Modern
// Horizons 3):
//
//	"Devoid (This card has no color.)
//	 When you cast this spell, create three 0/1 colorless Eldrazi Spawn
//	 creature tokens with "Sacrifice this token: Add {C}."
//	 {1}{C}, Sacrifice an Eldrazi: This creature deals 2 damage to any
//	 target."
//
// Abundant Maw's cast trigger (it resolves above the spell, so the
// Spawn arrive even if the Commander is countered) and Airdrop Condor's
// sacrifice cost with the Eldrazi subtype. The {C} is colourless mana
// specifically; a Spawn is an Eldrazi, so the three tokens feed it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3641ceef-4f26-4647-ac12-c958524d68ca",
		Name:            "Spawn-Gang Commander",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Triggered: []game.TriggeredAbility{
			effects.WhenYouCastThisSpell("Spawn-Gang Commander — create three 0/1 colorless Eldrazi Spawn",
				effects.Do(effects.CreateToken{Template: effects.EldraziSpawnToken(), N: 3})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{C}, Sacrifice an Eldrazi: This creature deals 2 damage to any target.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Cost:    effects.Plus(effects.ManaCost("{1}{C}"), game.AbilityCost{SacrificeOther: effects.SacrificeSpec("an Eldrazi", effects.OfCreatureType("Eldrazi"))}),
			Targets: effects.TargetAny(),
			Effect:  effects.DamageToFirstTarget(2),
		}},
	})
}
