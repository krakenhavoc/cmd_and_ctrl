package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Anticausal Vestige — Creature — Eldrazi {6}, 7/5:
//
//	"When this creature leaves the battlefield, draw a card, then you
//	 may put a permanent card with mana value less than or equal to
//	 the number of lands you control from your hand onto the
//	 battlefield tapped.
//	 Warp {4} (You may cast this card from your hand for its warp
//	 cost. Exile this creature at the beginning of the next end step,
//	 then you may cast it from exile on a later turn.)"
//
// The card #324 was filed against ("Warp is not working"): it was not
// in the catalog, so the warp offer never appeared. Warp itself is the
// shared constructor — the {4} price, the CR 603.7 end-step exile and
// the "on a later turn" cast-from-exile grant all ride Warp("{4}"),
// exactly as on Bygone Colossus.
//
// The two halves meet on purpose: warp's end-step exile IS a
// leaves-the-battlefield, so a warped Vestige draws and cheats a
// permanent in on the turn it is cast, and again whenever it leaves
// for real later. On(EventLTB, Self, …) fires on every exit — dying,
// exile, bounce — not the dies-only gate.
//
// "Draw a card, THEN you may put …": the draw happens first, so the
// card just drawn is a candidate. The land count is read after the
// draw, as the ability resolves; nothing can change it while the pick
// is open, so it is fixed into the candidate filter. The pick is a
// "you may" (declining is a real answer), it is any PERMANENT card
// (PutFromHandOntoBattlefield drops nonpermanents whatever the filter
// says, CR 110.4), and the permanent arrives tapped through the entry
// event rather than being tapped afterwards. It is a put, not a cast
// or a land play (CR 305.4), so a land put this way does not use the
// turn's land drop.
//
// "You" is the controller of the trigger — the player who controlled
// the Vestige as it left.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "aceea999-90ab-472c-86a7-48af1542cbcf",
		Name:             "Anticausal Vestige",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Warp("{4}")},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, Self,
				"Anticausal Vestige — draw a card, then you may put a permanent card onto the battlefield",
				anticausalVestigeLeaves),
		},
	})
}

// anticausalVestigeLeaves is the leaves-the-battlefield body: draw,
// then offer the tapped put, bounded by the lands the controller
// controls once the draw is done.
func anticausalVestigeLeaves(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
		return err
	}
	lands := b02CountLandsControlledBy(g, item.Controller)
	return PutFromHandOntoBattlefield{
		Player:   item.Controller,
		Match:    And(Permanent(), ManaValueLE(lands)),
		Tapped:   true,
		Optional: true,
		Label: "Anticausal Vestige — you may put a permanent card with mana value " +
			"less than or equal to the number of lands you control onto the battlefield tapped",
	}.Apply(ctx)
}
