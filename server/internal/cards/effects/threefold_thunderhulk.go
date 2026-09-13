package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Threefold Thunderhulk — Artifact Creature — Gnome {7}, 0/0 (EDHREC
// rank 1850):
//
//	"This creature enters with three +1/+1 counters on it.
//	 Whenever this creature enters or attacks, create a number of 1/1
//	 colorless Gnome artifact creature tokens equal to its power.
//	 {2}, Sacrifice another artifact: Put a +1/+1 counter on this
//	 creature."
//
// The artifact deck's go-wide engine: three Gnomes on entry, more
// every attack, and each Gnome fed to it grows the next batch. The
// counters are the CR 614 self-replacement (b10EntersWithCounters),
// on the card before the ETB trigger reads its power. "Enters or
// attacks" is one ability with two conditions (Sun Titan's shape),
// and "its power" is read at resolution — counters, anthems and all
// — so a response that shrinks it makes fewer Gnomes, as printed.
// A Thunderhulk killed in response makes none: the printed card
// would use its last-known power, but a trigger from a live
// permanent has no LKI to read, so the graveyard's printed 0 is
// what it sees. Weaker, never stronger.
// The sacrifice cost is "another artifact": the Thunderhulk is an
// artifact itself and excludes itself by name (the Warren Soultrader
// shape); a second Thunderhulk cannot be fed to the first, the one
// declared retreat.
func init() {
	Register(Spec{
		OracleID:     "b8020a8b-557b-465d-865d-b59fecd7abc1",
		Name:         "Threefold Thunderhulk",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The sacrifice cost won't accept another Threefold Thunderhulk — every other artifact you control is fine."},
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters("+1/+1", 3, "Threefold Thunderhulk: enters with three +1/+1 counters"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB, game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Threefold Thunderhulk — Gnomes equal to its power",
					func(g *game.Game, item *game.StackItem) error {
						self, ok := g.LookupCardForEffect(item.SourceCardID)
						if !ok {
							return nil
						}
						n := self.CurrentPower()
						if n <= 0 {
							return nil
						}
						return CreateToken{Controller: item.Controller, Template: b17GnomeToken(), N: n}.Apply(NewContext(g, item))
					})
			},
		}},
		Activated: []ActivatedAbility{{
			Label: "{2}, Sacrifice another artifact: Put a +1/+1 counter on Threefold Thunderhulk",
			Cost: Plus(ManaCost("{2}"), game.AbilityCost{
				SacrificeOther: sacrificeSpec("another artifact", Artifact(), b03NotNamed("Threefold Thunderhulk")),
			}),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
