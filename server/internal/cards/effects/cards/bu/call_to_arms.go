package cardsbu

import (
	"slices"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Call to Arms — Enchantment {1}{W} (Ice Age, S82 OLD batch 846, #5636):
//
//	"As this enchantment enters, choose a color and an opponent.
//	 White creatures get +1/+1 as long as the chosen color is the most common
//	 color among nontoken permanents the chosen player controls but isn't
//	 tied for most common.
//	 When the chosen color isn't the most common color among nontoken
//	 permanents the chosen player controls or is tied for most common,
//	 sacrifice this enchantment."
//
// Both choices are made as it enters (Black Vise's opponent, a colour as
// Coldsteel Heart's): stored on the permanent. The count is
// ColorIsMostCommonUntiedAmongNontoken (#6207), layered colours of the chosen
// player's nontoken permanents. The bonus is a layer 7c static over every white
// creature, whoever controls it; the sacrifice is a CR 603.8 state trigger,
// which asks the same question the other way round and waits for both choices
// (an unchosen colour or player holds nothing).
//
// No simplification.
func init() {
	chosen := func(g *game.Game, source *game.Card) (uuid.UUID, string, bool) {
		player := g.ChosenPlayerOf(source.InstanceID)
		color := g.ChosenColorOf(source.InstanceID)
		return player, color, player != uuid.Nil && color != ""
	}
	askColor := effects.ChooseColorAsEnters(game.ColorForFilter, "Call to Arms")
	askPlayer := effects.ChoosePlayerAsEnters("Call to Arms", effects.Opponents)
	effects.Register(effects.Spec{
		OracleID:     "72eaaa73-3263-46e9-aa86-5862d3aed20a",
		Name:         "Call to Arms",
		Completeness: effects.CompletenessFull,
		AsEnters: func(card *game.Card, ctx *effects.Context) error {
			if err := askColor(card, ctx); err != nil {
				return err
			}
			return askPlayer(card, ctx)
		},
		Static: []game.StaticAbility{{
			Label: "White creatures get +1/+1 as long as the chosen color is the most common color among nontoken permanents the chosen player controls but isn't tied for most common.",
			Layer: game.Layer7PT, SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				if !target.IsCreature() || !slices.Contains(target.EffectiveColors(), "W") {
					return false
				}
				player, color, ok := chosen(g, source)
				return ok && effects.ColorIsMostCommonUntiedAmongNontoken(g, player, color)
			},
			Apply: effects.PumpBy(1, 1),
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenState("Call to Arms — sacrifice it", func(g *game.Game, source *game.Card, _ uuid.UUID) bool {
				player, color, ok := chosen(g, source)
				return ok && !effects.ColorIsMostCommonUntiedAmongNontoken(g, player, strings.ToUpper(color))
			}, effects.SacrificeThisIfStillOnBattlefield),
		},
	})
}
