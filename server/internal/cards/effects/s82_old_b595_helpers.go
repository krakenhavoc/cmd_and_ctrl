package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b595_helpers.go — shared bodies of the Betrayers of Kamigawa creatures (S82 OLD
// batch 595, #5602).

// OldB595CantAttackUnlessYouHaveCastACreatureSpell is "This creature can't attack unless
// you've cast a creature spell this turn" (Goblin Cohort; CR 508.1c): a fact about the
// creature's controller read live at declaration, so it refuses every target while unmet.
func OldB595CantAttackUnlessYouHaveCastACreatureSpell() game.StaticAbility {
	return selfAttackTargetRestriction(game.AttackTargetRestriction{ControllerMustHaveCastCreature: true})
}

// OldB595TapThis is "tap this creature" as a trigger's effect (Ogre Recluse).
func OldB595TapThis(g *game.Game, item *game.StackItem) error {
	return TapTarget{Target: item.SourceCardID}.Apply(NewContext(g, item))
}

// OldB595ProtectionFromColorsAmongYourPermanents is "This creature has protection from
// each color among permanents you control" (Empty-Shrine Kannushi): one protection
// quality per colour, read as the layer pass reaches layer 6 (so after colours have
// settled), so it follows the board. The creature counts among the permanents.
func OldB595ProtectionFromColorsAmongYourPermanents() game.StaticAbility {
	return game.StaticAbility{
		Label: "This creature has protection from each color among permanents you control.",
		Layer: game.Layer6Ability,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.InstanceID == source.InstanceID
		},
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			seen := map[string]bool{}
			for _, p := range g.BattlefieldCardsForEffect() {
				if p.Controller != source.Controller {
					continue
				}
				for _, col := range p.EffectiveColors() {
					seen[col] = true
				}
			}
			for _, col := range game.AllColors {
				if seen[col] {
					c.Abilities = game.AppendKeywordAbility(c.Abilities, game.ProtectionFromColor(col))
				}
			}
		},
	}
}

// OldB595LegendaryPermanentWasPutIntoAGraveyard is the condition of Yomiji's trigger:
// a legendary permanent other than the source was put into a graveyard from the
// battlefield, read from the permanent as it last existed there (CR 603.10a).
func OldB595LegendaryPermanentWasPutIntoAGraveyard(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard || ev.CardID == source.InstanceID {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && LeftAsSupertype(ev, c, "legendary")
}

// OldB595ReturnThatCardToItsOwnersHand is "return that card to its owner's hand": the card
// the event names, if it is still in a graveyard as the ability resolves. A token ceased
// to exist and is not a card.
func OldB595ReturnThatCardToItsOwnersHand(g *game.Game, item *game.StackItem) error {
	id := item.Trigger.Event.CardID
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}.Apply(NewContext(g, item))
}

const (
	ogreMarauderName      = "Ogre Marauder"
	ogreMarauderLetLabel  = "Let it become unblockable"
	ogreMarauderSacLabel  = "Sacrifice a creature"
	ogreMarauderQuestion  = ogreMarauderName + " — sacrifice a creature, or let it become unblockable this turn?"
	ogreMarauderBlockText = ogreMarauderName + " — can't be blocked this turn"
)

// OldB595OgreMarauderAsk is "it gains 'this creature can't be blocked' until end of turn
// unless defending player sacrifices a creature of their choice": the defending player
// decides; with no creature to sacrifice the creature is simply unblockable.
func OldB595OgreMarauderAsk(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	victim := DefendingPlayerOfAttack(g, item.Trigger.Event)
	if victim == uuid.Nil || g.PlayerByIDForEffect(victim) == nil || len(CreaturesControlledByPlayer(g, victim)) == 0 {
		return ogreMarauderUnblockable(ctx)
	}
	return PickOption{
		Player:   victim,
		Question: ogreMarauderQuestion,
		Options:  []game.ChoiceOption{{Label: ogreMarauderLetLabel}, {Label: ogreMarauderSacLabel}},
		ThenKey:  ogreMarauderAnswer,
	}.Apply(ctx)
}

func ogreMarauderUnblockable(ctx *Context) error {
	if !SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		return nil
	}
	return RestrictUntilEOT{Target: ctx.Source(), Restrictions: game.CantBeBlocked, Label: ogreMarauderBlockText}.Apply(ctx)
}

// ogreMarauderAnswer is the defending player's answer. An on-disk key: never renamed,
// never reused. A refusal to answer lets the creature through.
var ogreMarauderAnswer = OptionPickThen("option-pick/ogre-marauder-answer", func(ctx *Context, r game.OptionPicked) error {
	if r.Option != nil && r.Option.Label == ogreMarauderSacLabel {
		return SacrificeChoice{
			Player:     r.Chooser,
			Candidates: CreaturesControlledByPlayer(ctx.Game, r.Chooser),
			Question:   ogreMarauderName + " — sacrifice a creature",
		}.Apply(ctx)
	}
	return ogreMarauderUnblockable(ctx)
})

// OldB595ShizukoMana is "that player adds {G}{G}{G}, kept until end of turn" for the player
// whose upkeep it is.
func OldB595ShizukoMana(g *game.Game, item *game.StackItem) error {
	player := TriggeringActor(item)
	if player == uuid.Nil || g.PlayerByIDForEffect(player) == nil {
		return nil
	}
	return AddMana{Player: player, Produced: "{G}{G}{G}", Riders: []game.ManaSpendRider{KeepManaUntilEndOfTurn()}}.Apply(NewContext(g, item))
}
