package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Invasion of Arcavios // Invocation of the Founders — Battle — Siege,
// defense 7, for {3}{U}{U} (March of the Machine, S82 batch 23):
//
//	"(As a Siege enters, choose an opponent to protect it. You and others
//	  can attack it. When it's defeated, exile it, then cast it
//	  transformed.)
//	 When this Siege enters, search your library, graveyard, and/or outside
//	 the game for an instant or sorcery card you own, reveal it, and put it
//	 into your hand. If you search your library this way, shuffle."
//
// One search over three places, asked as the prompts the engine already has
// (Turtles Forever's two-place search with Finale of Devastation's
// graveyard):
//
//  1. Where to look: the library and/or graveyard together with outside the
//     game, the library and/or graveyard alone, or outside the game alone.
//     "And/or" is the searcher's choice, asked before anything is shown.
//  2. The library and/or graveyard half: an ordinary search (the engine
//     asks which of the two it looks through), revealing the card it finds
//     and shuffling the library if, and only if, it was searched (CR
//     701.23a). Searching a hidden zone for a stated quality may find
//     nothing (CR 701.23b).
//  3. The outside half (CR 701.23j: the player "may choose" a card): their
//     sideboard, or the card database for a deck with none (ADR 0191,
//     ADR 0198), revealed and put into the hand. When the searcher chose
//     both halves it is asked only if the first found nothing, since one
//     card is all the clause puts into the hand; the searcher has seen the
//     library and graveyard matches before declining them for it.
//
// Every prompt is keyed, so a table waiting on any of them is a restore
// point.
func init() {
	Register(Spec{
		OracleID:     invasionOfArcaviosOracleID,
		Name:         "Invasion of Arcavios",
		Purpose:      game.Purpose{Tutors: 1},
		Completeness: CompletenessFull,
		Battle: &BattleSpec{
			Defense: 7,
			Subtype: BattleSubtypeSiege,
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Invasion of Arcavios — search your library, graveyard, and/or outside the game for an instant or sorcery card",
				func(g *game.Game, item *game.StackItem) error {
					return PickOption{
						Question: "Invasion of Arcavios — search your library, graveyard, and/or outside the game for an instant or sorcery card you own",
						Options: []game.ChoiceOption{
							arcaviosBoth:        {Label: "Search your library and/or graveyard, and outside the game"},
							arcaviosLibraryOnly: {Label: "Search only your library and/or graveyard"},
							arcaviosOutsideOnly: {Label: "Search only outside the game"},
						},
						ThenKey: invasionOfArcaviosZones,
					}.Apply(NewContext(g, item))
				}),
			DefeatedTrigger("Invasion of Arcavios — defeated: exile it, then cast Invocation of the Founders", SiegeDefeated()),
		},
	})
}

// invasionOfArcaviosOracleID is shared with the back face's spec, which
// registers under this ID plus "#1".
const invasionOfArcaviosOracleID = "9794e31f-e73d-4aa0-a421-dc496080c987"

// The zone question's options, in the order offered. The first is the
// always-legal one the enumerator relies on.
const (
	arcaviosBoth = iota
	arcaviosLibraryOnly
	arcaviosOutsideOnly
)

// invasionOfArcaviosZones is the answer to "where do you search?". An
// on-disk key. Nobody answered (the controller left the game) does nothing.
var invasionOfArcaviosZones = OptionPickThen("option-pick/invasion-of-arcavios-zones",
	func(ctx *Context, r game.OptionPicked) error {
		switch r.Index {
		case arcaviosBoth, arcaviosLibraryOnly:
			alsoOutside := 0
			if r.Index == arcaviosBoth {
				alsoOutside = 1
			}
			return SearchLibrary{
				Player:        ctx.Controller(),
				Predicate:     func(c game.Card) bool { return c.IsInstant() || c.IsSorcery() },
				Dest:          game.ZoneHand,
				Limit:         1,
				Reveal:        true,
				Shuffle:       true,
				AlsoGraveyard: true,
				Reason:        "Invasion of Arcavios — an instant or sorcery card you own",
				ThenTail:      ctx.TailN(invasionOfArcaviosLibraryFound, alsoOutside),
			}.Apply(ctx)
		case arcaviosOutsideOnly:
			return invasionOfArcaviosAskOutside(ctx)
		}
		return nil
	})

// invasionOfArcaviosLibraryFound is the library and graveyard search's
// continuation, told the card it found (none when it found nothing);
// Tail.N is 1 when the controller also searches outside the game, which
// only matters when nothing was found. An on-disk key.
var invasionOfArcaviosLibraryFound = CardTail("invasion-of-arcavios/library-found", func(ctx *Context, o game.TailOutcome) error {
	if o.Tail.N != 1 || len(o.IDs) > 0 {
		return nil
	}
	return invasionOfArcaviosAskOutside(ctx)
})

// invasionOfArcaviosAskOutside asks for the outside half: one instant or
// sorcery card the controller owns outside the game, revealed and put into
// their hand. A closed pool or nothing that matches asks nothing.
func invasionOfArcaviosAskOutside(ctx *Context) error {
	return TakeFromOutside{
		Pick:     invasionOfArcaviosOutside,
		Question: "Invasion of Arcavios — reveal an instant or sorcery card you own from outside the game and put it into your hand, or none",
	}.Apply(ctx)
}

// invasionOfArcaviosOutside is the outside half's keyed pick. An on-disk
// key.
var invasionOfArcaviosOutside = registerOutsideTake("choose-cards/invasion-of-arcavios-outside",
	func(c game.Card) bool { return c.IsInstant() || c.IsSorcery() }, true)
