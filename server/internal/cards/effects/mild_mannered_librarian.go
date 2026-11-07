package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mild-Mannered Librarian — Creature — Human {G}, 1/1 (Foundations,
// #4465):
//
//	"{3}{G}: This creature becomes a Werewolf. Put two +1/+1 counters on
//	 it and you draw a card. Activate only once."
//
// "Activate only once" is the OnlyOnce bit (#5221, CR 602.5b): once per
// object, so a Librarian that leaves and comes back may activate it again
// (CR 400.7).
//
// "Becomes a Werewolf" with no "in addition to its other types" sets the
// creature type, so the Werewolf is no longer a Human (CR 205.1a). The
// change prints no duration, so it lasts until the end of the game
// (CR 611.2a), pinned to this object. A Librarian that left in response
// is not changed and gets no counters, and its controller still draws.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fb7070ab-c234-4d54-8cdc-caf106bf24a6",
		Name:         "Mild-Mannered Librarian",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:    "{3}{G}: This creature becomes a Werewolf. Put two +1/+1 counters on it and you draw a card. Activate only once.",
			Purpose:  game.Purpose{Answers: game.AnswerPump, Draws: 1},
			Cost:     ManaCost("{3}{G}"),
			OnlyOnce: true,
			Effect:   mildManneredLibrarianBecomesAWerewolf,
		}},
	})
}

// mildManneredLibrarianBecomesAWerewolf is the ability's effect, in
// printed order.
func mildManneredLibrarianBecomesAWerewolf(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if self := item.SourceCardID; sourceIsStillThisPermanent(g, item) {
		if err := (ScopedEffectFor{
			Target:   self,
			Mods:     []game.Mod{game.SetCreatureTypesMod(ctx.TextSubtypes("Werewolf")...)},
			Duration: g.PinnedTo(game.IndefiniteDuration(), self),
			Label:    "Mild-Mannered Librarian — a Werewolf",
		}).Apply(ctx); err != nil {
			return err
		}
		if err := (AddCounter{Target: self, Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
			return err
		}
	}
	return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
}
