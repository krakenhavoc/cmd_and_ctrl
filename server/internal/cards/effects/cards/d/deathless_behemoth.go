package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deathless Behemoth — Creature — Eldrazi {6}, 6/6 (Battle for Zendikar,
// #5547):
//
//	"Vigilance
//	 Sacrifice two Eldrazi Scions: Return this card from your graveyard to
//	 your hand. Activate only as a sorcery."
//
// The ability works from the graveyard only (Zones), and the Scions are the
// cost, so the activation is refused without two of them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "78ee8cd8-de12-440f-8890-a78ab40b6be4",
		Name:            "Deathless Behemoth",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []effects.ActivatedAbility{{
			Label:        "Sacrifice two Eldrazi Scions: Return this card from your graveyard to your hand. Activate only as a sorcery.",
			Cost:         effects.SacrificeN(2, "two Eldrazi Scions", effects.HasSubtype("Eldrazi"), effects.HasSubtype("Scion")),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Effect:       effects.ReturnThisCardFromYourGraveyardToYourHand,
		}},
	})
}
