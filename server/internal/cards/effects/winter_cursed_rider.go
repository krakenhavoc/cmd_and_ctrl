package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Winter, Cursed Rider — Legendary Creature — Human Warlock {U}{B}, 3/2
// (Aetherdrift, #4464):
//
//	"Ward—Pay 2 life.
//	 Artifacts you control have "Ward—Pay 2 life."
//	 Exhaust — {2}{U}{B}, {T}, Exile X artifact cards from your graveyard:
//	 Each other nonartifact creature gets -X/-X until end of turn.
//	 (Activate each exhaust ability only once.)"
//
// Both wards are ward's own trigger (Hexing Squelcher's WardLife), the
// granted one through WardGranted over artifacts you control, read through
// the layers. Each is a separate instance (CR 702.21a), so if Winter is an
// artifact, as the ruling says, it has two and a spell that targets it
// triggers both. The exhaust's X appears only in the exile cost, so the
// activator announces it by naming that many artifact cards
// (game.ExileCost.CountIsX, ExileXFromGraveyardAsCost), and the effect
// reads it back as the number exiled (CR 107.3a). The -X/-X is read as the
// ability resolves, for each nonartifact creature other than Winter, and
// lasts until end of turn (CR 611.2c): a creature that arrives later is not
// shrunk. X = 0 is legal and does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "343833c1-19fd-4e8f-b82e-0f20903caaff",
		Name:         "Winter, Cursed Rider",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardLife(2), "Winter, Cursed Rider — ward, pay 2 life"),
			WardGranted(WardLife(2), "Winter, Cursed Rider — artifacts you control have ward, pay 2 life",
				func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.IsArtifact() && target.Controller == source.Controller
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "Exhaust — {2}{U}{B}, {T}, Exile X artifact cards from your graveyard: Each other nonartifact creature gets -X/-X until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, AmountIsX: true, Partial: true}},
			Exhaust: true,
			Cost: Plus(ManaCost("{2}{U}{B}"), TapCost(),
				ExileXFromGraveyardAsCost("X artifact cards", func(c game.Card) bool { return c.IsArtifact() })),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				x := len(ctx.Exiled())
				if x == 0 {
					return nil
				}
				source := item.SourceCardID
				return BoostUntilEOT{
					Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
						return c.IsCreature() && !c.IsArtifact() && c.InstanceID != source
					},
					Power:     -x,
					Toughness: -x,
					Label:     "Winter, Cursed Rider — -X/-X until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
