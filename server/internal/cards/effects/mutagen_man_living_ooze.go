package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mutagen Man, Living Ooze — {X}{G}{G} Legendary Creature — Ooze Mutant
// 2/3 (Teenage Mutant Ninja Turtles, #4457):
//
//	"Trample
//	 Activated abilities of artifact tokens you control cost {1} less to
//	 activate.
//	 When Mutagen Man enters, create X Mutagen tokens. (They're artifacts
//	 with "{1}, {T}, Sacrifice this token: Put a +1/+1 counter on target
//	 creature. Activate only as a sorcery.")"
//
// The discount is Sam, Loyal Attendant's activation pass
// (ActivationCostsLess): the ability's source is an artifact token its
// controller controls, and only generic mana comes off (CR 601.2f), so a
// Mutagen costs {T} and a sacrifice, and a Treasure is unchanged. The
// enters trigger is Farmer Cotton's: X is read from the cast as the
// trigger is put on the stack (CR 107.3m) and carried on the item, so it
// still resolves for that X if Mutagen Man leaves in response. Cast for
// X=0 it is still a 2/3 trampler, so it does not declare XMatters.
// The Mutagen token is the catalog's CR 111.10v template (MutagenToken, #4676).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "472d9a18-19a0-46c0-a4bf-328ebec7ce41",
		Name:            "Mutagen Man, Living Ooze",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		CostModifiers: []game.CostModifier{
			ActivationCostsLess(1,
				"Activated abilities of artifact tokens you control cost {1} less to activate.",
				ofAnArtifactTokenYouControl()),
		},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: Self,
			Key:       "Mutagen Man, Living Ooze — create X Mutagen tokens",
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Mutagen Man, Living Ooze — create X Mutagen tokens")
				item.Params.Amount = source.CastX()
				return item
			},
			Effect: Do(Instruction(func(ctx *Context) error {
				x := ctx.Item.Params.Amount
				if x <= 0 {
					return nil
				}
				return CreateToken{Template: MutagenToken(), N: x}.Apply(ctx)
			})),
		}},
	})
}

// ofAnArtifactTokenYouControl — "activated abilities of artifact tokens you
// control": the ability's source is an artifact token under the
// modifier's controller.
func ofAnArtifactTokenYouControl() CostPredicate {
	return func(q game.CostQuery) bool {
		return q.Card.Controller == q.Source.Controller && q.Card.IsArtifact() && IsToken(q.Card)
	}
}
