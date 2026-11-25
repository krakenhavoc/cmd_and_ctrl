package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reidane, God of the Worthy // Valkmira, Protector's Shield — modal double-
// faced card, {2}{W} // {3}{W} (Kaldheim).
//
// Front, Legendary Creature — God, 2/3:
//
//	"Flying, vigilance
//	 Snow lands your opponents control enter tapped.
//	 Noncreature spells your opponents cast with mana value 4 or greater cost
//	 {2} more to cast."
//
// The land clause is a CR 614 entry replacement over the lands an opponent's
// entry would put onto the battlefield (S82OldB064OpponentsSnowLandsEnterTapped).
// The tax is a cost increase (CR 601.2f) on opponents' spells that are not
// creatures, judged by mana value with {X} counted at the chosen value.
//
// Back, Legendary Artifact (registered under "<oracle_id>#1", ADR 0034):
//
//	"If a source an opponent controls would deal damage to you or a permanent
//	 you control, prevent 1 of that damage.
//	 Whenever you or another permanent you control becomes the target of a
//	 spell or ability an opponent controls, counter that spell or ability
//	 unless its controller pays {1}."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "35ae7b25-6d0f-49f8-a40d-dd2cf88e6f44",
		Name:            "Reidane, God of the Worthy",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Replacements: []game.ReplacementEffect{
			effects.S82OldB064OpponentsSnowLandsEnterTapped("Reidane, God of the Worthy: snow lands enter tapped"),
		},
		CostModifiers: []game.CostModifier{
			effects.CostsMore(2, "Noncreature spells your opponents cast with mana value 4 or greater cost {2} more to cast.",
				effects.OpponentsSpell(), effects.NoncreatureSpell(), effects.SpellManaValueAtLeast(4)),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "35ae7b25-6d0f-49f8-a40d-dd2cf88e6f44#1",
		Name:         "Valkmira, Protector's Shield",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.S82OldB064PreventOneFromOpponents()},
		Triggered: []game.TriggeredAbility{
			effects.S82OldB064ValkmiraWard("Valkmira, Protector's Shield — whenever you or another permanent you control becomes the target of a spell or ability an opponent controls, counter that spell or ability unless its controller pays {1}"),
		},
	})
}
