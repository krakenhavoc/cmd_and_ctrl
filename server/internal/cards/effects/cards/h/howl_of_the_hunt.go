package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Howl of the Hunt — Enchantment — Aura {2}{G} (Innistrad: Midnight Hunt):
//
//	"Flash
//	 Enchant creature
//	 When this Aura enters, if enchanted creature is a Wolf or Werewolf,
//	 untap that creature.
//	 Enchanted creature gets +2/+2 and has vigilance."
//
// The entry trigger is an intervening "if" (CR 603.4): it is checked as the
// Aura enters and again as the trigger resolves, against the creature the
// Aura enchants then, read through the layers. "Wolf" and "Werewolf" are
// words of this card's text (CR 612.2). The pump is PumpAttached and
// vigilance is GrantToAttached.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "22fe038e-033e-4632-a0d7-116492cded83",
		Name:            "Howl of the Hunt",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("vigilance"),
		},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(howlOfTheHuntWolf, effects.WhenThisEnters("Howl of the Hunt — untap enchanted creature if it is a Wolf or Werewolf",
				effects.Do(effects.Instruction(howlOfTheHuntUntap)))),
		},
	})
}

// howlOfTheHuntWolf is "if enchanted creature is a Wolf or Werewolf",
// read for the Aura as the ability resolves (or as it was attached when it
// left, CR 608.2h).
var howlOfTheHuntWolf = effects.CondBoard{Name: "enchanted creature is a Wolf or Werewolf", Test: func(ctx *effects.Context) bool {
	id, ok := effects.EnchantedCreatureAsItResolves(ctx)
	if !ok {
		return false
	}
	c, ok := ctx.Game.LookupCardForEffect(id)
	return ok && (c.HasSubtype(ctx.TextSubtype("Wolf")) || c.HasSubtype(ctx.TextSubtype("Werewolf")))
}}

func howlOfTheHuntUntap(ctx *effects.Context) error {
	if id, ok := effects.EnchantedCreatureAsItResolves(ctx); ok {
		return effects.UntapTarget{Target: id}.Apply(ctx)
	}
	return nil
}
