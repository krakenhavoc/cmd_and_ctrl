package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gilt-Leaf's Embrace — Enchantment — Aura {2}{G} (Lorwyn Eclipsed,
// #4458):
//
//	"Flash
//	 Enchant creature
//	 When this Aura enters, enchanted creature gains trample and
//	 indestructible until end of turn. (Damage and effects that say
//	 "destroy" don't destroy it. If its toughness is 0 or less, it still
//	 dies.)
//	 Enchanted creature gets +2/+0."
//
// Captain's Tutelage's shape with two keywords: the enters trigger names
// "enchanted creature" as it resolves, or as the Aura last was if the Aura
// has gone (CR 608.2h), and grants trample and indestructible until end of
// turn; the +2/+0 is the Aura's own static and lasts as long as it is
// attached.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5eb8bb1d-7567-4c24-913b-463b5fe463c4",
		Name:            "Gilt-Leaf's Embrace",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         EnchantCreature(),
		Static:          []game.StaticAbility{PumpAttached(2, 0)},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Gilt-Leaf's Embrace — enchanted creature gains trample and indestructible until end of turn", giltLeafsEmbraceEnters),
		},
	})
}

func giltLeafsEmbraceEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	host, ok := enchantedCreatureAsItResolves(ctx)
	if !ok {
		return nil
	}
	return GrantKeywordUntilEOT{
		Target:   host,
		Keywords: []string{"trample", "indestructible"},
		Label:    "Gilt-Leaf's Embrace — trample and indestructible until end of turn",
	}.Apply(ctx)
}
