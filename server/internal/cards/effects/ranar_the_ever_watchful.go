package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ranar the Ever-Watchful — {2}{W}{U} Legendary Creature — Spirit
// Warrior 2/3:
//
//	"Flying, vigilance
//	 The first card you foretell each turn costs {0} to foretell.
//	 Whenever one or more cards are put into exile from your hand or a
//	 spell or ability you control exiles one or more permanents from
//	 the battlefield, create a 1/1 white Spirit creature token with
//	 flying."
//
// The trigger waited on #1320: nothing on a zone-change event said
// which spell or ability moved a permanent into exile, so "a spell or
// ability YOU CONTROL exiles" could not be told from an opponent's
// Swords to Plowshares on the same permanent, or from a sandbox drag.
// Events now carry the move's cause (game/move_cause.go), and this card
// reads it through game.ExiledBySpellOrAbilityOf.
//
// The two clauses, as the Oracle update and its rulings read them:
//
//   - "put into exile from your hand" does NOT care who or what did it —
//     foretell, a cost that exiles a card from hand (Force of Will's
//     pitch), madness, an opponent's effect. The card's owner is the
//     "your": every hand in this engine holds only its owner's cards.
//     A madness discard into exile arrives as EventDiscardCard rather
//     than EventZoneMove, which is why both kinds are watched.
//   - "a spell or ability you control exiles one or more permanents"
//     cares who controlled the spell or ability, and not who controlled
//     the permanent: exiling an opponent's creature, or a token,
//     counts. A permanent exiled to pay a COST is not counted — the
//     engine records a cost as its own cause, and whether paying a cost
//     is the ability "exiling" is not settled by any ruling; not
//     counting it is the weaker reading.
//
// "Only once for each time", however many cards: OncePerBatch, since a
// batch is one resolution's worth of events (CR 603.2c). #1341 closed
// the last gap in that reading: a special action (foretell, suspend,
// turning a permanent face up) is CR 116.2's own instance of "play
// moving on", so PerformSpecialAction now opens its own event batch
// and two foretells taken back to back — nothing resolving between
// them — are two occurrences, as CR 603.2c asks, not one.
//
// The foretell discount is #1319's proof card: SpecialActionCostsLess
// partitions the CR 601.2f pass a third way (CostModifier.SpecialActions,
// beside .Activations for CR 602 and the bare default for a cast), and
// TheFirstOneThisTurn reads the per-turn tally
// (Game.ForetoldThisTurn / ForetoldCountThisTurn) that
// PerformSpecialAction now bumps once a foretold card actually lands
// in exile. ASpecialActionOfKind(SpecialActionForetell) is what keeps
// the clause off suspend and turn_face_up, which this card doesn't
// even print but a shared predicate has to get right regardless.
//
// Both printed simplifications are closed as of #1341 and #1319, so
// this card carries no caveat. One general engine limitation remains,
// tracked at the engine level rather than here (ADR 0013 "Not closed
// here", docs/engine-seams.md): a destruction that a replacement turns
// into an exile (Rest in Peace) carries no cause, so it does not count
// as "a spell you control exiled it". That is a gap in the move-cause
// system every reader of it shares, not a Ranar-specific simplification.
func init() {
	Register(Spec{
		OracleID:        "c73a9939-0742-4919-94d6-c3b537697f17",
		Name:            "Ranar the Ever-Watchful",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		CostModifiers: []game.CostModifier{
			SpecialActionCostsLess(2, "The first card you foretell each turn costs {0} to foretell.",
				ASpecialActionOfKind(game.SpecialActionForetell), TheFirstOneThisTurn()),
		},
		Triggered: []game.TriggeredAbility{{
			OncePerBatch: true,
			Watches:      []game.EventKind{game.EventZoneMove, game.EventDiscardCard},
			AppliesTo:    ranarSawAnExile,
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Ranar the Ever-Watchful — create a 1/1 Spirit",
					Do(CreateToken{Template: TokenCard("1/1 white Spirit with flying"), N: 1}))
			},
		}},
	})
}

// ranarSawAnExile is the trigger condition: a card of yours put into
// exile from your hand, or a permanent exiled from the battlefield by a
// spell or ability you control.
func ranarSawAnExile(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.NewZone != game.ZoneExile {
		return false
	}
	switch ev.OldZone {
	case game.ZoneHand:
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && c.Owner == source.Controller
	case game.ZoneBattlefield:
		return game.ExiledBySpellOrAbilityOf(ev, source.Controller)
	}
	return false
}
