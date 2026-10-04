package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vincent Valentine // Galian Beast — transforming double-faced card,
// Legendary Creature — Assassin {2}{B}{B}, 2/2 // Legendary Creature —
// Werewolf Beast, 3/2:
//
//	Vincent Valentine:
//	"Whenever a creature an opponent controls dies, put a number of
//	 +1/+1 counters on Vincent Valentine equal to that creature's
//	 power.
//	 Whenever Vincent Valentine attacks, you may transform it."
//	Galian Beast:
//	"Trample, lifelink
//	 When Galian Beast dies, return it to the battlefield tapped (front
//	 face up)."
//
// The counters are the dead creature's power as it last stood on the
// battlefield (CR 603.10a, via TriggeringPermanent), counters and
// pumps included, never below zero, and they go on Vincent only if he
// is still the permanent that triggered. The attack trigger is an
// ordinary "you may" with the transform helper, so a Vincent that
// already transformed (the trigger watches only the front face's
// ability) does nothing more. The back face is its own catalog row
// (oracle ID + "#1"): a card leaving the battlefield is front face up
// in every other zone (CR 711.8), so the dies trigger returns the
// front, Vincent, tapped, and he keeps nothing from the beast.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     vincentValentineOracleID,
		Name:         "Vincent Valentine",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, b33OpponentsCreatureDiedWhen,
				"Vincent Valentine — put +1/+1 counters equal to that creature's power on it",
				vincentValentineGrowFromThePower),
			Optional(
				On(game.EventAttack, ThisAttacked, "Vincent Valentine — you may transform it", vincentValentineTransform),
				"Vincent Valentine — transform it?"),
		},
	})

	Register(Spec{
		OracleID:        vincentValentineOracleID + "#1",
		Name:            "Galian Beast",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "lifelink"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Galian Beast — return it to the battlefield tapped, front face up", galianBeastReturns),
		},
	})
}

const vincentValentineOracleID = "f40e6bbf-1fed-4ec5-869d-18c2dd396d14"

func b33OpponentsCreatureDiedWhen(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return b33OpponentsCreatureDied(ev, source, g)
}

// vincentValentineGrowFromThePower is the first trigger's body.
func vincentValentineGrowFromThePower(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	info, ok := ctx.TriggeringPermanent()
	if !ok || info.Power <= 0 {
		return nil
	}
	if g.AbilitySourceGoneForEffect(item) {
		return nil
	}
	return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: info.Power}.Apply(ctx)
}

// vincentValentineTransform is "you may transform it", the Garruk
// guard: only the permanent that attacked, and only while it is still
// on the battlefield with its front face up.
func vincentValentineTransform(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	c, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || !onBattlefield(g, item.SourceCardID) || c.ActiveFace != 0 || ctx.isNewSourceObject(item.SourceCardID) {
		return nil
	}
	return TransformThis{}.Apply(ctx)
}

// galianBeastReturns is "return it to the battlefield tapped (front
// face up)": the card in the graveyard is Vincent again, and it comes
// back tapped. Nothing to do when something moved the card away in
// response.
func galianBeastReturns(g *game.Game, item *game.StackItem) error {
	id := item.SourceCardID
	if id == uuid.Nil {
		return nil
	}
	if zone := g.FindCardZoneForEffect(id); zone == nil || zone.Kind != game.ZoneGraveyard {
		return nil
	}
	return ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield, Tapped: true}.Apply(NewContext(g, item))
}
