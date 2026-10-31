package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Donald Blake, Guise of Thor — {1}{W} Legendary Creature — Human Doctor 1/3:
//
//	"Lifelink
//	 Power-up — {4}{W}{W}: Put two +1/+1 counters and a flying counter on
//	 Donald Blake. He becomes a God Warrior Hero. (He loses all other
//	 creature types. Activate each power-up ability only once. Reduce the
//	 cost by his mana cost if he entered this turn.)"
//
// The ability is one effect: two +1/+1 counters and a flying counter (real
// counters, CR 122.1b, so the flying is the layered view's) and then a
// no-duration creature-type SET (game.SetCreatureTypesMod, CR 205.1a: "loses
// all other creature types"), pinned to this object so a blink starts fresh
// (CR 400.7). If he has left the battlefield by resolution nothing happens.
//
// No simplification. Marvel Super Heroes (#4455).
func init() {
	ab := PowerUp(
		"{4}{W}{W}: Put two +1/+1 counters and a flying counter on Donald Blake. He becomes a God Warrior Hero.",
		ManaCost("{4}{W}{W}"), func(g *game.Game, item *game.StackItem) error {
			if !sourceIsStillThisPermanent(g, item) {
				return nil
			}
			ctx := NewContext(g, item)
			if err := (AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
				return err
			}
			if err := (AddCounter{Target: item.SourceCardID, Kind: game.CounterFlying, N: 1}).Apply(ctx); err != nil {
				return err
			}
			return ScopedEffectFor{
				Target:   item.SourceCardID,
				Mods:     []game.Mod{game.SetCreatureTypesMod(ctx.TextSubtypes("God", "Warrior", "Hero")...)},
				Duration: g.PinnedTo(game.IndefiniteDuration(), item.SourceCardID),
				Label:    "Donald Blake — a God Warrior Hero",
			}.Apply(ctx)
		})
	// Counters on itself (a pump) and a type change.
	ab.Purpose = game.Purpose{Answers: game.AnswerPump}
	Register(Spec{
		OracleID:        "f966ae40-4717-4f41-9437-0e38b01a8be9",
		Name:            "Donald Blake, Guise of Thor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Activated:       []ActivatedAbility{ab},
	})
}
