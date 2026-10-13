package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Angel of Serenity — Creature — Angel {4}{W}{W}{W}, 5/6 (EDHREC rank
// 3758):
//
//	"Flying
//	 When this creature enters, you may exile up to three other target
//	 creatures from the battlefield and/or creature cards from
//	 graveyards.
//	 When this creature leaves the battlefield, return the exiled
//	 cards to their owners' hands."
//
// Seven mana for a 5/6 flier that takes three creatures with it, and
// gives them back — to their owners' HANDS, not the battlefield — if
// she ever leaves. Flying rides PrintedKeywords. The entry trigger is
// optional and targeted at up to three creatures; the leave trigger
// fires on ANY exit (dying, exile, bounce, a flicker) and returns
// every card her entry trigger exiled that is still in exile. The entry
// exile is LINKED (ExileTarget.Linked, ADR 0155 §8), so each card lands
// with Card.ExiledWith naming this Angel object, and the leave trigger
// reads the link through its own source object (her last-known record).
// A card something else has since moved on has lost the link and is not
// pulled out of wherever it went.
//
// The target clause spans two zones (#3172): "creatures from the
// battlefield and/or creature cards from graveyards", so its Zones are the
// battlefield and the graveyards, and the creature predicate reads a
// graveyard card's printed type line. "Another" is the clause's own
// ExcludeSource, so a token copy of her is a legal target, as printed.
func init() {
	Register(Spec{
		OracleID:        "f4ce6078-8c7b-4f68-b324-13130c63a983",
		Name:            "Angel of Serenity",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			{
				Watches:        []game.EventKind{game.EventETB},
				AppliesTo:      b06SelfETB,
				OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Angel of Serenity — exile up to three other target creatures?"},
				Targets:        angelOfSerenityTargets(),
				Key:            b36AngelOfSerenityExileLabel,
				Effect:         b36ExileChosenCreatures,
			},
			On(game.EventLTB, Self, "Angel of Serenity — return the exiled cards to their owners' hands", b36ReturnCardsExiledWithToOwnersHands),
		},
	})
}

// angelOfSerenityTargets is "up to three other target creatures from the
// battlefield and/or creature cards from graveyards".
func angelOfSerenityTargets() *game.TargetSpec {
	spec := Another(TargetCreature("up to three other target creatures and/or creature cards in graveyards")).WithCount(0, 3)
	spec.Zones = []game.ZoneKind{game.ZoneBattlefield, game.ZoneGraveyard}
	return spec
}
