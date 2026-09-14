package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dwynen, Gilt-Leaf Daen — Legendary Creature — Elf Warrior {2}{G}{G},
// 3/4 (EDHREC rank 2630):
//
//	"Reach
//	 Other Elf creatures you control get +1/+1.
//	 Whenever Dwynen attacks, you gain 1 life for each attacking Elf
//	 you control."
//
// The Elf lord with a lifegain rider. Reach rides PrintedKeywords;
// the anthem is TribalAnthem over "other Elves you control" — both
// words printed, so it buffs only your side; the attack trigger is
// attackDeclared, and the count is taken as the trigger RESOLVES
// (b24AttackingCreaturesOfSubtype), so an Elf declared after Dwynen
// in the same declaration still counts. Dwynen herself is an
// attacking Elf and counts, as printed. Post-layer subtype, so a
// changeling counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "30d0d75f-e94c-460b-b957-9f1d655c0f65",
		Name:            "Dwynen, Gilt-Leaf Daen",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Elf"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclared(ev, source)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Dwynen, Gilt-Leaf Daen — gain 1 life for each attacking Elf you control",
					func(g *game.Game, item *game.StackItem) error {
						n := b24AttackingCreaturesOfSubtype(g, item.Controller, "Elf")
						return GainLife{Player: item.Controller, Amount: n}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
