package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cataclysmic Gearhulk — Artifact Creature — Construct {3}{W}{W}, 4/5:
//
//	"Vigilance
//	 When this creature enters, each player chooses an artifact, a creature,
//	 an enchantment, and a planeswalker from among the nonland permanents
//	 they control, then sacrifices the rest."
//
// Each player is asked once, over their nonland permanents, for the ones
// they keep, and when everyone has answered everything else of theirs that
// is not a land is sacrificed at once (CR 701.21a). The set rule is
// Liliana, Dreadhorde General's keep-set (lilianaKeepsOneOfEachType) over the types present, only these four being offered: every
// type the player has among their nonland permanents has a kept permanent of
// that type, and each kept permanent stands for a type of its own, so an
// artifact creature may be both the artifact and the creature but then
// nothing else need be kept for either. A permanent with none of the four
// types (a battle) can't be chosen and is sacrificed. A player who would
// keep everything anyway is not asked, and lands are never touched.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2e4865c5-a590-4820-81ec-926014ddb8c4",
		Name:            "Cataclysmic Gearhulk",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Cataclysmic Gearhulk — each player chooses an artifact, a creature, an enchantment, and a planeswalker from among the nonland permanents they control, then sacrifices the rest",
				cataclysmicGearhulkEffect),
		},
	})
}

func cataclysmicGearhulkEffect(g *game.Game, item *game.StackItem) error {
	return effects.GearhulkKeepOneOfEach(g, item, false)
}
