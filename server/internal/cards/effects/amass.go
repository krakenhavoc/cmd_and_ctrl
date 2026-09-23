package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// amass.go — the card-facing half of Amass [subtype] N (CR 701.47,
// #1236).
//
//	"Amass Orcs 1. (Put a +1/+1 counter on an Army you control. It's
//	 also an Orc. If you don't control an Army, create a 0/0 black Orc
//	 Army creature token first.)"
//
// The verb itself is `game.AmassForEffect` (game/amass.go), where the
// CR 614 window, the find-or-create branch, the multi-Army prompt, the
// counters and the layer-4 subtype live. This file is the two things
// the engine end deliberately does not know:
//
//   - the TOKEN, because `internal/game` owns no token templates; and
//   - the DECLARATION, so that a printed "Amass Orcs 1" is one line in
//     a card file and needs no bespoke code.
//
// # The Army token is BUILT, not tabled
//
// Every other token in the catalog is a row in tokens_table.go keyed
// by its printed text ("2/2 black Zombie"), or — when it has an
// ability of its own — a template in token_catalog.go (ADR 0083). The
// Army token is neither, and that is a decision rather than an
// oversight (ADR 0087 decision 2):
//
//   - It has no abilities, so ADR 0083's catalog is the wrong place:
//     a template there must declare one, and `checkTokenTemplate`
//     refuses one that does not.
//   - Its characteristics come from the KEYWORD, not from the card.
//     CR 701.47a says "create a 0/0 black [subtype] Army creature
//     token", where [subtype] is whatever the card printed. Four
//     species have a printed Army token card today (Zombie, Orc,
//     Goblin, Sliver) and four more are named by amass cards outside
//     tournament-legal sets (Ooze, Rat, Bird, Fan); the next set can
//     name a ninth. A table row per species is data the rule already
//     derives, and a card that amassed a species with no row would
//     panic in TokenCard at the moment it resolved.
//
// So `ArmyToken` derives the template from the subtype, exactly as the
// rule does. `TestArmyTokenMatchesThePrintedArmyToken` pins the
// derivation against the real printed Army tokens in the Scryfall dump
// so "derived" cannot quietly drift from "printed".

// ArmyToken is the token amass mints: "a 0/0 black [subtype] Army
// creature token" (CR 701.47a).
//
// `PrintedPTKnown` because a 0/0 with no counters really is a 0/0 and
// the CR 704.5f state-based action has to be allowed to kill it — an
// amass 0 with no Army out creates a token and puts nothing on it, and
// the token dies. It is the Phyrexian Germ's reason, one token over.
func ArmyToken(subtype string) game.Card {
	if subtype == "" {
		// CR 701.47d: the War of the Spark cards printed "amass N"
		// with no subtype and have been errata'd to "amass Zombies N".
		subtype = "Zombie"
	}
	return game.Card{
		Name:           subtype + " " + game.ArmySubtype,
		TypeLine:       "Token Creature — " + subtype + " " + game.ArmySubtype,
		Power:          0,
		Toughness:      0,
		Colors:         []string{"B"},
		PrintedPTKnown: true,
	}
}

// Amass is the keyword action "Amass [Subtype] N" (CR 701.47a): if you
// control no Army creature, create a 0/0 black [Subtype] Army creature
// token; choose an Army creature you control; put N +1/+1 counters on
// it; and if it isn't a [Subtype], it becomes one in addition to its
// other types.
//
// A card declares it and nothing else — `Do(Amass{Subtype: "Orc", N:
// 1})` is the whole of Orcish Bowmasters' second sentence.
type Amass struct {
	// Controller is the player amassing. Zero value means the
	// effect's controller, which is every printed amass but Azog,
	// Moria's Ruin's ("its controller amasses Goblins X").
	Controller uuid.UUID

	// Subtype is the creature type the keyword names, SINGULAR and
	// capitalised as the type line spells it — "Orc", not "Orcs".
	// Empty means Zombie (CR 701.47d).
	Subtype string

	// N is the number of +1/+1 counters. Zero is a real instruction
	// and not a no-op: the token is still created and still chosen
	// (CR 701.47a; the War of the Spark release note spells it out).
	N int

	// Then is CR 701.47c's "the Army you amassed" / "the amassed
	// Army": the rest of the sentence, handed the creature the amass
	// chose. It runs after the counters are placed, which is what
	// lets "the amassed Army deals damage equal to its power" read the
	// power the amass just gave it.
	//
	// The UUID is uuid.Nil when there was no Army to choose and none
	// could be made. CR 701.47b says the player amassed anyway, so the
	// clause still runs — with nothing to point at, which every caller
	// has to handle.
	//
	// nil for a card whose sentence ends at the amass, which is most
	// of them.
	Then func(ctx *Context, army uuid.UUID) error
}

func (a Amass) Apply(ctx *Context) error {
	controller := a.Controller
	if controller == uuid.Nil {
		controller = ctx.Controller()
	}
	item := ctx.Item
	then := a.Then
	var tail func(g *game.Game, army uuid.UUID) error
	if then != nil {
		tail = func(g *game.Game, army uuid.UUID) error {
			return then(NewContext(g, item), army)
		}
	}
	return ctx.Game.AmassForEffect(controller, ctx.Source(), ArmyToken(a.Subtype), a.Subtype, a.N, tail)
}

// Army passes for an Army creature — the read side of "an Army you
// control". `And(Creature(), Not(Army()))` is Widespread Brutality's
// "each non-Army creature"; `And(Army(), YouControl())` is the target
// clause Sauron, the Dark Lord and March from the Black Gate want when
// they are registered.
//
// It is `OfCreatureType(game.ArmySubtype)` under a name, because Army
// is a rules word in this engine rather than one tribe of many: the
// amass verb, the token it mints and four printed cards' own board
// queries all have to agree on it, and a misspelling in any of them
// reads as "no Armies" rather than failing.
//
// The BOARD-WALK half of the read side is `game.IsArmy` and
// `game.ArmiesControlledForEffect`, in the engine because the verb
// needs it there; a card-side clause wants this predicate instead.
func Army() CardPredicate { return OfCreatureType(game.ArmySubtype) }
