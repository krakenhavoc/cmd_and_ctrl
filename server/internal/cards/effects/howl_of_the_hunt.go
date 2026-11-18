package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "22fe038e-033e-4632-a0d7-116492cded83",
		Name:            "Howl of the Hunt",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttached("vigilance"),
		},
		Triggered: []game.TriggeredAbility{
			InterveningIf(howlOfTheHuntWolf, WhenThisEnters("Howl of the Hunt — untap enchanted creature if it is a Wolf or Werewolf",
				Do(Instruction(howlOfTheHuntUntap)))),
		},
	})
}

// howlOfTheHuntWolf is "if enchanted creature is a Wolf or Werewolf",
// read for the Aura as the ability resolves (or as it was attached when it
// left, CR 608.2h).
var howlOfTheHuntWolf = CondBoard{Name: "enchanted creature is a Wolf or Werewolf", Test: func(ctx *Context) bool {
	id, ok := enchantedCreatureAsItResolves(ctx)
	if !ok {
		return false
	}
	c, ok := ctx.Game.LookupCardForEffect(id)
	return ok && (c.HasSubtype(ctx.TextSubtype("Wolf")) || c.HasSubtype(ctx.TextSubtype("Werewolf")))
}}

func howlOfTheHuntUntap(ctx *Context) error {
	if id, ok := enchantedCreatureAsItResolves(ctx); ok {
		return UntapTarget{Target: id}.Apply(ctx)
	}
	return nil
}
