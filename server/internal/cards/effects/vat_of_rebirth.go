package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vat of Rebirth — Artifact {B}:
//
//		"Whenever another artifact or creature you control is put into a
//		 graveyard from the battlefield, put an oil counter on this
//		 artifact.
//		 {2}{B}, {T}, Remove four oil counters from this artifact: Return
//		 target creature card from your graveyard to the battlefield.
//		 Activate only as a sorcery."
//
//	  - The trigger is one EventLTB per permanent put into a graveyard,
//	    and "artifact or creature" is one occurrence, never two: an
//	    artifact creature adds one counter. Both types and the
//	    controller are read as the permanent last existed (CR 603.10a,
//	    leftAsType and leftUnderControlOf), so a Vehicle that was only an
//	    artifact when it died, or a creature an opponent controlled, are
//	    judged by what they were on the battlefield. Tokens count: they
//	    are put into a graveyard before they cease to exist. "Another"
//	    leaves out the Vat itself.
//	  - The counter lands only while the Vat is still the permanent that
//	    triggered (CounterOnThis).
//	  - The activation pays {2}{B}, the tap and exactly four oil counters
//	    at announce, so a Vat with three cannot be activated and one with
//	    five keeps one. Sorcery speed is the row's SorcerySpeed.
//	  - The target is a creature card in YOUR graveyard, re-checked as the
//	    ability resolves, and it returns under your control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "10c31317-71e8-42e0-85e0-3e64bd0c3dd3",
		Name:         "Vat of Rebirth",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, vatOfRebirthCondition,
				"Vat of Rebirth — put an oil counter on it",
				Do(CounterOnThis{Kind: vatOfRebirthCounter, N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label: "{2}{B}, {T}, Remove four oil counters from this artifact: Return target creature card from your graveyard to the battlefield. Activate only as a sorcery.",
			Cost: Plus(ManaCost("{2}{B}"), TapCost(),
				RemoveCountersFromThis(vatOfRebirthCounter, 4)),
			Targets:      TargetCardInGraveyard("target creature card from your graveyard", Creature(), YouOwn()),
			SorcerySpeed: true,
			Effect:       returnFirstLegalGraveyardTargetToBattlefield,
		}},
	})
}

const vatOfRebirthCounter = "oil"

// vatOfRebirthCondition is "another artifact or creature you control is
// put into a graveyard from the battlefield".
func vatOfRebirthCondition(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard || ev.CardID == source.InstanceID {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok {
		return false
	}
	return (leftAsType(ev, c, "artifact") || leftAsType(ev, c, "creature")) && leftUnderControlOf(ev, c) == source.Controller
}
