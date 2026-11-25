package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nevinyrral, Urborg Tyrant — Legendary Creature — Zombie Wizard {3}{W}{U}{B},
// 3/6 (Commander Legends):
//
//	"Hexproof from artifacts, creatures, and enchantments
//	 When Nevinyrral enters, create a tapped 2/2 black Zombie creature token for
//	 each creature that died this turn.
//	 When Nevinyrral dies, you may pay {1}. When you do, destroy all artifacts,
//	 creatures, and enchantments."
//
// Hexproof from card types is the quality grammar protection shares, one token
// per type. The enter trigger counts the table-wide tally of creatures that
// died this turn (Sautekh Immortal's), tokens included, as it resolves, and the
// Zombies enter tapped. The death trigger is Sparktongue Dragon's reflexive
// shape: only a paid {1} creates the "when you do" trigger (CR 603.12), whose
// body is Nevinyrral's Disk's sweep; Nevinyrral is already gone from it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3ddde895-12aa-488a-94fb-27a1064bdb72",
		Name:            "Nevinyrral, Urborg Tyrant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"hexproof from artifacts", "hexproof from creatures", "hexproof from enchantments"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Nevinyrral — create a tapped 2/2 black Zombie creature token for each creature that died this turn",
				effects.Do(effects.S82OldB083TappedTokensForEachCreatureDied{Template: effects.TokenCard("2/2 black Zombie")})),
			effects.WhenThisDies(nevinyrralLabel, nevinyrralOffer),
		},
	})
}

const nevinyrralLabel = "Nevinyrral — you may pay {1}; when you do, destroy all artifacts, creatures, and enchantments"

func nevinyrralOffer(g *game.Game, item *game.StackItem) error {
	return effects.MayPay{
		Chooser:  item.Controller,
		Cost:     "{1}",
		Question: nevinyrralLabel + "?",
		Then:     nevinyrralPaid,
	}.Apply(effects.NewContext(g, item))
}

var nevinyrralPaid = effects.PayUnlessBranches("pay-unless/nevinyrral-sweep",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.ReflexiveTrigger{
			Label: "Nevinyrral — destroy all artifacts, creatures, and enchantments",
			Body:  nevinyrralSweepBody,
		}.Apply(ctx)
	}, nil)

var nevinyrralSweepBody = game.ReflexiveBody("nevinyrral-urborg-tyrant/sweep",
	effects.SimpleBody(effects.TspB2DestroyArtifactsCreaturesEnchantments), nil)
