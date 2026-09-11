package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Craterhoof Behemoth — Creature — Beast, {5}{G}{G}{G}, 5/5:
//
//	"Haste"
//	"When this creature enters, creatures you control gain trample
//	 and get +X/+X until end of turn, where X is the number of
//	 creatures you control."
//
// The green finisher. Eight mana that converts a wide board into
// lethal on the spot, which is why every creature-based Commander
// deck that can cast it does.
//
// # Why this is writable now
//
// The batch-02 triage (#295) filed Craterhoof under "cost
// modification, until EOT". The cost-modification half was a
// misfiling — nothing on this card modifies a cost. The until-EOT
// half was real and is now closed: S32's turn-scoped continuous
// effects shipped BoostUntilEOT and GrantKeywordUntilEOT, and
// Craterhoof is Overrun with a computed X, on a body, as an ETB
// trigger.
//
// # X, and when it is measured
//
// X is the number of creatures you control, counted AS THE TRIGGER
// RESOLVES — with the Behemoth itself already on the battlefield and
// included. A five-creature board plus Craterhoof is X=6, so the
// Behemoth swings as an 11/11 and every other creature gets +6/+6.
// Counting at trigger time instead would miss a creature that
// entered in response.
//
// The same count is used for the snapshot and for the pump because
// they are one effect; the two primitives exist only because a P/T
// change and an ability grant genuinely live in different CR 613
// layers (7c and 6) and one registry entry cannot sort into both.
// Overrun's file states that reasoning at length.
//
// # CR 611.2c
//
// The affected set is snapshotted when the trigger resolves, so a
// creature cast afterwards this turn gets neither the pump nor the
// trample, and a creature flickered out and back loses both. That
// falls out of eotSnapshot and is the printed behaviour.
//
// # Trample and haste are both real here
//
// Both are among the twelve keywords the combat code honours:
// granted trample really does push excess damage through, and
// Craterhoof's printed haste really does let it attack the turn it
// lands — which is the entire point of the card.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "8c52bd39-0586-48ca-b263-17210cf9feb6",
		Name:            "Craterhoof Behemoth",
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Craterhoof Behemoth — trample and +X/+X",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						x := 0
						for _, c := range g.BattlefieldCardsForEffect() {
							if c.IsCreature() && c.Controller == item.Controller {
								x++
							}
						}
						if x == 0 {
							return nil
						}
						yours := And(Creature(), YouControl())
						if err := (BoostUntilEOT{
							Match:     yours,
							Power:     x,
							Toughness: x,
							Label:     "Craterhoof Behemoth — +X/+X",
						}).Apply(ctx); err != nil {
							return err
						}
						return GrantKeywordUntilEOT{
							Match:    yours,
							Keywords: []string{"trample"},
							Label:    "Craterhoof Behemoth — trample",
						}.Apply(ctx)
					})
			},
		}},
	})
}
