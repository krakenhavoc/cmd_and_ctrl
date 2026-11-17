package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Invasion of Ikoria // Zilortha, Apex of Ikoria — Battle — Siege, defense 6,
// for {X}{G}{G} (March of the Machine, S82 batch 23):
//
//	"(As a Siege enters, choose an opponent to protect it. You and others
//	  can attack it. When it's defeated, exile it, then cast it
//	  transformed.)
//	 When this Siege enters, search your library and/or graveyard for a
//	 non-Human creature card with mana value X or less and put it onto the
//	 battlefield. If you search your library this way, shuffle."
//
// Finale of Devastation's search over both zones, keyed to the X the Siege
// was cast with (CR 107.3m), read off the permanent as the trigger is built
// and carried on the item, so it is exact even if the Siege has left by
// then. A Siege put onto the battlefield without being cast has X = 0 and
// can find only a mana value 0 creature. "Human" is a word of the Siege's
// text (CR 612.2).
//
// XMatters is left unset, as Invasion of New Phyrexia leaves it: at X = 0 the
// battle still enters with its defense counters and its transformed cast, so
// X = 0 is a real move and not a no-op the bot should stop being offered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     invasionOfIkoriaOracleID,
		Name:         "Invasion of Ikoria",
		Completeness: CompletenessFull,
		Battle: &BattleSpec{
			Defense: 6,
			Subtype: BattleSubtypeSiege,
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersWithCastX("Invasion of Ikoria — search your library and/or graveyard for a non-Human creature card with mana value X or less and put it onto the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					x := item.Params.Amount
					human := g.TextReaderForEffect(item).TextSubtype("Human")
					return SearchLibrary{
						Player:        item.Controller,
						Predicate:     func(c game.Card) bool { return c.IsCreature() && !c.HasSubtype(human) && c.ManaValue() <= x },
						Dest:          game.ZoneBattlefield,
						Limit:         1,
						Shuffle:       true,
						AlsoGraveyard: true,
						Reason:        "Invasion of Ikoria — a non-Human creature card with mana value X or less",
					}.Apply(NewContext(g, item))
				}),
			DefeatedTrigger("Invasion of Ikoria — defeated: exile it, then cast Zilortha, Apex of Ikoria", SiegeDefeated()),
		},
	})
}

// invasionOfIkoriaOracleID is shared with the back face's spec, which
// registers under this ID plus "#1".
const invasionOfIkoriaOracleID = "21d90735-5e04-4a8c-a546-ff26d4d5998d"
