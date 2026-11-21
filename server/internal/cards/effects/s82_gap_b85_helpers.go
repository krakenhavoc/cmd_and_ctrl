package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gapB85YouCastFromOutsideYourHand is the Paradox condition (Doctor Who):
// the source's controller cast a spell from anywhere other than their hand.
// The zone is the one the spell left.
func gapB85YouCastFromOutsideYourHand(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventCast && ev.Actor == source.Controller &&
		ev.OldZone != "" && ev.OldZone != game.ZoneHand
}

// gapB85DoublePlusOneCountersOnThis is "double the number of +1/+1 counters
// on this creature": as many more as it has as the ability resolves.
func gapB85DoublePlusOneCountersOnThis(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	c, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || ctx.isNewSourceObject(item.SourceCardID) || !onBattlefield(g, item.SourceCardID) {
		return nil
	}
	n := c.Counters[game.CounterPlusOne]
	if n <= 0 {
		return nil
	}
	return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
}

// gapB85HistoricPermanentYouControlEnters is Displaced Dinosaurs' entry
// condition: a historic permanent (an artifact, a legendary, a Saga) is
// entering the battlefield under the source's controller's control.
func gapB85HistoricPermanentYouControlEnters(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
	if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield || src == nil {
		return false
	}
	entering, ok := g.LookupCardForEffect(ev.CardID)
	return ok && entering.Controller == src.Controller && b09IsHistoric(entering)
}

// gapB85PlusOneOnEachHuman is "put a +1/+1 counter on each Human you
// control": the set is snapshotted first.
func gapB85PlusOneOnEachHuman(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	// "Human" is a creature-type word of the source's text (CR 612.2, #3984).
	human := ctx.TextSubtype("Human")
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsCreature() && c.HasSubtype(human) {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
			continue
		}
		if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}

// rassilonUpkeep is Rassilon's upkeep trigger: the controller loses 2 life,
// then exiles the top card of their library and may play it for as long as it
// remains exiled.
func rassilonUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (GainLife{Amount: -2}).Apply(ctx); err != nil {
		return err
	}
	return ExileTopWithPermission{From: item.Controller, GrantTo: item.Controller, N: 1, WhileExiled: true}.Apply(ctx)
}

// gapB85NoncreatureSpellFromExile is Rassilon's grant filter: a noncreature
// spell being cast from exile. The optional costs are offered while the card
// still sits in exile, so the zone is read where the card is now.
func gapB85NoncreatureSpellFromExile(g *game.Game, caster uuid.UUID, spell game.Card) bool {
	if spell.IsCreature() {
		return false
	}
	z := g.FindCardZoneForEffect(spell.InstanceID)
	return z != nil && z.Kind == game.ZoneExile
}

// gapB85ETBWithCastX is "When this enters, <effect with X>" for a permanent cast with
// X (CR 107.3m): X is read off the permanent as the trigger is built and
// carried on the item as Params.Amount. A copy or a cheat onto the battlefield
// has X = 0.
func gapB85ETBWithCastX(label string, effect Effect) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches:   []game.EventKind{game.EventETB},
		AppliesTo: Self,
		Key:       label,
		Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label)
			item.Params.Amount = source.CastX()
			return item
		},
		Effect: effect,
	}
}

// gapB85SkulkGrantPurpose is The Master, Mesmerist's activated ability as the
// planner reads it: the target gains skulk until end of turn (a benefit to it,
// #4302) and is goaded.
func gapB85SkulkGrantPurpose() game.Purpose {
	p := ForTargets(game.TargetPurpose{Slot: 0, UntilEndOfTurn: game.TargetBenefit, Pump: &game.Pump{Keywords: []string{"skulk"}}})
	p.Answers = game.AnswerCombatGrant
	return p
}
