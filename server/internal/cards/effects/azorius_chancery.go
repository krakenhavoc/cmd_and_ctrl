package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Azorius Chancery — Land:
//
//	"This land enters tapped. When this land enters, return a land
//	you control to its owner's hand. {T}: Add {W}{U}."
//
// A "bounce land": the tempo cost is real (it enters tapped AND
// sets you back a land drop), and the payoff is that one land taps
// for two coloured mana.
//
// Enters-tapped is a real CR 614 self-replacement (SelfEntersTapped),
// the same one the karoo table in bounce_lands.go uses: the land is
// never untapped on the battlefield and no tap event is emitted. This
// file used to say a catalog replacement could not fire on its own
// source's entry and tapped the land from a hook a beat later; that
// stopped being true with the Temple cycle's entering-card block in
// gatherActiveReplacementsLocked (#360, #578).
//
// Sandbox simplification: the printed bounce is a CHOICE, not a
// target — "return a land you control", with no "target" in the
// oracle text. It is modelled as a target clause so the controller
// gets the existing picker, which makes it targetable in response
// and locks the choice in at announce rather than resolution. The
// trigger can never fizzle for want of a legal choice, because the
// Chancery itself is always a land its controller controls — and
// returning itself is a normal, sometimes correct, line.
func init() {
	Register(Spec{
		OracleID:     "189fc8f4-17ac-4f1d-82c8-8401445bdaf4",
		Name:         "Azorius Chancery",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The land to bounce is picked as a target when the trigger goes on the stack, not on resolution."},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W}{U}",
			Label:    "Add {W}{U}",
		}},
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Targets: TargetPermanent("a land you control", And(Land(), YouControl())),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Azorius Chancery — return a land you control",
					func(g *game.Game, item *game.StackItem) error {
						if len(item.Targets) == 0 {
							return nil
						}
						return BounceToHand{Target: item.Targets[0].ID}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
