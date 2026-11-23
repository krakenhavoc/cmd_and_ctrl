package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seasoned Dungeoneer — Creature — Human Warrior {3}{W}, 3/4 (Commander
// Legends: Battle for Baldur's Gate, #5948):
//
//	"When this creature enters, you take the initiative.
//	 Whenever you attack, target attacking Cleric, Rogue, Warrior, or Wizard
//	 gains protection from creatures until end of turn. It explores."
//
// The initiative is Feywild Caretaker's shared entry trigger. "Whenever you
// attack" is Battlesong Berserker's once-per-declaration trigger; the target
// is an attacking creature of one of four types (layered subtypes), chosen
// as the trigger goes on the stack. It gains protection from creatures (CR
// 702.16) and then explores (CR 701.44): the explorer is the target if it is
// still legal as the ability resolves, and the protection and the explore
// both belong to that one creature (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5ea675aa-88d6-4387-a14d-e606fa9c3aec",
		Name:         "Seasoned Dungeoneer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouTakeTheInitiative("Seasoned Dungeoneer"),
			effects.OncePerBatch(effects.Targeting(
				effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.AttackDeclaredByYou(ev, source.Controller)
				}, "Seasoned Dungeoneer — target attacking Cleric, Rogue, Warrior, or Wizard gains protection from creatures until end of turn and explores",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"protection from creatures"},
							Label: "Seasoned Dungeoneer — protection from creatures"}).Apply(ctx); err != nil {
							return err
						}
						return effects.Explores{Explorer: id}.Apply(ctx)
					}),
				effects.TargetCreature("target attacking Cleric, Rogue, Warrior, or Wizard",
					effects.AttackingCreature(), effects.AnySubtype("Cleric", "Rogue", "Warrior", "Wizard")))),
		},
	})
}
