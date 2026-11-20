package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soulblade Corrupter — Creature — Human Warrior {4}{B}, 3/3 (Battlebond,
// #5529):
//
//	"Partner with Soulblade Renewer (When this creature enters, target
//	 player may put Soulblade Renewer into their hand from their library,
//	 then shuffle.)
//	 Deathtouch
//	 Whenever a creature with a +1/+1 counter on it attacks one of your
//	 opponents, that creature gains deathtouch until end of turn."
//
// Not legendary, so of CR 702.124j's two abilities only the entry search
// does anything (PartnerWith). Deathtouch rides PrintedKeywords.
//
// The attack trigger is one per attacker, whoever controls it: at a
// Commander table an opponent's creature attacking another of your
// opponents triggers it too, as printed. The creature is judged as it is
// declared (bbdCounteredCreatureAttacksAnOpponent): it has a +1/+1 counter
// and attacks a player other than the Corrupter's controller, so an attack
// at a planeswalker or battle does not trigger it. The grant goes to the
// attacker named by the event if it is still on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8d708573-10b8-44fb-b61b-9ad65e3eb465",
		Name:            "Soulblade Corrupter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			PartnerWith("Soulblade Corrupter", "Soulblade Renewer"),
			On(game.EventAttack, bbdCounteredCreatureAttacksAnOpponent,
				"Soulblade Corrupter — that creature gains deathtouch until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil || !onBattlefield(g, item.Trigger.Event.CardID) {
						return nil
					}
					return GrantKeywordUntilEOT{
						Target:   item.Trigger.Event.CardID,
						Keywords: []string{"deathtouch"},
						Label:    "Soulblade Corrupter — deathtouch until end of turn",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
