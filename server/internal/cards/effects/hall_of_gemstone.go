package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hall of Gemstone — World Enchantment {1}{G}{G}:
//
//	"At the beginning of each player's upkeep, that player chooses a
//	 color. Until end of turn, lands tapped for mana produce mana of the
//	 chosen color instead of any other color."
//
// #1997, one of the world enchantments ADR 0109 PR 7 left out. The
// upkeep player (not the Hall's controller) chooses the colour as the
// trigger resolves. The rest of the turn, every land any player taps for
// mana makes the chosen colour instead of any other colour: a CR 614
// replacement on the mana production (CR 106.12b), recorded as data
// (game.LandsTappedProduceColorThisTurnForEffect, controlled by the
// player who chose). Colourless mana is not a colour (CR 106.1a), so it
// stays colourless, and the amount never changes. The effect comes from
// the resolved trigger (CR 611.2a), so it lasts the turn even if the
// Hall leaves. The auto-tap planner books each land as the colour it
// will really make, so "auto-tap" pays with it.
//
// A world permanent: the world rule (CR 704.5k, ADR 0109 §8) puts it
// into its owner's graveyard when a newer one enters.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "3da9fecc-064a-44e6-b88e-4b10194e367e",
		Name:         "Hall of Gemstone",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtEachUpkeep("Hall of Gemstone — that player chooses a color; lands tapped for mana produce it this turn", hallOfGemstoneUpkeep),
		},
	})
}

func hallOfGemstoneUpkeep(g *game.Game, item *game.StackItem) error {
	player := triggeringActor(item)
	if player == uuid.Nil || g.PlayerByIDForEffect(player) == nil {
		return nil
	}
	source := item.SourceCardID
	ChooseColorThen(game.ColorForMana, g, player, source, "Hall of Gemstone — choose a color; lands tapped for mana produce it this turn",
		func(g *game.Game, color string) error {
			g.LandsTappedProduceColorThisTurnForEffect(source, player, color,
				"Hall of Gemstone — lands tapped for mana produce "+game.ColorName(color)+" mana this turn")
			return nil
		})
	return nil
}
