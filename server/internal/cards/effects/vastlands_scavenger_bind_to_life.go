package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vastlands Scavenger // Bind to Life — a preparation card (CR 722, ADR
// 0090), Secrets of Strixhaven (#4456):
//
//	Vastlands Scavenger {1}{G}{G}, Creature — Bear Druid 4/4:
//	"Deathtouch
//	 This creature enters prepared. (While it's prepared, you may cast a copy
//	 of its spell. Doing so unprepares it.)"
//	Bind to Life {4}{G}, Instant:
//	"Mill seven cards. Then put a creature card from among them onto the
//	 battlefield."
//
// Deathtouch is a printed keyword and "enters prepared" is SelfEntersPrepared.
// Bind to Life mills seven (the caster's library), then, with the mill done
// (it can pause on a replacement prompt, so the pick is the mill's keyed
// tail), the caster puts a creature card from among the cards that mill put
// into the graveyard onto the battlefield. The pick is not optional: with a
// creature card among them one comes back (a lone one outright), and with none
// nothing does. A card the mill sent elsewhere (a replacement) was not milled
// into the graveyard and is not offered.
//
// No simplification.
func init() {
	const id = "3a97bc8d-bd81-468e-9277-eba4fac11ee0"
	Register(Spec{
		OracleID:        id,
		Name:            "Vastlands Scavenger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Replacements:    []game.ReplacementEffect{SelfEntersPrepared()},
	})
	Register(Spec{
		OracleID:     id + "#1",
		Name:         "Bind to Life",
		Completeness: CompletenessFull,
		OnResolve:    socB38BindToLife,
	})
}
