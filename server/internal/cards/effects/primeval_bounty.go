package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Primeval Bounty — Enchantment {5}{G} (EDHREC rank 2177):
//
//	"Whenever you cast a creature spell, create a 3/3 green Beast
//	 creature token.
//	 Whenever you cast a noncreature spell, put three +1/+1 counters
//	 on target creature you control.
//	 Landfall — Whenever a land you control enters, you gain 3 life."
//
// The six-mana value engine, three triggers. The two cast triggers
// read the spell off the stack, where its type line is intact
// (b12CreatureSpellCastByYou / b10NoncreatureSpellCastByYou); the
// counter trigger targets a creature the controller controls, so
// with none on the battlefield it is removed (CR 603.3d) and the
// target is re-checked as it resolves. Landfall is a land entering
// under the controller's control from anywhere — a play, a fetch, a
// Cultivate — and gains 3 each time.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f633c16d-d943-421c-ad18-af35db9ec9fc",
		Name:         "Primeval Bounty",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b12CreatureSpellCastByYou(ev, source, g)
			}, "Primeval Bounty — create a 3/3 Beast", Do(CreateToken{Template: TokenCard("3/3 colorless Beast"), N: 1})),
			{
				Watches:   []game.EventKind{game.EventCast},
				AppliesTo: b10NoncreatureSpellCastByYou,
				Targets:   TargetCreature("target creature you control", YouControl()),
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Primeval Bounty — put three +1/+1 counters on target creature you control",
						func(g *game.Game, item *game.StackItem) error {
							ctx := NewContext(g, item)
							id, ok := b16FirstLegalTargetCard(ctx)
							if !ok {
								return nil
							}
							return AddCounter{Target: id, Kind: "+1/+1", N: 3}.Apply(ctx)
						})
				},
			},
			Landfall("Primeval Bounty — gain 3 life", Do(GainLife{Amount: 3})),
		},
	})
}
