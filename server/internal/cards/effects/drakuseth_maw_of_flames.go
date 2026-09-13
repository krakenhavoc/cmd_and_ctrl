package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Drakuseth, Maw of Flames — Legendary Creature — Dragon {4}{R}{R}{R},
// 7/7 (EDHREC rank 1114):
//
//	"Flying
//	 Whenever Drakuseth attacks, it deals 4 damage to any target and 3
//	 damage to each of up to two other targets."
//
// The seven-mana Dragon that clears a board on the way in. The attack
// trigger is ONE target clause with one to three slots — "any target"
// plus "up to two other targets" — answered through the multi-slot
// pick_target prompt, and the effect is positional (Arc Trail's
// shape): the first slot takes 4, the second and third take 3 each.
// Every slot is re-checked at resolution on its own (CR 608.2b), so a
// target that left in response is skipped and the rest are still
// hit. The three targets must be distinct, which is what "other"
// means and what the picker enforces.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "060deaff-44d6-4f03-9568-bcb7add80255",
		Name:            "Drakuseth, Maw of Flames",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclared(ev, source)
			},
			Targets: b10DrakusethTargets(),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Drakuseth — 4 damage to the first target, 3 to each other",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						for i, t := range item.Targets {
							if !ctx.IsTargetLegal(t) {
								continue
							}
							amount := 3
							if i == 0 {
								amount = 4
							}
							if err := (DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: amount}).Apply(ctx); err != nil {
								return err
							}
						}
						return nil
					})
			},
		}},
	})
}

// b10DrakusethTargets is "any target, and up to two other targets": one
// to three distinct any-target slots.
func b10DrakusethTargets() *game.TargetSpec {
	spec := TargetAny().WithCount(1, 3)
	spec.Label = "any target, then up to two other targets"
	return spec
}
