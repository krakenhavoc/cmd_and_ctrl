package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// m14_5563_helpers.go — shared shapes for the Magic 2014 batch (#5563).

// m14StaffOfTheMagus is the Staff cycle's one ability: "Whenever you cast a
// <color> spell or a <land type> you control enters, you gain 1 life." It is
// one printed ability with two trigger conditions (OnAny), so a spell cast and
// a land entering each trigger it once. The spell's color and the land type
// are words of the Staff's text (CR 612.2), read through it.
func m14StaffOfTheMagus(name, color, landType string) game.TriggeredAbility {
	cast := YouCast(OfColor(color))
	land := cmdr2S14SubtypeLandEntersUnderYourControl(landType)
	return OnAny([]game.EventKind{game.EventCast, game.EventETB},
		func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
			if ev.Kind == game.EventETB {
				return land(ev, source, lki, g)
			}
			return cast(ev, source, lki, g)
		}, name+" — you gain 1 life", Do(GainLife{Amount: 1}))
}

// m14ThisOrAnotherHumanYouControlDied is Xathrid Necromancer's "this creature
// or another Human creature you control dies": the source's own death counts
// whatever its types, and another creature counts when it was a Human under
// the source's controller as it last existed (leftAsSubtype,
// leftUnderControlOf, CR 603.10a).
func m14ThisOrAnotherHumanYouControlDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := diedCreature(ev, g)
	if !ok {
		return false
	}
	if ev.CardID == source.InstanceID {
		return true
	}
	return leftUnderControlOf(ev, dead) == source.Controller && leftAsSubtype(ev, dead, source.TextSubtype("Human"))
}

// m14MaraudingMaulhornMustAttack is "This creature attacks each combat if able
// unless you control a creature named Advocate of the Beast": Bloodrock
// Cyclops's requirement, switched off while the exemption holds. The name is
// read in the layer pass (youControlACreatureNamed).
func m14MaraudingMaulhornMustAttack() game.StaticAbility {
	exempt := youControlACreatureNamed("Advocate of the Beast")
	s := AttacksEachCombatWhere(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return selfOnly(target, g, source) && !exempt(g, source)
	})
	s.Label = "Attacks each combat if able unless you control a creature named Advocate of the Beast."
	return s
}

// ringOfThreeWishesTutor is Ring of Three Wishes' "Search your library for a
// card, put that card into your hand, then shuffle."
func ringOfThreeWishesTutor(g *game.Game, item *game.StackItem) error {
	return SearchLibrary{
		Player:  item.Controller,
		Dest:    game.ZoneHand,
		Limit:   1,
		Shuffle: true,
		Source:  item.SourceCardID,
		Reason:  "Ring of Three Wishes — a card, to your hand",
	}.Apply(NewContext(g, item))
}
