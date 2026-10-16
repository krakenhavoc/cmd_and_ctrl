package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Touch the Spirit Realm — Enchantment {2}{W} (the Squall deck
// request, #3918):
//
//	"When this enchantment enters, exile up to one target artifact or
//	 creature until this enchantment leaves the battlefield.
//	 Channel — {1}{W}, Discard this card: Exile target artifact or
//	 creature. Return it to the battlefield under its owner's control
//	 at the beginning of the next end step."
//
// The entry trigger is the Oblivion Ring family's (Static Prison):
// a targeted trigger whose exile lasts until this permanent leaves
// (CR 610.3), with "up to one" letting the trigger target nothing.
//
// Channel is Eiganjo's shape, an activated ability from the hand whose
// cost discards the card. Its body is Waterbender's Restoration's: exile
// now, and a delayed trigger (CR 603.7) returns the card under its
// owner's control at the next end step. A token exiled this way does not
// come back (CR 111.8).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "52f0dcbd-ccf3-4525-8e55-2d38b4991834",
		Name:         "Touch the Spirit Realm",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: Self,
			Targets:   TargetPermanent("up to one target artifact or creature", Or(Artifact(), Creature())).WithCount(0, 1),
			Key:       "Touch the Spirit Realm — exile up to one target artifact or creature until it leaves",
			Effect:    exileChosenTargetUntilThisLeaves("Touch the Spirit Realm — the exiled card returns when Touch the Spirit Realm leaves the battlefield"),
		}},
		Activated: []ActivatedAbility{{
			Label:   "Channel — {1}{W}, Discard this card: Exile target artifact or creature. Return it to the battlefield under its owner's control at the beginning of the next end step.",
			Cost:    game.AbilityCost{Mana: "{1}{W}", DiscardSelf: true},
			Zones:   []game.ZoneKind{game.ZoneHand},
			Targets: TargetPermanent("target artifact or creature", Or(Artifact(), Creature())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return exileTargetsThenScheduleReturn(NewContext(g, item),
					"Touch the Spirit Realm — return the exiled card")
			},
		}},
	})
}
