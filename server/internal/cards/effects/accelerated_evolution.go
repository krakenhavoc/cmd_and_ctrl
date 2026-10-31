package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Accelerated Evolution — Enchantment — Aura {2}{G} (Marvel Super
// Heroes, #4455):
//
//	"Flash
//	 Enchant creature you control
//	 When this Aura enters, enchanted creature gains hexproof until end
//	 of turn. (It can't be the target of spells or abilities your
//	 opponents control.)
//	 Enchanted creature gets +2/+2."
//
// Airtight Alibi's shape without its untap and suspicion clauses: a
// flash Aura on a creature you control, a +2/+2 static on the host, and
// an enters trigger that grants hexproof until end of turn to the host
// as it resolves (CR 608.2h, the way Airtight Alibi reads it).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6401d0ae-4635-497d-91ef-ce2eabefb893",
		Name:            "Accelerated Evolution",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         EnchantCreature(YouControl()),
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Accelerated Evolution — enchanted creature gains hexproof until end of turn", acceleratedEvolutionEnters),
		},
	})
}

// acceleratedEvolutionEnters gives the enchanted creature hexproof until
// end of turn, or does nothing if the host is no longer there.
func acceleratedEvolutionEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	host, ok := enchantedCreatureAsItResolves(ctx)
	if !ok {
		return nil
	}
	return GrantKeywordUntilEOT{
		Target:   host,
		Keywords: []string{"hexproof"},
		Label:    "Accelerated Evolution — hexproof until end of turn",
	}.Apply(ctx)
}
