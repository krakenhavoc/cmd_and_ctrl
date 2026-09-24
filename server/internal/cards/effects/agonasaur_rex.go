package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Agonasaur Rex — Creature — Dinosaur {3}{G}{G}, 8/8 (EDHREC rank
// 4528):
//
//	"Trample
//	 Cycling {2}{G} ({2}{G}, Discard this card: Draw a card.)
//	 When you cycle this card, put two +1/+1 counters on up to one
//	 target creature or Vehicle. It gains trample and indestructible
//	 until end of turn."
//
// An 8/8 trampler for five that is never a dead card: cycle it early
// for a card and a combat trick, cast it late as a threat. The
// flexibility is why it shows up in Dinosaur and stompy lists that
// would not otherwise play a five-mana vanilla body.
//
// #660 shipped the two things that used to be missing —
// game.AbilityCost.DiscardSelf and ActivatedAbilityShape.Zones —
// so cycling itself is Cycling("{2}{G}") like every other cycling
// card. The "when you cycle this card" trigger fires from the
// GRAVEYARD (CR 702.29c — the card is already there by the time the
// ability triggers), the same shape Magmakin Artillerist's own cycle
// trigger uses: InGraveyard(On(game.EventCycle, Self, …)). It carries
// a real target clause — "up to one target creature or Vehicle" is
// Min 0, so the trigger still goes on the stack with nothing chosen
// and its Effect no-ops.
//
// The two counters are a permanent Layer 7d effect (AddCounter); the
// trample and indestructible are turn-scoped Layer 6 grants
// (GrantKeywordUntilEOT) — two primitives because they are two kinds
// of thing, not two lines of the same one.
//
// The body half is complete: trample rides PrintedKeywords and feeds
// the combat engine's overflow math and the CR 510.1c damage-
// assignment prompt.
func init() {
	Register(Spec{
		OracleID:        "1ad5766b-9ae3-437b-bd91-c4c988a8b095",
		Name:            "Agonasaur Rex",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated:       []ActivatedAbility{Cycling("{2}{G}")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				InGraveyard(On(game.EventCycle, Self,
					"Agonasaur Rex — cycled: two +1/+1 counters, trample and indestructible until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						ts := ctx.LegalTargets()
						if len(ts) == 0 {
							return nil
						}
						id := ts[0].ID
						if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
							return err
						}
						return GrantKeywordUntilEOT{
							Target:   id,
							Keywords: []string{"trample", "indestructible"},
							Label:    "Agonasaur Rex — trample and indestructible",
						}.Apply(ctx)
					})),
				TargetPermanent("up to one target creature or Vehicle", Or(Creature(), Subtype("Vehicle"))).WithCount(0, 1),
			),
		},
	})
}
