package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spider Manifestation — 2/2 Creature — Spider Avatar for {1}{R/G}
// (EDHREC rank 3992):
//
//	"Reach
//	 {T}: Add {R} or {G}.
//	 Whenever you cast a spell with mana value 4 or greater, untap
//	 this creature."
//
// A two-mana mana creature that pays for itself again every time the
// deck casts something big. It is in the batch because it is the
// first card in the catalog to put a MANA ability and an untap
// TRIGGER on the same permanent, and the two have to compose: the
// mana ability taps it, the trigger untaps it, and the whole point is
// that the loop runs more than once a turn.
//
// The trigger fires on the CAST, so the mana the Spider made is
// already spent on the very spell that untaps it — tap for {R},
// spend it on a four-drop, and the Spider is untapped again before
// that spell resolves. It untaps whether or not the Spider was
// tapped, and whether or not it made the mana; nothing about the
// trigger reads either.
//
// "Mana value 4 or greater" is the SPELL's mana value on the stack,
// so an {X} spell counts the X chosen for it (CR 202.3e) and a spell
// cast for an alternative cost still counts its printed value.
//
// The hybrid pip {R/G} is a casting-cost question, not a card-file
// one. The ability the Spider itself offers is the ordinary printed
// pair.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8281f2b9-e81b-48da-812a-8713b2adf8ab",
		Name:            "Spider Manifestation",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		ManaAbilities:   []ManaAbility{dualManaAbility("R", "G")},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(ManaValueGE(4), "Spider Manifestation — untap it",
				func(g *game.Game, item *game.StackItem) error {
					return UntapTarget{Target: item.SourceCardID}.Apply(NewContext(g, item))
				}),
		},
	})
}
