package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Treacherous Vampire — Creature — Vampire {4}{B}, 4/4 (Judgment, #5611):
//
//	"Flying
//	 Whenever this creature attacks or blocks, sacrifice it unless you
//	 exile a card from your graveyard.
//	 Threshold — As long as there are seven or more cards in your
//	 graveyard, this creature gets +2/+2 and has "When this creature dies,
//	 you lose 6 life.""
//
// The attack-or-block trigger is one registered trigger asking the pay-unless
// prompt; the threshold half is Fledgling Dragon's pair of a layer 7c
// self-modifier and a layer 6 self-grant of a dies trigger.
//
// No simplification.
func init() {
	grant, static := effects.OldB622ThresholdSelfDiesLoseLife("Treacherous Vampire", "treacherous-vampire/dies-lose-6", 6)
	effects.Register(effects.Spec{
		OracleID:        "1d33d49f-ba12-4bf6-878b-04cee5d5b48d",
		Name:            "Treacherous Vampire",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventAttack, game.EventBlock}, effects.ThisAttacksOrBlocks,
				"Treacherous Vampire — sacrifice it unless you exile a card from your graveyard",
				effects.OldB622SacrificeUnlessYouExileFromYourGraveyard("Treacherous Vampire")),
		},
		Grants: []effects.AbilityGrant{grant},
		Static: []game.StaticAbility{effects.FraThresholdSelfPT(2, 2), static},
	})
}
