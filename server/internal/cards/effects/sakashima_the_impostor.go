package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sakashima the Impostor — "You may have Sakashima the Impostor
// enter as a copy of any creature on the battlefield, except its
// name is Sakashima the Impostor, it's legendary in addition to its
// other types, and it has '{2}{U}{U}: Return Sakashima the Impostor
// to its owner's hand at the beginning of the next end step.'"
//
// The name exception is the one that proves the model. A copy takes
// the copied card's NAME along with everything else — that is what
// makes a Clone of your own commander die to the CR 704.5j legend
// rule — and Sakashima is printed the way it is precisely to dodge
// that. Keeping the name means the copy is a different legendary
// permanent from the thing it copied, so both survive; the added
// supertype is what puts it under the legend rule against a SECOND
// Sakashima.
//
// The granted activated ability was this card's declared
// simplification until #665: a copy effect replaces the flat printed
// fields with the copied card's, and the catalog's own activated
// abilities key on oracle ID, which after the copy is the copied
// card's — so there was nowhere for a granted ability to live. There
// is now. The bundle below is static catalog data; the except clause
// names it, and only the NAME rides in the copiable values, which is
// what makes it copiable again (CR 707.9a) and snapshot-safe.
const sakashimaReturnGrant = "sakashima-the-impostor/return-at-end"

func init() {
	Register(Spec{
		OracleID:     "a7243d25-22a2-4df5-adaf-1f40f5330ec1",
		Name:         "Sakashima the Impostor",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:  sakashimaReturnGrant,
			Text: "{2}{U}{U}: Return this creature to its owner's hand at the beginning of the next end step.",
			Activated: []ActivatedAbility{{
				Label: "{2}{U}{U}: Return this creature to its owner's hand at the beginning of the next end step.",
				Cost:  ManaCost("{2}{U}{U}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return ScheduleDelayedTrigger{
						Label:  "Sakashima the Impostor — return it to its owner's hand",
						Cards:  []uuid.UUID{item.SourceCardID},
						Effect: sakashimaReturnListedPermanentsToHand,
					}.Apply(NewContext(g, item))
				},
			}},
		}},
		Replacements: []game.ReplacementEffect{
			EntersAsCopyOf(
				"Sakashima the Impostor",
				anyCreatureOnBattlefield,
				func(_ *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
					v.SetName("Sakashima the Impostor")
					v.AddSupertype("Legendary")
					v.GrantAbility(sakashimaReturnGrant)
				},
			),
		},
	})
}

// sakashimaReturnListedPermanentsToHand is the delayed trigger's
// body: return the permanents named in the payload to their owners'
// hands, skipping any that has left the battlefield in the meantime
// (CR 603.7 — the delayed ability still triggers, and then does as
// much as it can).
//
// Reads the IDs off item.Targets rather than closing over them,
// which is the contract ScheduleDelayedTrigger documents and what
// keeps the trigger correct across an undo. Lives here rather than
// in helpers.go because Sakashima is its only caller; promote it the
// day a second card wants it.
func sakashimaReturnListedPermanentsToHand(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard {
			continue
		}
		if z := g.FindCardZoneForEffect(t.ID); z == nil || z.Kind != game.ZoneBattlefield {
			continue
		}
		if err := (BounceToHand{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
