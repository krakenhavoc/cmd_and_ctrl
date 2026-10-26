package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Summon: Valefor — Enchantment Creature — Saga Drake {4}{U}, 5/4:
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after IV.)
//	 I — Sonic Wings — Each opponent chooses a creature with the
//	 greatest mana value among creatures they control. Return those
//	 creatures to their owners' hands.
//	 II, III, IV — Tap up to one target creature and put a stun counter
//	 on it.
//	 Flying"
//
// Chapter I asks each opponent in turn order, each choosing on their own
// board among their creatures tied for the greatest mana value (a
// choice, not a target), and then returns every chosen creature at once.
// An opponent with no creature chooses nothing.
//
// Chapters II–IV tap the picked creature and put one stun counter on it;
// with nothing picked they do nothing.
//
// No simplification.
func init() {
	const name = "Summon: Valefor"
	stun := func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		t, ok := ctx.ClauseTarget(0)
		if !ok {
			return nil
		}
		if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
		return AddCounter{Target: t.ID, Kind: game.CounterStun, N: 1}.Apply(ctx)
	}
	upToOne := func() *game.TargetSpec {
		return TargetCreature("up to one target creature").WithCount(0, 1)
	}
	label := SagaChapterLabel(name, 2, "tap up to one target creature and put a stun counter on it")
	Register(Spec{
		OracleID:        "67d29e56-9ca8-46bc-9ee0-20184e867ce4",
		Name:            name,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "Sonic Wings — each opponent returns a creature with the greatest mana value among creatures they control"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return valeforSonicWings(ctx, ctx.Opponents(), nil)
				}),
			ChapterTriggerTargeting(2, label, upToOne(), stun),
			ChapterTriggerTargeting(3, SagaChapterLabel(name, 3, "tap up to one target creature and put a stun counter on it"), upToOne(), stun),
			ChapterTriggerTargeting(4, SagaChapterLabel(name, 4, "tap up to one target creature and put a stun counter on it"), upToOne(), stun),
		},
	})
}

// valeforSonicWings asks the first of `opps` to choose, then the rest,
// and returns every creature chosen once the last has answered.
func valeforSonicWings(ctx *Context, opps, chosen []uuid.UUID) error {
	if len(opps) == 0 {
		for _, id := range chosen {
			if !onBattlefield(ctx.Game, id) {
				continue
			}
			if err := nothingIfGone(ctx.Game.BounceToHandForEffect(id)); err != nil {
				return err
			}
		}
		return nil
	}
	opp, rest := opps[0], append([]uuid.UUID(nil), opps[1:]...)
	so := append([]uuid.UUID(nil), chosen...)
	return ChoosePermanents{
		Player:   opp,
		Of:       []uuid.UUID{opp},
		Question: "Summon: Valefor — choose a creature with the greatest mana value among creatures you control to return to its owner's hand",
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			best := -1
			var ids []uuid.UUID
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller != of || !c.IsCreature() {
					continue
				}
				switch mv := c.ManaValue(); {
				case mv > best:
					best, ids = mv, []uuid.UUID{c.InstanceID}
				case mv == best:
					ids = append(ids, c.InstanceID)
				}
			}
			if len(ids) == 0 {
				return nil, 0, 0
			}
			return ids, 1, 1
		},
		Keyed: valeforChosen,
		Carry: packIDLists(rest, so),
	}.Apply(ctx)
}

// valeforChosen is one opponent's pick, then the next opponent's, keyed
// (#4290). The carry is the opponents still to ask, then the creatures
// chosen so far. Registered in init because the answer asks the next
// opponent, which names the key again.
var valeforChosen PermanentsThen

func init() {
	valeforChosen = RegisterPermanentsPick("prompt-run/summon-valefor-sonic-wings", PermanentsDef{
		ThenRun: func(ctx *Context, picked game.PromptedPicks, r game.RunDone) error {
			rest, so := unpackIDPair(r.Carry)
			return valeforSonicWings(ctx, rest, append(so, picked.Cards()...))
		},
	})
}
