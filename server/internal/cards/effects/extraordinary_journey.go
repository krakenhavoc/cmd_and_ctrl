package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Extraordinary Journey — Enchantment {X}{X}{U}{U}:
//
//	"When this enchantment enters, exile up to X target creatures. For
//	 each of those cards, its owner may play it for as long as it
//	 remains exiled.
//	 Whenever one or more nontoken creatures enter, if one or more of
//	 them entered from exile or was cast from exile, you draw a card.
//	 This ability triggers only once each turn."
//
// Two abilities.
//
// The exile is Blitzball Stadium's shape for an "up to X" clause on an
// entering permanent: the target clause is built when the trigger is put
// on the stack, from the X the enchantment was cast with
// (Card.CastX, CR 107.3m), so a Journey that was blinked or reanimated
// has X of zero and exiles nothing. Each creature leaves for exile with
// an unbounded permission for its OWNER to PLAY it (game.CastPermission,
// WhileInZoneDuration — the grant is gone the moment the card leaves
// exile by any route, so a card exiled again later by something else
// does not inherit it). It is "play", not "cast": CastOnly stays off so
// a land-creature could be played as a land, and timing still applies on
// top of the permission (a creature is cast at sorcery speed unless it
// has flash). A token that was exiled is not a card, ceases to exist
// there (CR 111.7) and gets nothing. Targets that became illegal before
// resolution are skipped (CR 608.2b), not errors.
//
// The draw is Celes, Rune Knight's shape with the zone changed: a batch
// trigger ("one or more", CR 603.2c, OncePerBatch) whose condition is in
// AppliesTo, so the batch's slot is spent only by an event that passes
// it. "Entered from exile" is EventETB.EnteredFrom (a flicker coming
// back, a card put onto the battlefield from exile) and "was cast from
// exile" is Card.Provenance.FromZone (CR 400.7d), the zone the spell that
// became THIS object was cast from — a creature spell cast from exile
// enters from the STACK, so the first reading alone would miss the whole
// point of the card. The creature may be anyone's: the clause says
// "nontoken creatures enter", not "you control", and the card is drawn
// by Journey's controller. A token never qualifies. "Only once each turn"
// is the per-turn tally of this ability's own stack label
// (b11TriggeredThisTurn), the Exemplar of Light pattern: a trigger that
// was countered still counted, as printed.
//
// No simplification.
func init() {
	const drawLabel = "Extraordinary Journey — draw a card"
	enters := WhenThisEnters("Extraordinary Journey — exile up to X target creatures", extraordinaryJourneyExile)
	enters.TargetsFrom = UpToCastXCreatureTargets
	Register(Spec{
		OracleID:     "69288566-3015-45de-b391-6fe079865f0c",
		Name:         "Extraordinary Journey",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			enters,
			OncePerBatch(On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if b11TriggeredThisTurn(g, source.InstanceID, drawLabel) {
					return false
				}
				return extraordinaryJourneyEntrant(ev, g)
			}, drawLabel, Do(DrawCards{N: 1}))),
		},
	})
}

// extraordinaryJourneyEntrant is the second ability's trigger and its
// intervening if in one: a nontoken creature entered and it either came
// out of exile or was cast from exile.
func extraordinaryJourneyEntrant(ev game.Event, g *game.Game) bool {
	if ev.Kind != game.EventETB {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !c.IsCreature() || IsToken(c) {
		return false
	}
	return ev.EnteredFrom == game.ZoneExile || c.Provenance.FromZone == game.ZoneExile
}

// extraordinaryJourneyExile is the enters trigger's body: exile every
// target that is still legal, each with its owner's unbounded permission
// to play it.
func extraordinaryJourneyExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := g.ExileCardWithPermissionForEffect(t.ID, game.CastPermission{
			Duration: game.WhileInZoneDuration(),
		}); err != nil {
			return err
		}
	}
	return nil
}
