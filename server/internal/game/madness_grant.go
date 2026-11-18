package game

// madness_grant.go — madness granted by a permanent rather than printed on
// the card (Falkenrath Gorger, CR 702.35).
//
// A card's own madness is two abilities it has in hand (see madness.go).
// "Each Vampire creature card you own that isn't on the battlefield has
// madness" gives those same two abilities to cards that print none, so the
// permanent that grants it carries them: a replacement over the discard
// (CR 702.35a, "that player discards it, but exiles it instead of putting
// it into their graveyard") and a triggered ability ("when this card is
// exiled this way, its owner may cast it by paying [cost] rather than
// paying its mana cost"). The offer is madness's own (madnessOfferBody), so
// the cast, the decline to the graveyard and the restore point are shared.
//
// `grants` says which discarded cards have the granted madness, judged on
// the card and the permanent that grants it.

// GrantedMadnessReplacement is CR 702.35a's replacement for a card whose
// madness a permanent grants. It applies while the discard is still headed
// for the graveyard (CR 614.5), so a card that already has its own madness,
// or that another replacement moved, is not rewritten twice.
func GrantedMadnessReplacement(grants func(card, source *Card) bool) ReplacementEffect {
	r := MadnessReplacement()
	r.AppliesTo = func(ev *ReplacementEvent, g *Game, src *Card) bool {
		if ev.Kind != RepEventDiscard || ev.NewZone != ZoneGraveyard || src == nil {
			return false
		}
		c := g.findCardByIDLocked(ev.CardID)
		return c != nil && grants(c, src)
	}
	return r
}

// GrantedMadnessTrigger is CR 702.35a's second half for a granted madness:
// the discarded card went from a hand to exile, and its owner may cast it
// for its mana cost. The offer's cost is the card's own mana cost, "equal to
// its mana cost"; a card with none is never offered (the cost is not one a
// player could pay). The offer is read as the ability resolves, against the
// card as it is then in exile (CR 608.2), so a card that has left exile since
// is offered nothing.
func GrantedMadnessTrigger(grants func(card, source *Card) bool) TriggeredAbility {
	return TriggeredAbility{
		Watches: []EventKind{EventDiscardCard},
		Key:     MadnessTriggerLabel,
		AppliesTo: func(ev Event, source *Card, _ Characteristic, g *Game) bool {
			if source == nil || ev.OldZone != ZoneHand || ev.NewZone != ZoneExile {
				return false
			}
			c := exiledCardByIDLocked(g, ev.CardID)
			return c != nil && c.ManaCost != "" && grants(c, source)
		},
		Effect: func(g *Game, item *StackItem) error {
			c := exiledCardByIDLocked(g, item.Trigger.Event.CardID)
			if c == nil {
				return nil
			}
			return g.offerMadnessCastLocked(c.Owner, c.InstanceID, c.ManaCost)
		},
	}
}
