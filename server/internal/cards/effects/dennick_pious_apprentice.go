package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dennick, Pious Apprentice // Dennick, Pious Apparition (#1885, ADR
// 0109 §6) — a disturb card.
//
// Front face, Legendary Creature — Human Soldier {W}{U}, 2/3:
//
//	"Lifelink
//	 Cards in graveyards can't be the targets of spells or abilities.
//	 Disturb {2}{W}{U} (You may cast this card from your graveyard
//	 transformed for its disturb cost.)"
//
// Back face, Legendary Creature — Spirit Soldier, 3/2:
//
//	"Flying
//	 Whenever one or more creature cards are put into graveyards from
//	 anywhere, investigate. This ability triggers only once each turn.
//	 (Create a Clue token. It's an artifact with "{2}, Sacrifice this
//	 token: Draw a card.")
//	 If Dennick would be put into a graveyard from anywhere, exile it
//	 instead."
//
// THE FRONT FACE'S STATIC is ADR 0109 §6's TargetingRestrictions, read at
// the engine's two targeting choke points, so no spell or ability may
// target a card in any graveyard while Dennick is on the battlefield —
// its controller's included — and one already aimed at a graveyard card
// loses that target (CR 601.2c, 608.2b). Dennick's own disturb is a cast
// from the graveyard, not a target, so it is untouched; so is a cost
// that exiles a card from a graveyard.
//
// THE BACK FACE'S TRIGGER watches every way a card reaches a graveyard —
// a death, a discard, a mill, a countered spell, any other move — and
// asks whether a creature card (a token is not a card, CR 108.2) is now
// there. "One or more" is OncePerBatch: a board wipe is one trigger.
// "Only once each turn" is the turn tally, read in AppliesTo, which gates
// the second batch of the turn out entirely. Disturb is effects.Disturb;
// the exile clause is DisturbedExile.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         dennickOracleID,
		Name:             "Dennick, Pious Apprentice",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"lifelink"},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Disturb("{2}{W}{U}")},
		TargetingRestrictions: []game.TargetingRestriction{
			CardsInGraveyardsCantBeTargeted("Cards in graveyards can't be the targets of spells or abilities."),
		},
	})
	Register(Spec{
		OracleID:        dennickOracleID + "#1",
		Name:            "Dennick, Pious Apparition",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(OnAny(
				[]game.EventKind{game.EventZoneMove, game.EventDiscardCard, game.EventMill, game.EventCounterSpell},
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return creatureCardPutIntoAGraveyard(ev, g) &&
						!b11TriggeredThisTurn(g, source.InstanceID, dennickInvestigateLabel)
				},
				dennickInvestigateLabel, Do(CreateToken{Template: ClueToken(), N: 1}))),
		},
		Replacements: []game.ReplacementEffect{DisturbedExile("Dennick, Pious Apparition")},
	})
}

const (
	dennickOracleID         = "45802eb2-6848-416c-95e0-c1c1ea0620d0"
	dennickInvestigateLabel = "Dennick, Pious Apparition — investigate"
)

// creatureCardPutIntoAGraveyard reports whether the event put a creature
// CARD into a graveyard, from anywhere. The engine emits one event per
// move (game/zone_route.go): a zone move (a death, a tutor, a spell
// resolving), a discard and a mill each name their destination, and a
// countered spell carries none, so its card is looked for where it is
// now. A card already in a graveyard that moves to another is not "put
// into a graveyard" from anywhere new, and a token is not a card
// (CR 108.2). The card is judged in the graveyard, by its printed types.
func creatureCardPutIntoAGraveyard(ev game.Event, g *game.Game) bool {
	if ev.CardID == uuid.Nil {
		return false
	}
	switch ev.Kind {
	case game.EventZoneMove, game.EventDiscardCard, game.EventMill:
		if ev.NewZone != game.ZoneGraveyard || ev.OldZone == game.ZoneGraveyard {
			return false
		}
	case game.EventCounterSpell:
		if ev.Target != ev.CardID {
			return false
		}
		if z := g.FindCardZoneForEffect(ev.CardID); z == nil || z.Kind != game.ZoneGraveyard {
			return false
		}
	default:
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && !c.IsToken() && c.IsCreature()
}
