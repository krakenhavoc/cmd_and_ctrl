package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Military Discipline — Enchantment — Aura {W} (The Brothers' War):
//
//	"Flash
//	 Enchant creature
//	 When this Aura enters, enchanted creature gains first strike until end
//	 of turn.
//	 Enchanted creature gets +1/+0."
//
// The enters trigger reads the host from the Aura as it resolves, so a
// creature that left in response gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c910c1c0-2b78-4e0f-bb55-e0581b2e07e8",
		Name:            "Military Discipline",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Static:          []game.StaticAbility{PumpAttached(1, 0)},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Military Discipline — enchanted creature gains first strike until end of turn",
				militaryDisciplineFirstStrike),
		},
	})
}

func militaryDisciplineFirstStrike(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	aura, ok := ctx.TriggeringPermanent()
	if !ok || aura.AttachedTo.Kind != game.TargetCard {
		return nil
	}
	return GrantKeywordUntilEOT{
		Target:   aura.AttachedTo.ID,
		Keywords: []string{"first strike"},
		Label:    "Military Discipline — first strike until end of turn",
	}.Apply(ctx)
}
