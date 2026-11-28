package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OldB637SpellTargetsACreature matches a spell on the stack with at least one
// creature on the battlefield among its targets (Confound's "spell that
// targets a creature"). A spell that targets only players, or a creature that
// has already left, does not qualify.
func OldB637SpellTargetsACreature() CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		item := g.StackItemForEffect(c.InstanceID)
		if item == nil {
			return false
		}
		for _, t := range item.Targets {
			if t.Kind != game.TargetCard {
				continue
			}
			z := g.FindCardZoneForEffect(t.ID)
			if z == nil || z.Kind != game.ZoneBattlefield {
				continue
			}
			if tc, ok := g.LookupCardForEffect(t.ID); ok && tc.IsCreature() {
				return true
			}
		}
		return false
	}
}

// OldB637DestroyTargetThenDraw is "Destroy target <permanent>. Draw a card."
// (Aura Blast, Implode): the destroy, then the draw (CR 608.2c). A target that
// became illegal counters the whole spell, so nothing is drawn (CR 608.2b).
func OldB637DestroyTargetThenDraw() func(*game.StackItem, *Context) error {
	return ResolveSteps(
		OnTheTargetCard(func(ctx *Context, id uuid.UUID) error { return DestroyTarget{Target: id}.Apply(ctx) }),
		DrawCards{N: 1},
	)
}

// OldB637DamageToYourCreaturesMinusOne is Lashknife Barrier's "If a source
// would deal damage to a creature you control, it deals that much damage minus
// 1 to that creature instead": a replacement (CR 614.1a), not a prevention, so
// "damage can't be prevented" does not stop it. Damage to players and to
// other players' creatures is untouched.
func OldB637DamageToYourCreaturesMinusOne(name string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventDealDamage},
		// Not a prevention effect (CR 615.1): it rewrites the amount.
		Prevention: false,
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if src == nil || ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 {
				return false
			}
			t, ok := g.LookupCardForEffect(ev.DamageTarget)
			return ok && t.IsCreature() && t.Controller == src.Controller
		},
		// The same arithmetic as "prevent 1 of that damage": take 1 off, cancel at 0.
		Replace:    S82OldB317PreventOneOfThatDamage,
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID { return src.Controller },
		Label:      name + " — damage to a creature you control is reduced by 1",
	}
}

// OldB637UrzasGuilt is "Each player draws two cards, then discards three
// cards, then loses 4 life." Every draw lands first, every player then chooses
// their own three discards as one run, and the life loss waits for the last
// discard (CR 608.2c).
func OldB637UrzasGuilt(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	players := TablePlayers(ctx)
	for _, id := range players {
		if err := (DrawCards{Player: id, N: 2}).Apply(ctx); err != nil {
			return err
		}
	}
	return g.PlayersDiscardThenKeyedForEffect(players,
		game.DiscardPrompt{Source: item.SourceCardID, N: 3},
		oldB637UrzasGuiltLoseLife, players)
}

// oldB637UrzasGuiltLoseLife is the run's rest: every player (the carry) loses
// 4 life. An on-disk key: never renamed, never reused.
var oldB637UrzasGuiltLoseLife = game.RegisterPromptRunThen("prompt-run/old-b637-urzas-guilt-lose-life", func(g *game.Game, r game.RunDone) error {
	for _, id := range r.Carry {
		if err := g.ChangePlayerLifeForEffect(r.Source, id, -4); err != nil {
			return err
		}
	}
	return nil
})
