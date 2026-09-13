package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// b15GalaGreetersLabel is the stack label the per-turn tally keys
// on — every resolution of this trigger is one mode used up.
const b15GalaGreetersLabel = "Gala Greeters — alliance"

// Gala Greeters — Creature — Elf Druid {1}{G}, 1/1 (EDHREC rank
// 1706):
//
//	"Alliance — Whenever another creature you control enters, choose
//	 one that hasn't been chosen this turn —
//	 • Put a +1/+1 counter on this creature.
//	 • Create a tapped Treasure token.
//	 • You gain 2 life."
//
// A two-drop that pays a counter, a Treasure and two life for the
// first three creatures each turn. Corpse Knight's condition
// (b13AnotherCreatureYouControlEntered); the modes are each one
// primitive.
//
// Sandbox simplification: the mode is not chosen — it is the first
// mode in printed order that has not been used this turn. The modal
// machinery is cast-time only (Spec.Modes rides cast_spell), so a
// resolution-time pick for a trigger has no prompt; Tireless
// Provisioner's "always a Treasure" is the same posture. "Hasn't
// been chosen this turn" is a count of this trigger's resolutions
// since the turn's upkeep began (b15ResolvedThisTurn — resolutions,
// not triggers, so two Greeters triggers queued together each take
// their own mode). The fourth creature and later do nothing, as
// printed. Weaker than printed — the controller cannot take the
// Treasure before the counter — never stronger.
func init() {
	Register(Spec{
		OracleID:     "cce081eb-8820-415a-a7b2-3c5b9d4a2601",
		Name:         "Gala Greeters",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The mode isn't chosen: each turn the first creature gives the +1/+1 counter, the second the tapped Treasure, the third the 2 life."},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b13AnotherCreatureYouControlEntered(ev, source, g)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, b15GalaGreetersLabel,
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						// The resolve event for THIS resolution is already
						// logged, so the count includes it: 1 → first mode.
						switch b15ResolvedThisTurn(g, item.SourceCardID, b15GalaGreetersLabel) {
						case 1:
							if !b15OnBattlefield(g, item.SourceCardID) {
								return nil
							}
							return AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: 1}.Apply(ctx)
						case 2:
							return b13CreateTappedTreasures(ctx, item.Controller, 1)
						case 3:
							return GainLife{Player: item.Controller, Amount: 2}.Apply(ctx)
						}
						return nil
					})
			},
		}},
	})
}
