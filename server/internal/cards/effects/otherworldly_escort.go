package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Otherworldly Escort — Creature — Human Detective {3}{W}, 4/3 (Murders at
// Karlov Manor Commander):
//
//	"Flash
//	 When this creature dies, if it's not a Spirit, return it to the
//	 battlefield under its owner's control with four charge counters on it.
//	 It's a Spirit Detective. (It's no longer a Human.)
//	 {1}{W}, {T}, Remove a charge counter from this creature: Destroy target
//	 creature that dealt damage to you this turn."
//
// The dies trigger's intervening "if" (CR 603.4) reads the creature's
// last-known characteristics (CR 603.10): a Spirit that dies stays dead.
// The return is one entry that carries the four counters (so a counter
// replacement sees them) and the type overwrite, which replaces the
// creature types with Spirit Detective (ADR 0210, CR 205.3d) from the moment
// it lands. The card must still be that object in the graveyard
// (CR 400.7, CR 608.2b). The ability's target is a creature that has dealt
// damage to its controller this turn (DealtDamageToYouThisTurn).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c0ebc1df-51ea-413b-bb88-7666a53ad783",
		Name:            "Otherworldly Escort",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, lki game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source) && !keywordSliceContains(lki.Subtypes, "Spirit")
			}, "Otherworldly Escort — return it to the battlefield with four charge counters; it's a Spirit Detective", otherworldlyEscortReturn),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}, {T}, Remove a charge counter from this creature: Destroy target creature that dealt damage to you this turn.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    Plus(ManaCost("{1}{W}"), TapCost(), RemoveCountersFromThis("charge", 1)),
			Targets: TargetCreature("target creature that dealt damage to you this turn", DealtDamageToYouThisTurn()),
			Effect:  destroyFirstLegalTarget,
		}},
	})
}

func otherworldlyEscortReturn(g *game.Game, item *game.StackItem) error {
	z := g.FindCardZoneForEffect(item.SourceCardID)
	if z == nil || z.Kind != game.ZoneGraveyard || !sameGraveyardObject(z, item) {
		return nil
	}
	return putCardsFromZoneTogetherWith(NewContext(g, item), game.ZoneGraveyard, []uuid.UUID{item.SourceCardID},
		game.ZoneEntryOptions{
			EntersWithCounters: map[string]int{"charge": 4},
			EntersAs: &game.EntersAs{
				AddTypes:      []string{"Creature"},
				CreatureTypes: []string{"Spirit", "Detective"},
				Label:         "Otherworldly Escort — it's a Spirit Detective",
			},
		}, game.TailRef{}, 0)
}
