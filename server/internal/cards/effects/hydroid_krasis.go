package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hydroid Krasis — Creature — Jellyfish Hydra Beast {X}{G}{U}, 0/0
// (EDHREC rank 1536):
//
//	"When you cast this spell, you gain half X life and draw half X
//	 cards. Round down each time.
//	 Flying, trample
//	 This creature enters with X +1/+1 counters on it."
//
// The Simic X-drop whose payout cannot be countered. The cast trigger
// is a FromStack ability (cascade's mechanism): it fires when the
// spell is announced, goes on the stack above it, and resolves first
// — so a Counterspell on the Krasis still leaves the life and the
// cards behind, as printed. X is read off the stack item at trigger
// time and halved, rounding down, each half separately.
//
// Sandbox simplification, declared, for the X counters — Goldvein
// Hydra's: an entry replacement cannot see the X announced for the
// spell (the stack item is gone by the time the entry pipeline
// runs), so the counters go on as the spell RESOLVES, a beat before
// the card moves from the stack to the battlefield, which is the last
// moment X is readable. They are on the card when it lands, so the
// 0/0 body never meets the state-based check without them, every ETB
// watcher sees the finished creature, and a counter doubler or
// Hardened Scales applies, exactly as it does to a printed "enters
// with". The one observable difference is that a "whenever you put
// counters on a permanent" payoff does not see them, because the card
// was not a permanent yet — weaker, never stronger.
//
// One engine-side gap, not the card's: cast for X=0 or X=1 the Krasis
// is a printed 0/0 with no counters, which the toughness state check
// deliberately skips (the placeholder convention on Card.Power), so
// it stays on the battlefield instead of dying at once.
func init() {
	Register(Spec{
		OracleID:        "6bd872b2-5c40-4e11-9a7f-0136a51b0642",
		Name:            "Hydroid Krasis",
		XMatters:        true,
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"The X +1/+1 counters are put on the Krasis as the spell resolves, a beat before it enters, so effects that watch you put counters on a permanent don't see them."},
		PrintedKeywords: []string{"flying", "trample"},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: ctx.X()}.Apply(ctx)
		},
		Triggered: []game.TriggeredAbility{{
			FromStack: true,
			Watches:   []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				half := 0
				if item := g.StackItemForEffect(source.InstanceID); item != nil && item.XValue > 0 {
					half = item.XValue / 2
				}
				return &game.StackItem{
					Kind:         game.StackItemTriggered,
					Controller:   ev.Actor,
					Owner:        ev.Actor,
					SourceCardID: source.InstanceID,
					Label:        "Hydroid Krasis — gain half X life and draw half X cards",
					Effect: func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						if err := (GainLife{Player: item.Controller, Amount: half}).Apply(ctx); err != nil {
							return err
						}
						return DrawCards{Player: item.Controller, N: half}.Apply(ctx)
					},
				}
			},
		}},
	})
}
