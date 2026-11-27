package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b666_helpers.go — S82 OLD batch 666 (Invasion): Teferi's Moat's chosen-colour attack bar and Temporal
// Distortion's hourglass counters.

// OldB666HourglassCounter is Temporal Distortion's counter kind. An on-disk
// identity: never renamed.
const OldB666HourglassCounter = "hourglass"

// OldB666CreaturesOfTheChosenColorWithoutFlyingCantAttackYou is Teferi's
// Moat's "Creatures of the chosen color without flying can't attack you"
// (CR 508.1c): the player alone, not their planeswalkers or battles. The
// colour is the one stored on the printing permanent, read as the layer pass
// runs; before a colour is chosen nothing is barred.
func OldB666CreaturesOfTheChosenColorWithoutFlyingCantAttackYou(label string) game.StaticAbility {
	return game.StaticAbility{
		Label: label,
		Layer: game.Layer6Ability,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.IsCreature() && source != nil && source.ChosenColor != ""
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			if source == nil || source.ChosenColor == "" {
				return
			}
			c.AttackTargetRestrictions = append(c.AttackTargetRestrictions, game.AttackTargetRestriction{
				Source:        source.InstanceID,
				SourceName:    source.Name,
				NotPlayer:     source.Controller,
				NotPlayerOnly: true,
				OnlyAttackersMatching: []game.PermanentQuery{{
					Types: []string{"creature"}, Colors: []string{source.ChosenColor}, WithoutKeyword: "flying",
				}},
			})
		},
	}
}

// OldB666HourglassOnTap is Temporal Distortion's "Whenever a creature or land
// becomes tapped, put an hourglass counter on it": a tap (EventTapCard) or an
// attack that tapped the attacker (EventAttack; vigilance taps nothing).
func OldB666HourglassOnTap(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || item.Trigger.Event.CardID == uuid.Nil {
		return nil
	}
	id := item.Trigger.Event.CardID
	if !OnBattlefield(g, id) {
		return nil
	}
	return AddCounter{Target: id, Kind: OldB666HourglassCounter, N: 1}.Apply(NewContext(g, item))
}

// OldB666PermanentBecameTappedCreatureOrLand is the condition of that trigger.
func OldB666PermanentBecameTappedCreatureOrLand(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !(c.IsCreature() || c.IsLand()) {
		return false
	}
	switch ev.Kind {
	case game.EventTapCard:
		return true
	case game.EventAttack:
		return c.Tapped
	}
	return false
}

// OldB666HasHourglassCounter is Temporal Distortion's "each permanent with an
// hourglass counter on it".
func OldB666HasHourglassCounter() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.Counters[OldB666HourglassCounter] > 0 }
}

// OldB666RemoveHourglassCountersOfTheActivePlayer is "At the beginning of each
// player's upkeep, remove all hourglass counters from permanents that player
// controls".
func OldB666RemoveHourglassCountersOfTheActivePlayer(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	player := item.Trigger.Event.Actor
	ctx := NewContext(g, item)
	type held struct {
		id uuid.UUID
		n  int
	}
	var all []held
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.Counters[OldB666HourglassCounter] > 0 {
			all = append(all, held{c.InstanceID, c.Counters[OldB666HourglassCounter]})
		}
	}
	for _, h := range all {
		if err := (AddCounter{Target: h.id, Kind: OldB666HourglassCounter, N: -h.n}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
