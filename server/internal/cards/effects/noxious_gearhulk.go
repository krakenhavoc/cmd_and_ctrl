package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Noxious Gearhulk — Artifact Creature — Construct, {4}{B}{B}, 5/4
// (EDHREC rank 993):
//
//	"Menace
//	 When this creature enters, you may destroy another target
//	 creature. If a creature is destroyed this way, you gain life
//	 equal to its toughness."
//
// The black Gearhulk: a removal spell on a 5/4 menace body, with the
// life back. One optional, targeted ETB trigger. "Destroyed this way"
// is honoured: the toughness is read before the destroy, and the
// life is gained only if the creature actually left the battlefield
// — an indestructible creature stays and pays nothing.
//
// Sandbox simplification: "another" is enforced by NAME rather than
// by instance, because the trigger's target clause is declared at
// init() before any Gearhulk exists. In a singleton format that is
// the same creature; a token copy of the Gearhulk would also be
// excluded, which is WEAKER than printed, never stronger. The Effect
// declines its own source as well, so the restriction holds at
// resolution too.
func init() {
	Register(Spec{
		OracleID:        "a77b5be2-f361-4135-ba25-670a74d268ac",
		Name:            "Noxious Gearhulk",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"The trigger can't target another creature named Noxious Gearhulk, such as a token copy of it."},
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{{
			Watches:        []game.EventKind{game.EventETB},
			AppliesTo:      b06SelfETB,
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Noxious Gearhulk — destroy another target creature?"},
			Targets:        TargetCreature("another target creature", b03NotNamed("Noxious Gearhulk")),
			Key:            "Noxious Gearhulk — destroy another target creature, gain life equal to its toughness",
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				target := item.Targets[0].ID
				if target == item.SourceCardID {
					return nil
				}
				victim, ok := g.LookupCardForEffect(target)
				if !ok {
					return nil
				}
				toughness := victim.CurrentToughness()
				ctx := NewContext(g, item)
				if err := (DestroyTarget{Target: target}).Apply(ctx); err != nil {
					return err
				}
				if z := g.FindCardZoneForEffect(target); z != nil && z.Kind == game.ZoneBattlefield {
					return nil // indestructible: not destroyed this way
				}
				return GainLife{Player: item.Controller, Amount: toughness}.Apply(ctx)
			},
		}},
	})
}
