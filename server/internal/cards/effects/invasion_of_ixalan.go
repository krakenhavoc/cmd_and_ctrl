package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Invasion of Ixalan // Belligerent Regisaur — Battle — Siege, defense 4,
// for {1}{G} (March of the Machine):
//
//	"(As a Siege enters, choose an opponent to protect it. You and others
//	  can attack it. When it's defeated, exile it, then cast it
//	  transformed.)
//	 When this Siege enters, look at the top five cards of your library.
//	 You may reveal a permanent card from among them and put it into your
//	 hand. Put the rest on the bottom of your library in a random order."
//
// Back face, Creature — Dinosaur 4/3:
//
//	"Trample
//	 Whenever you cast a spell, this creature gains indestructible until
//	 end of turn."
//
// The protector prompt, defense counters and the transformed cast are engine
// behaviour keyed on the battle type (battle.go). The Siege's own text is the
// Oath of Nissa look with "a permanent card". The back face registers under
// "<oracle_id>#1" (game.CatalogKey) and is Belligerent Regisaur's own
// trigger: a spell cast by its controller, any spell, grants indestructible
// until end of turn to the Regisaur itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     invasionOfIxalanOracleID,
		Name:         "Invasion of Ixalan",
		Completeness: CompletenessFull,
		Battle: &BattleSpec{
			Defense: 4,
			Subtype: BattleSubtypeSiege,
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Invasion of Ixalan — look at the top five cards of your library",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					player := ctx.Controller()
					return TakeFromLibraryToHand{
						Player:   player,
						Cards:    g.LookAtTopOfLibraryForEffect(player, 5),
						Match:    Permanent(),
						Max:      1,
						Optional: true,
						Label:    "Invasion of Ixalan — reveal a permanent card and put it into your hand",
						Key:      takeRevealRestBottomRandom,
					}.Apply(ctx)
				}),
			DefeatedTrigger("Invasion of Ixalan — defeated: exile it, then cast Belligerent Regisaur", SiegeDefeated()),
		},
	})

	Register(Spec{
		OracleID:        invasionOfIxalanOracleID + "#1",
		Name:            "Belligerent Regisaur",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(nil,
				"Belligerent Regisaur — gains indestructible until end of turn",
				Do(Instruction(func(ctx *Context) error {
					return GrantKeywordUntilEOT{
						Target:   ctx.Item.SourceCardID,
						Keywords: []string{"indestructible"},
						Label:    "Belligerent Regisaur — indestructible",
					}.Apply(ctx)
				}))),
		},
	})
}

// invasionOfIxalanOracleID is shared with the back face, which registers
// under it plus "#1".
const invasionOfIxalanOracleID = "1e5a5a7d-2c34-4e56-96e9-ebcf7745a03c"
