package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Super Speed — {R} Enchantment — Aura (Marvel Super Heroes, #4455):
//
//	"Flash
//	 Enchant creature
//	 When this Aura enters, enchanted creature gains first strike until
//	 end of turn.
//	 Enchanted creature gets +1/+0 and has haste."
//
// Airtight Alibi's shape: the enters trigger names "enchanted creature" as
// it resolves, then grants the keyword until end of turn; the +1/+0 and
// haste are the Aura's own statics and last as long as it is attached.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "88923ca1-a793-42f0-b9f8-ed9ff9c1185d",
		Name:            "Super Speed",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         EnchantCreature(),
		Static: []game.StaticAbility{
			PumpAttached(1, 0),
			GrantToAttached("haste"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Super Speed — enchanted creature gains first strike until end of turn", superSpeedEnters),
		},
	})
}

// superSpeedEnters gives the enchanted creature first strike until end of
// turn.
func superSpeedEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	host, ok := enchantedCreatureAsItResolves(ctx)
	if !ok {
		return nil
	}
	return GrantKeywordUntilEOT{
		Target:   host,
		Keywords: []string{"first strike"},
		Label:    "Super Speed — first strike until end of turn",
	}.Apply(ctx)
}
