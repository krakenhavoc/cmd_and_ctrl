package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Outland Liberator // Frenzied Trapbreaker — {1}{G} Creature — Human
// Werewolf 2/2 // Creature — Werewolf 3/3 (#2586, ADR 0132):
//
//	Front: "{1}, Sacrifice this creature: Destroy target artifact or
//	        enchantment.
//	        Daybound"
//	Back:  "{1}, Sacrifice this creature: Destroy target artifact or
//	        enchantment.
//	        Whenever this creature attacks, destroy target artifact or
//	        enchantment defending player controls.
//	        Nightbound"
//
// The sacrifice ability is Cathar Commando's. The attack trigger's
// clause is built from the attack itself (TargetsFrom, CR 603.3d): the
// defending player is the one this creature is attacking (CR 802.2a),
// the same read Goblin Racketeer's target clause makes.
//
// No simplification.
func init() {
	const oracle = "9545f126-062f-4410-a362-e16255a128d6"
	sacrifice := func() effects.ActivatedAbility {
		return effects.ActivatedAbility{
			Label:   "{1}, Sacrifice this creature: Destroy target artifact or enchantment.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
			Effect:  effects.DestroyChosenPermanent,
		}
	}
	attack := effects.WheneverThisAttacks("Frenzied Trapbreaker — destroy target artifact or enchantment defending player controls", effects.DestroyChosenPermanent)
	attack.TargetsFrom = targetArtifactOrEnchantmentDefendingPlayerControls
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Outland Liberator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"daybound"},
		Activated:       []effects.ActivatedAbility{sacrifice()},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Frenzied Trapbreaker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"nightbound"},
		Activated:       []effects.ActivatedAbility{sacrifice()},
		Triggered:       []game.TriggeredAbility{attack},
	})
}

// targetArtifactOrEnchantmentDefendingPlayerControls is the TargetsFrom
// clause for "target artifact or enchantment defending player controls".
// It reads only the source's own attack, so a restore rebuilds the same
// clause.
func targetArtifactOrEnchantmentDefendingPlayerControls(_ game.TriggerContext, source *game.Card, g *game.Game) *game.TargetSpec {
	defender := g.DefendingPlayerForAttackForEffect(source.AttackingTarget)
	if defender == uuid.Nil {
		return nil
	}
	return &game.TargetSpec{
		Mode: "permanent", Label: "target artifact or enchantment defending player controls",
		Zones: []game.ZoneKind{game.ZoneBattlefield},
		CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return (c.IsArtifact() || c.IsEnchantment()) && c.Controller == defender
		},
		Min: 1, Max: 1,
	}
}
