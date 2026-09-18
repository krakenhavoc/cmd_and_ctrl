package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thopter Fabricator — Artifact — Vehicle {2}{U}, 4/4 (EDHREC rank
// 4279):
//
//	"Flying
//	 Whenever you draw your second card each turn, create a 1/1
//	 colorless Thopter artifact creature token with flying.
//	 Crew 2"
//
// A three-mana 4/4 flier that also builds its own crew: every turn you
// draw a second card it makes a Thopter, and two Thopters crew it. In
// a blue deck with any draw engine at all that is one token a turn for
// free, and the Vehicle is a 4/4 that dodges sorcery-speed removal
// between attacks.
//
// "Your SECOND card each turn" is a once-a-turn trigger, not a
// from-the-second-onwards one, and the tally is what says so: the
// count is read after the draw (the harvester runs inside the same
// EmitEvent that records it), so the second draw reports exactly two
// and the third reports three. The draw-step draw is the first, which
// is why the card wants a cantrip and not a Howling Mine.
//
// CREW is an ordinary activated ability whose cost taps other
// creatures, so it uses the stack and can be responded to, and the
// creatures tap when the ability is ANNOUNCED. The Vehicle becomes an
// artifact creature until end of turn — a layer-4 type change, pinned
// to this permanent, swept at cleanup. Its printed 4/4 and its printed
// flying are already on the card, so the crew effect adds the CREATURE
// type and nothing else; setting P/T here would be a second source of
// truth that would silently win over a +1/+1 counter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8aab9181-c306-471a-8061-9a7a7654dab5",
		Name:            "Thopter Fabricator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Thopter Fabricator — create a 1/1 Thopter with flying",
				Do(CreateToken{Template: TokenCard("1/1 colorless Thopter artifact with flying"), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:  "Crew 2",
			Cost:   CrewCost(2),
			Effect: CrewEffect("Thopter Fabricator"),
		}},
	})
}
